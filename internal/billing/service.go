package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stripe/stripe-go/v82"
	checkoutsession "github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/webhook"
)

var planSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Service struct {
	queries *db.Queries
	config  Config
}

func NewService(queries *db.Queries, config Config) *Service {
	if config.Enabled() {
		stripe.Key = config.StripeSecretKey
	}
	return &Service{queries: queries, config: config}
}

func (s *Service) ListPlans(ctx context.Context) ([]PlanResponse, error) {
	rows, err := s.queries.ListPlans(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]PlanResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toPlanResponse(row))
	}
	return items, nil
}

func (s *Service) GetPlatformStatus(ctx context.Context) (PlatformStatusResponse, error) {
	mode, err := s.getBoolSetting(ctx, "maintenance_mode")
	if err != nil {
		return PlatformStatusResponse{}, err
	}
	msg, _ := s.queries.GetPlatformSetting(ctx, "maintenance_message")
	resp := PlatformStatusResponse{MaintenanceMode: mode}
	if msg != "" {
		resp.MaintenanceMessage = msg
	}
	return resp, nil
}

func (s *Service) GetPlatformSettings(ctx context.Context) (PlatformSettingsResponse, error) {
	mode, err := s.getBoolSetting(ctx, "maintenance_mode")
	if err != nil {
		return PlatformSettingsResponse{}, err
	}
	msg, err := s.queries.GetPlatformSetting(ctx, "maintenance_message")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			msg = ""
		} else {
			return PlatformSettingsResponse{}, err
		}
	}
	return PlatformSettingsResponse{
		MaintenanceMode:    mode,
		MaintenanceMessage: msg,
	}, nil
}

func (s *Service) UpdatePlatformSettings(ctx context.Context, req UpdatePlatformSettingsRequest) (PlatformSettingsResponse, error) {
	if req.MaintenanceMode == nil && req.MaintenanceMessage == nil {
		return PlatformSettingsResponse{}, ErrInvalidRequest
	}
	if req.MaintenanceMode != nil {
		if err := s.queries.UpsertPlatformSetting(ctx, db.UpsertPlatformSettingParams{
			Key:   "maintenance_mode",
			Value: strconv.FormatBool(*req.MaintenanceMode),
		}); err != nil {
			return PlatformSettingsResponse{}, err
		}
	}
	if req.MaintenanceMessage != nil {
		if err := s.queries.UpsertPlatformSetting(ctx, db.UpsertPlatformSettingParams{
			Key:   "maintenance_message",
			Value: strings.TrimSpace(*req.MaintenanceMessage),
		}); err != nil {
			return PlatformSettingsResponse{}, err
		}
	}
	return s.GetPlatformSettings(ctx)
}

func (s *Service) IsMaintenanceMode(ctx context.Context) (bool, string, error) {
	status, err := s.GetPlatformStatus(ctx)
	if err != nil {
		return false, "", err
	}
	return status.MaintenanceMode, status.MaintenanceMessage, nil
}

func (s *Service) OnAppCreated(ctx context.Context, appID int64, planSlug string) error {
	planSlug = strings.TrimSpace(planSlug)
	var plan db.Plan
	var err error

	if planSlug != "" {
		plan, err = s.queries.GetActivePlan(ctx, planSlug)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvalidPlan
			}
			return err
		}
	} else {
		plan, err = s.queries.GetDefaultPlan(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvalidPlan
			}
			return err
		}
	}

	_, err = s.queries.CreateAppSubscription(ctx, db.CreateAppSubscriptionParams{
		AppID:    appID,
		PlanSlug: plan.Slug,
		Status:   StatusIncomplete,
	})
	return err
}

func (s *Service) GetAppBilling(ctx context.Context, appID int64) (AppBillingResponse, error) {
	sub, err := s.queries.GetAppSubscription(ctx, appID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AppBillingResponse{}, ErrNotFound
		}
		return AppBillingResponse{}, err
	}

	usageCount, err := s.queries.GetAppMessageUsage(ctx, appID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return AppBillingResponse{}, err
	}
	messages := int64(0)
	if err == nil {
		messages = usageCount
	}

	return AppBillingResponse{
		AppID:        appID,
		Subscription: toSubscriptionResponse(sub),
		Usage: UsageResponse{
			MessagesThisMonth: messages,
			MessagesLimit:     sub.MessagesLimit,
		},
	}, nil
}

func (s *Service) CreateCheckoutSession(ctx context.Context, userID int64, userEmail string, appID int64, planSlug string) (CheckoutResponse, error) {
	if !s.config.Enabled() {
		return CheckoutResponse{}, ErrStripeNotConfigured
	}

	app, err := s.queries.GetAppByID(ctx, appID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CheckoutResponse{}, ErrNotFound
		}
		return CheckoutResponse{}, err
	}
	if app.OwnerID != userID {
		return CheckoutResponse{}, ErrForbidden
	}

	planSlug = strings.TrimSpace(planSlug)
	if planSlug == "" {
		return CheckoutResponse{}, ErrInvalidPlan
	}
	plan, err := s.queries.GetActivePlan(ctx, planSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CheckoutResponse{}, ErrInvalidPlan
		}
		return CheckoutResponse{}, err
	}

	if !plan.StripePriceID.Valid || strings.TrimSpace(plan.StripePriceID.String) == "" {
		return CheckoutResponse{}, ErrStripeNotConfigured
	}
	priceID := plan.StripePriceID.String

	customerID, err := s.ensureStripeCustomer(ctx, userID, userEmail)
	if err != nil {
		return CheckoutResponse{}, err
	}

	successURL := fmt.Sprintf(
		"%s/apps/%d/billing?session_id={CHECKOUT_SESSION_ID}",
		s.config.CheckoutBaseURL,
		appID,
	)
	cancelURL := fmt.Sprintf("%s/apps/%d/billing", s.config.CheckoutBaseURL, appID)

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{Price: stripe.String(priceID), Quantity: stripe.Int64(1)},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"app_id":    strconv.FormatInt(appID, 10),
			"plan_slug": planSlug,
			"user_id":   strconv.FormatInt(userID, 10),
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"app_id":    strconv.FormatInt(appID, 10),
				"plan_slug": planSlug,
			},
		},
	}

	sess, err := checkoutsession.New(params)
	if err != nil {
		return CheckoutResponse{}, err
	}

	_, _ = s.queries.UpdateAppSubscription(ctx, db.UpdateAppSubscriptionParams{
		PlanSlug:         pgtype.Text{String: planSlug, Valid: true},
		StripeCustomerID: pgtype.Text{String: customerID, Valid: true},
		AppID:            appID,
	})

	return CheckoutResponse{URL: sess.URL}, nil
}

func (s *Service) UserEmail(ctx context.Context, userID int64) (string, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return user.Email, nil
}

func (s *Service) AssertAppOwner(ctx context.Context, appID, userID int64) error {
	app, err := s.queries.GetAppByID(ctx, appID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if app.OwnerID != userID {
		return ErrForbidden
	}
	return nil
}

func (s *Service) CheckMessageQuota(ctx context.Context, appID int64) error {
	sub, err := s.queries.GetAppSubscription(ctx, appID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if !statusAllowsUsage(sub.Status) {
		return ErrSubscriptionInactive
	}
	count, err := s.queries.GetAppMessageUsage(ctx, appID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if int64(sub.MessagesLimit) > 0 && count >= int64(sub.MessagesLimit) {
		return ErrMessageLimit
	}
	return nil
}

func (s *Service) CheckMessageQuotaByWorkspace(ctx context.Context, workspaceID int64) error {
	app, err := s.queries.GetAppByWorkspaceID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if workspaceID == app.SandboxWorkspaceID {
		return nil
	}
	return s.CheckMessageQuota(ctx, app.ID)
}

func (s *Service) RecordMessageByAppID(ctx context.Context, appID int64) error {
	_, err := s.queries.IncrementAppMessageUsage(ctx, appID)
	return err
}

func (s *Service) RecordMessageByWorkspaceID(ctx context.Context, workspaceID int64) error {
	app, err := s.queries.GetAppByWorkspaceID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if workspaceID == app.SandboxWorkspaceID {
		return nil
	}
	return s.RecordMessageByAppID(ctx, app.ID)
}

func (s *Service) CheckCanCreateWebhook(ctx context.Context, appID int64) error {
	sub, err := s.queries.GetAppSubscription(ctx, appID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrWebhooksNotAllowed
		}
		return err
	}
	if !statusAllowsUsage(sub.Status) {
		return ErrSubscriptionInactive
	}
	if !sub.WebhooksEnabled {
		return ErrWebhooksNotAllowed
	}
	return nil
}

func (s *Service) CheckCanStartVideoCall(ctx context.Context, workspaceID int64) error {
	app, err := s.queries.GetAppByWorkspaceID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	sub, err := s.queries.GetAppSubscription(ctx, app.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrVideoNotAllowed
		}
		return err
	}
	if !statusAllowsUsage(sub.Status) {
		return ErrSubscriptionInactive
	}
	if !sub.VideoCallsEnabled {
		return ErrVideoNotAllowed
	}
	return nil
}

func (s *Service) AdminGetOverview(ctx context.Context) (AdminBillingOverview, error) {
	count, err := s.queries.AdminCountActiveSubscriptions(ctx)
	if err != nil {
		return AdminBillingOverview{}, err
	}
	return AdminBillingOverview{ActiveSubscriptions: count}, nil
}

func (s *Service) AdminListSubscriptions(ctx context.Context, limit, offset int32) ([]AdminSubscriptionListItem, error) {
	rows, err := s.queries.AdminListAppSubscriptions(ctx, db.AdminListAppSubscriptionsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]AdminSubscriptionListItem, 0, len(rows))
	for _, row := range rows {
		item := AdminSubscriptionListItem{
			AppID:             row.AppID,
			AppName:           row.AppName,
			AppSlug:           row.AppSlug,
			PlanSlug:          row.PlanSlug,
			PlanName:          row.PlanName,
			Status:            row.Status,
			MessagesThisMonth: row.MessagesThisMonth,
		}
		if row.CurrentPeriodEnd.Valid {
			item.CurrentPeriodEnd = formatTime(row.CurrentPeriodEnd)
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) AdminUpdateSubscription(ctx context.Context, appID int64, req AdminUpdateSubscriptionRequest) (AppBillingResponse, error) {
	if req.PlanSlug == nil && req.Status == nil {
		return AppBillingResponse{}, ErrInvalidRequest
	}
	if _, err := s.queries.GetAppSubscription(ctx, appID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AppBillingResponse{}, ErrNotFound
		}
		return AppBillingResponse{}, err
	}
	params := db.UpdateAppSubscriptionParams{AppID: appID}
	if req.PlanSlug != nil {
		slug := strings.TrimSpace(*req.PlanSlug)
		if _, err := s.queries.GetPlan(ctx, slug); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return AppBillingResponse{}, ErrInvalidPlan
			}
			return AppBillingResponse{}, err
		}
		params.PlanSlug = pgtype.Text{String: slug, Valid: true}
	}
	if req.Status != nil {
		params.Status = pgtype.Text{String: strings.TrimSpace(*req.Status), Valid: true}
	}
	if _, err := s.queries.UpdateAppSubscription(ctx, params); err != nil {
		return AppBillingResponse{}, err
	}
	return s.GetAppBilling(ctx, appID)
}

func (s *Service) HandleStripeWebhook(ctx context.Context, payload []byte, signature string) error {
	if !s.config.Enabled() || s.config.StripeWebhookSecret == "" {
		return ErrStripeNotConfigured
	}
	event, err := webhook.ConstructEvent(payload, signature, s.config.StripeWebhookSecret)
	if err != nil {
		return err
	}

	switch event.Type {
	case "checkout.session.completed":
		return s.handleCheckoutCompleted(ctx, event)
	case "customer.subscription.updated", "customer.subscription.deleted":
		return s.handleSubscriptionChanged(ctx, event)
	default:
		return nil
	}
}

func (s *Service) handleCheckoutCompleted(ctx context.Context, event stripe.Event) error {
	var raw struct {
		Metadata     map[string]string `json:"metadata"`
		Customer     string            `json:"customer"`
		Subscription string            `json:"subscription"`
	}
	if err := json.Unmarshal(event.Data.Raw, &raw); err != nil {
		return err
	}
	appID, err := parseMetadataInt64(raw.Metadata, "app_id")
	if err != nil {
		return err
	}
	planSlug := raw.Metadata["plan_slug"]

	params := db.UpdateAppSubscriptionParams{
		AppID:  appID,
		Status: pgtype.Text{String: StatusActive, Valid: true},
	}
	if planSlug != "" {
		params.PlanSlug = pgtype.Text{String: planSlug, Valid: true}
	}
	if raw.Customer != "" {
		params.StripeCustomerID = pgtype.Text{String: raw.Customer, Valid: true}
	}
	if raw.Subscription != "" {
		params.StripeSubscriptionID = pgtype.Text{String: raw.Subscription, Valid: true}
	}
	_, err = s.queries.UpdateAppSubscription(ctx, params)
	return err
}

func (s *Service) handleSubscriptionChanged(ctx context.Context, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return err
	}
	if sub.ID == "" {
		return nil
	}

	stripeSubID := pgtype.Text{String: sub.ID, Valid: true}
	existing, err := s.queries.GetAppSubscriptionByStripeSubscriptionID(ctx, stripeSubID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			appID, parseErr := parseMetadataInt64(sub.Metadata, "app_id")
			if parseErr != nil {
				return nil
			}
			planSlug := sub.Metadata["plan_slug"]
			if planSlug == "" {
				defaultPlan, defaultErr := s.queries.GetDefaultPlan(ctx)
				if defaultErr != nil {
					return defaultErr
				}
				planSlug = defaultPlan.Slug
			}
			_, createErr := s.queries.CreateAppSubscription(ctx, db.CreateAppSubscriptionParams{
				AppID:    appID,
				PlanSlug: planSlug,
				Status:   mapStripeStatus(string(sub.Status)),
			})
			if createErr != nil {
				return createErr
			}
			existing, err = s.queries.GetAppSubscriptionByStripeSubscriptionID(ctx, stripeSubID)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	status := mapStripeStatus(string(sub.Status))
	if event.Type == "customer.subscription.deleted" {
		status = StatusCanceled
	}

	params := db.UpdateAppSubscriptionParams{
		AppID:  existing.AppID,
		Status: pgtype.Text{String: status, Valid: true},
	}
	if planSlug := sub.Metadata["plan_slug"]; planSlug != "" {
		params.PlanSlug = pgtype.Text{String: planSlug, Valid: true}
	}
	_, err = s.queries.UpdateAppSubscription(ctx, params)
	return err
}

func (s *Service) ensureStripeCustomer(ctx context.Context, userID int64, email string) (string, error) {
	stored, err := s.queries.GetUserStripeCustomerID(ctx, userID)
	if err == nil && stored.Valid && strings.TrimSpace(stored.String) != "" {
		return stored.String, nil
	}

	cust, err := customer.New(&stripe.CustomerParams{
		Email: stripe.String(email),
		Metadata: map[string]string{
			"user_id": strconv.FormatInt(userID, 10),
		},
	})
	if err != nil {
		return "", err
	}
	_ = s.queries.SetUserStripeCustomerID(ctx, db.SetUserStripeCustomerIDParams{
		ID:               userID,
		StripeCustomerID: pgtype.Text{String: cust.ID, Valid: true},
	})
	return cust.ID, nil
}

func (s *Service) getBoolSetting(ctx context.Context, key string) (bool, error) {
	val, err := s.queries.GetPlatformSetting(ctx, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return false, nil
	}
	return parsed, nil
}

func mapStripeStatus(status string) string {
	switch status {
	case string(stripe.SubscriptionStatusActive):
		return StatusActive
	case string(stripe.SubscriptionStatusTrialing):
		return StatusTrialing
	case string(stripe.SubscriptionStatusPastDue):
		return StatusPastDue
	case string(stripe.SubscriptionStatusCanceled), string(stripe.SubscriptionStatusUnpaid):
		return StatusCanceled
	case string(stripe.SubscriptionStatusIncomplete), string(stripe.SubscriptionStatusIncompleteExpired):
		return StatusIncomplete
	default:
		return StatusIncomplete
	}
}

func parseMetadataInt64(metadata map[string]string, key string) (int64, error) {
	raw := metadata[key]
	if raw == "" {
		return 0, fmt.Errorf("missing metadata %s", key)
	}
	return strconv.ParseInt(raw, 10, 64)
}

func toPlanResponse(p db.Plan) PlanResponse {
	return PlanResponse{
		Slug:              p.Slug,
		Name:              p.Name,
		Description:       p.Description,
		PriceCents:        p.PriceCents,
		PriceDisplay:      fmt.Sprintf("$%.0f", float64(p.PriceCents)/100),
		MessagesLimit:     p.MessagesLimit,
		ConnectionsLimit:  p.ConnectionsLimit,
		RetentionDays:     p.RetentionDays,
		WebhooksEnabled:   p.WebhooksEnabled,
		VideoCallsEnabled: p.VideoCallsEnabled,
		ModerationEnabled: p.ModerationEnabled,
		IsHighlighted:     p.IsHighlighted,
		SortOrder:         p.SortOrder,
	}
}

func toAdminPlanResponse(p db.Plan, subscriptionCount int64) AdminPlanResponse {
	resp := AdminPlanResponse{
		PlanResponse:      toPlanResponse(p),
		IsActive:          p.IsActive,
		SubscriptionCount: subscriptionCount,
	}
	if p.StripePriceID.Valid {
		resp.StripePriceID = p.StripePriceID.String
	}
	return resp
}

func normalizePlanSlug(raw string) (string, error) {
	slug := strings.ToLower(strings.TrimSpace(raw))
	if slug == "" || !planSlugPattern.MatchString(slug) {
		return "", ErrInvalidSlug
	}
	return slug, nil
}

func (s *Service) AdminListPlans(ctx context.Context) ([]AdminPlanResponse, error) {
	rows, err := s.queries.ListPlansAdmin(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]AdminPlanResponse, 0, len(rows))
	for _, row := range rows {
		count, err := s.queries.CountSubscriptionsForPlan(ctx, row.Slug)
		if err != nil {
			return nil, err
		}
		items = append(items, toAdminPlanResponse(row, count))
	}
	return items, nil
}

func (s *Service) AdminCreatePlan(ctx context.Context, req CreatePlanRequest) (AdminPlanResponse, error) {
	slug, err := normalizePlanSlug(req.Slug)
	if err != nil {
		return AdminPlanResponse{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return AdminPlanResponse{}, ErrInvalidRequest
	}
	if req.PriceCents < 0 || req.MessagesLimit < 0 || req.ConnectionsLimit < 0 || req.RetentionDays < 0 {
		return AdminPlanResponse{}, ErrInvalidRequest
	}

	if _, err := s.queries.GetPlan(ctx, slug); err == nil {
		return AdminPlanResponse{}, ErrSlugTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return AdminPlanResponse{}, err
	}

	params := db.CreatePlanParams{
		Slug:              slug,
		Name:              name,
		PriceCents:        req.PriceCents,
		MessagesLimit:     req.MessagesLimit,
		ConnectionsLimit:  req.ConnectionsLimit,
		RetentionDays:     req.RetentionDays,
		WebhooksEnabled:   req.WebhooksEnabled,
		VideoCallsEnabled: req.VideoCallsEnabled,
		ModerationEnabled: req.ModerationEnabled,
		IsActive:          req.IsActive,
		IsHighlighted:     req.IsHighlighted,
		SortOrder:         req.SortOrder,
		Description:       strings.TrimSpace(req.Description),
	}
	if stripeID := strings.TrimSpace(req.StripePriceID); stripeID != "" {
		params.StripePriceID = pgtype.Text{String: stripeID, Valid: true}
	}

	plan, err := s.queries.CreatePlan(ctx, params)
	if err != nil {
		return AdminPlanResponse{}, err
	}
	if req.IsHighlighted {
		if err := s.clearOtherHighlights(ctx, slug); err != nil {
			return AdminPlanResponse{}, err
		}
	}
	return toAdminPlanResponse(plan, 0), nil
}

func (s *Service) AdminUpdatePlan(ctx context.Context, slug string, req UpdatePlanRequest) (AdminPlanResponse, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return AdminPlanResponse{}, ErrInvalidRequest
	}
	if _, err := s.queries.GetPlan(ctx, slug); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminPlanResponse{}, ErrNotFound
		}
		return AdminPlanResponse{}, err
	}
	if req.Name == nil && req.Description == nil && req.PriceCents == nil &&
		req.StripePriceID == nil && req.MessagesLimit == nil && req.ConnectionsLimit == nil &&
		req.RetentionDays == nil && req.WebhooksEnabled == nil && req.VideoCallsEnabled == nil &&
		req.ModerationEnabled == nil && req.IsActive == nil && req.IsHighlighted == nil &&
		req.SortOrder == nil {
		return AdminPlanResponse{}, ErrInvalidRequest
	}

	params := db.UpdatePlanParams{Slug: slug}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return AdminPlanResponse{}, ErrInvalidRequest
		}
		params.Name = pgtype.Text{String: name, Valid: true}
	}
	if req.Description != nil {
		params.Description = pgtype.Text{String: strings.TrimSpace(*req.Description), Valid: true}
	}
	if req.PriceCents != nil {
		if *req.PriceCents < 0 {
			return AdminPlanResponse{}, ErrInvalidRequest
		}
		params.PriceCents = pgtype.Int4{Int32: *req.PriceCents, Valid: true}
	}
	if req.StripePriceID != nil {
		params.StripePriceID = pgtype.Text{String: strings.TrimSpace(*req.StripePriceID), Valid: true}
	}
	if req.MessagesLimit != nil {
		params.MessagesLimit = pgtype.Int4{Int32: *req.MessagesLimit, Valid: true}
	}
	if req.ConnectionsLimit != nil {
		params.ConnectionsLimit = pgtype.Int4{Int32: *req.ConnectionsLimit, Valid: true}
	}
	if req.RetentionDays != nil {
		params.RetentionDays = pgtype.Int4{Int32: *req.RetentionDays, Valid: true}
	}
	if req.WebhooksEnabled != nil {
		params.WebhooksEnabled = pgtype.Bool{Bool: *req.WebhooksEnabled, Valid: true}
	}
	if req.VideoCallsEnabled != nil {
		params.VideoCallsEnabled = pgtype.Bool{Bool: *req.VideoCallsEnabled, Valid: true}
	}
	if req.ModerationEnabled != nil {
		params.ModerationEnabled = pgtype.Bool{Bool: *req.ModerationEnabled, Valid: true}
	}
	if req.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *req.IsActive, Valid: true}
	}
	if req.IsHighlighted != nil {
		params.IsHighlighted = pgtype.Bool{Bool: *req.IsHighlighted, Valid: true}
	}
	if req.SortOrder != nil {
		params.SortOrder = pgtype.Int4{Int32: *req.SortOrder, Valid: true}
	}

	plan, err := s.queries.UpdatePlan(ctx, params)
	if err != nil {
		return AdminPlanResponse{}, err
	}
	if req.IsHighlighted != nil && *req.IsHighlighted {
		if err := s.clearOtherHighlights(ctx, slug); err != nil {
			return AdminPlanResponse{}, err
		}
		plan, err = s.queries.GetPlan(ctx, slug)
		if err != nil {
			return AdminPlanResponse{}, err
		}
	}

	count, err := s.queries.CountSubscriptionsForPlan(ctx, slug)
	if err != nil {
		return AdminPlanResponse{}, err
	}
	return toAdminPlanResponse(plan, count), nil
}

func (s *Service) AdminDeletePlan(ctx context.Context, slug string) error {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return ErrInvalidRequest
	}
	if _, err := s.queries.GetPlan(ctx, slug); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	count, err := s.queries.CountSubscriptionsForPlan(ctx, slug)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrPlanInUse
	}
	return s.queries.DeletePlan(ctx, slug)
}

func (s *Service) clearOtherHighlights(ctx context.Context, keepSlug string) error {
	plans, err := s.queries.ListPlansAdmin(ctx)
	if err != nil {
		return err
	}
	for _, plan := range plans {
		if plan.Slug == keepSlug || !plan.IsHighlighted {
			continue
		}
		_, err := s.queries.UpdatePlan(ctx, db.UpdatePlanParams{
			Slug:          plan.Slug,
			IsHighlighted: pgtype.Bool{Bool: false, Valid: true},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func toSubscriptionResponse(sub db.GetAppSubscriptionRow) SubscriptionResponse {
	resp := SubscriptionResponse{
		PlanSlug:          sub.PlanSlug,
		PlanName:          sub.PlanName,
		Status:            sub.Status,
		PriceCents:        sub.PriceCents,
		MessagesLimit:     sub.MessagesLimit,
		ConnectionsLimit:  sub.ConnectionsLimit,
		RetentionDays:     sub.RetentionDays,
		WebhooksEnabled:   sub.WebhooksEnabled,
		VideoCallsEnabled: sub.VideoCallsEnabled,
		ModerationEnabled: sub.ModerationEnabled,
		NeedsPayment:      sub.Status == StatusIncomplete || sub.Status == StatusPastDue,
	}
	if sub.CurrentPeriodEnd.Valid {
		resp.CurrentPeriodEnd = formatTime(sub.CurrentPeriodEnd)
	}
	return resp
}

func formatTime(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.UTC().Format(time.RFC3339Nano)
}

func ReadWebhookBody(r *http.Request, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		maxBytes = 65536
	}
	r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)
	return io.ReadAll(r.Body)
}
