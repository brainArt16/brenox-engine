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

// EnsureUserSubscription creates a free Starter subscription when missing.
func (s *Service) EnsureUserSubscription(ctx context.Context, userID int64) error {
	_, err := s.queries.GetUserSubscription(ctx, userID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	plan, err := s.queries.GetDefaultPlan(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidPlan
		}
		return err
	}
	_, err = s.queries.CreateUserSubscription(ctx, db.CreateUserSubscriptionParams{
		UserID:   userID,
		PlanSlug: plan.Slug,
		Status:   StatusActive,
	})
	return err
}

func (s *Service) OnUserRegistered(ctx context.Context, userID int64) error {
	return s.EnsureUserSubscription(ctx, userID)
}

func (s *Service) AssertCanCreateApp(ctx context.Context, userID int64) error {
	if err := s.EnsureUserSubscription(ctx, userID); err != nil {
		return err
	}
	sub, err := s.queries.GetUserSubscription(ctx, userID)
	if err != nil {
		return err
	}
	if !statusAllowsUsage(sub.Status) {
		return ErrSubscriptionInactive
	}
	if sub.MaxApps < 0 {
		return nil
	}
	count, err := s.queries.CountAppsByOwner(ctx, userID)
	if err != nil {
		return err
	}
	if count >= int64(sub.MaxApps) {
		return ErrAppLimit
	}
	return nil
}

func (s *Service) GetAccountBilling(ctx context.Context, userID int64) (AccountBillingResponse, error) {
	if err := s.EnsureUserSubscription(ctx, userID); err != nil {
		return AccountBillingResponse{}, err
	}
	sub, err := s.queries.GetUserSubscription(ctx, userID)
	if err != nil {
		return AccountBillingResponse{}, err
	}

	messages, err := s.queries.SumOwnerMessageUsage(ctx, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return AccountBillingResponse{}, err
	}

	appsCount, err := s.queries.CountAppsByOwner(ctx, userID)
	if err != nil {
		return AccountBillingResponse{}, err
	}

	breakdownRows, err := s.queries.ListOwnerAppMessageUsage(ctx, userID)
	if err != nil {
		return AccountBillingResponse{}, err
	}
	breakdown := make([]AppUsageBreakdown, 0, len(breakdownRows))
	for _, row := range breakdownRows {
		breakdown = append(breakdown, AppUsageBreakdown{
			AppID:             row.AppID,
			AppName:           row.AppName,
			AppSlug:           row.AppSlug,
			MessagesThisMonth: row.MessagesThisMonth,
		})
	}

	return AccountBillingResponse{
		Subscription: toSubscriptionResponse(sub),
		Usage: UsageResponse{
			MessagesThisMonth: messages,
			MessagesLimit:     sub.MessagesLimit,
		},
		AppsCount: appsCount,
		AppsLimit: sub.MaxApps,
		AppUsage:  breakdown,
	}, nil
}

func (s *Service) CreateCheckoutSession(ctx context.Context, userID int64, userEmail string, planSlug string) (CheckoutResponse, error) {
	if !s.config.Enabled() {
		return CheckoutResponse{}, ErrStripeNotConfigured
	}
	if err := s.EnsureUserSubscription(ctx, userID); err != nil {
		return CheckoutResponse{}, err
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
	if plan.PriceCents == 0 {
		return CheckoutResponse{}, ErrInvalidPlan
	}

	if !plan.StripePriceID.Valid || strings.TrimSpace(plan.StripePriceID.String) == "" {
		return CheckoutResponse{}, ErrPlanStripePriceMissing
	}
	priceID := plan.StripePriceID.String

	customerID, err := s.ensureStripeCustomer(ctx, userID, userEmail)
	if err != nil {
		return CheckoutResponse{}, err
	}

	successURL := fmt.Sprintf("%s/settings/billing?session_id={CHECKOUT_SESSION_ID}", s.config.CheckoutBaseURL)
	cancelURL := fmt.Sprintf("%s/settings/billing", s.config.CheckoutBaseURL)

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{Price: stripe.String(priceID), Quantity: stripe.Int64(1)},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"user_id":   strconv.FormatInt(userID, 10),
			"plan_slug": planSlug,
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"user_id":   strconv.FormatInt(userID, 10),
				"plan_slug": planSlug,
			},
		},
	}

	sess, err := checkoutsession.New(params)
	if err != nil {
		return CheckoutResponse{}, err
	}

	_, _ = s.queries.UpdateUserSubscription(ctx, db.UpdateUserSubscriptionParams{
		UserID:           userID,
		PlanSlug:         pgtype.Text{String: planSlug, Valid: true},
		StripeCustomerID: pgtype.Text{String: customerID, Valid: true},
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

func (s *Service) CheckMessageQuota(ctx context.Context, appID int64) error {
	app, err := s.queries.GetAppByID(ctx, appID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	return s.checkOwnerMessageQuota(ctx, app.OwnerID)
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
	return s.checkOwnerMessageQuota(ctx, app.OwnerID)
}

func (s *Service) checkOwnerMessageQuota(ctx context.Context, ownerID int64) error {
	if err := s.EnsureUserSubscription(ctx, ownerID); err != nil {
		return err
	}
	sub, err := s.queries.GetUserSubscription(ctx, ownerID)
	if err != nil {
		return err
	}
	if !statusAllowsUsage(sub.Status) {
		return ErrSubscriptionInactive
	}
	count, err := s.queries.SumOwnerMessageUsage(ctx, ownerID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if int64(sub.MessagesLimit) > 0 && count >= int64(sub.MessagesLimit) {
		return ErrMessageLimit
	}
	return nil
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
	app, err := s.queries.GetAppByID(ctx, appID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrWebhooksNotAllowed
		}
		return err
	}
	if err := s.EnsureUserSubscription(ctx, app.OwnerID); err != nil {
		return err
	}
	sub, err := s.queries.GetUserSubscription(ctx, app.OwnerID)
	if err != nil {
		return ErrWebhooksNotAllowed
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
	if err := s.EnsureUserSubscription(ctx, app.OwnerID); err != nil {
		return err
	}
	sub, err := s.queries.GetUserSubscription(ctx, app.OwnerID)
	if err != nil {
		return ErrVideoNotAllowed
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
	rows, err := s.queries.AdminListUserSubscriptions(ctx, db.AdminListUserSubscriptionsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]AdminSubscriptionListItem, 0, len(rows))
	for _, row := range rows {
		item := AdminSubscriptionListItem{
			UserID:            row.UserID,
			UserEmail:         row.UserEmail,
			UserUsername:      row.UserUsername,
			PlanSlug:          row.PlanSlug,
			PlanName:          row.PlanName,
			Status:            row.Status,
			MaxApps:           row.MaxApps,
			AppsCount:         row.AppsCount,
			MessagesThisMonth: row.MessagesThisMonth,
		}
		if row.CurrentPeriodEnd.Valid {
			item.CurrentPeriodEnd = formatTime(row.CurrentPeriodEnd)
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) AdminUpdateSubscription(ctx context.Context, userID int64, req AdminUpdateSubscriptionRequest) (AccountBillingResponse, error) {
	if req.PlanSlug == nil && req.Status == nil {
		return AccountBillingResponse{}, ErrInvalidRequest
	}
	if err := s.EnsureUserSubscription(ctx, userID); err != nil {
		return AccountBillingResponse{}, err
	}
	params := db.UpdateUserSubscriptionParams{UserID: userID}
	if req.PlanSlug != nil {
		slug := strings.TrimSpace(*req.PlanSlug)
		if _, err := s.queries.GetPlan(ctx, slug); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return AccountBillingResponse{}, ErrInvalidPlan
			}
			return AccountBillingResponse{}, err
		}
		params.PlanSlug = pgtype.Text{String: slug, Valid: true}
	}
	if req.Status != nil {
		params.Status = pgtype.Text{String: strings.TrimSpace(*req.Status), Valid: true}
	}
	if _, err := s.queries.UpdateUserSubscription(ctx, params); err != nil {
		return AccountBillingResponse{}, err
	}
	return s.GetAccountBilling(ctx, userID)
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
	userID, err := parseMetadataInt64(raw.Metadata, "user_id")
	if err != nil {
		return err
	}
	planSlug := raw.Metadata["plan_slug"]

	if err := s.EnsureUserSubscription(ctx, userID); err != nil {
		return err
	}

	params := db.UpdateUserSubscriptionParams{
		UserID: userID,
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
	_, err = s.queries.UpdateUserSubscription(ctx, params)
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
	existing, err := s.queries.GetUserSubscriptionByStripeSubscriptionID(ctx, stripeSubID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			userID, parseErr := parseMetadataInt64(sub.Metadata, "user_id")
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
			if ensureErr := s.EnsureUserSubscription(ctx, userID); ensureErr != nil {
				return ensureErr
			}
			_, updateErr := s.queries.UpdateUserSubscription(ctx, db.UpdateUserSubscriptionParams{
				UserID:               userID,
				PlanSlug:             pgtype.Text{String: planSlug, Valid: true},
				Status:               pgtype.Text{String: mapStripeStatus(string(sub.Status)), Valid: true},
				StripeSubscriptionID: stripeSubID,
			})
			return updateErr
		}
		return err
	}

	status := mapStripeStatus(string(sub.Status))
	if event.Type == "customer.subscription.deleted" {
		status = StatusCanceled
	}

	params := db.UpdateUserSubscriptionParams{
		UserID: existing.UserID,
		Status: pgtype.Text{String: status, Valid: true},
	}
	if planSlug := sub.Metadata["plan_slug"]; planSlug != "" {
		params.PlanSlug = pgtype.Text{String: planSlug, Valid: true}
	}
	_, err = s.queries.UpdateUserSubscription(ctx, params)
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
		PriceDisplay:      formatPriceDisplay(p.PriceCents),
		MessagesLimit:     p.MessagesLimit,
		ConnectionsLimit:  p.ConnectionsLimit,
		RetentionDays:     p.RetentionDays,
		MaxApps:           p.MaxApps,
		WebhooksEnabled:   p.WebhooksEnabled,
		VideoCallsEnabled: p.VideoCallsEnabled,
		ModerationEnabled: p.ModerationEnabled,
		IsHighlighted:     p.IsHighlighted,
		SortOrder:         p.SortOrder,
	}
}

func formatPriceDisplay(cents int32) string {
	if cents == 0 {
		return "Free"
	}
	return fmt.Sprintf("$%.0f", float64(cents)/100)
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
	if req.PriceCents < 0 || req.MessagesLimit < 0 || req.ConnectionsLimit < 0 || req.RetentionDays < 0 || req.MaxApps < -1 {
		return AdminPlanResponse{}, ErrInvalidRequest
	}
	maxApps := req.MaxApps
	if maxApps == 0 {
		maxApps = 1
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
		MaxApps:           maxApps,
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
		req.RetentionDays == nil && req.MaxApps == nil && req.WebhooksEnabled == nil &&
		req.VideoCallsEnabled == nil && req.ModerationEnabled == nil && req.IsActive == nil &&
		req.IsHighlighted == nil && req.SortOrder == nil {
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
	if req.MaxApps != nil {
		if *req.MaxApps < -1 {
			return AdminPlanResponse{}, ErrInvalidRequest
		}
		params.MaxApps = pgtype.Int4{Int32: *req.MaxApps, Valid: true}
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

func toSubscriptionResponse(sub db.GetUserSubscriptionRow) SubscriptionResponse {
	resp := SubscriptionResponse{
		PlanSlug:          sub.PlanSlug,
		PlanName:          sub.PlanName,
		Status:            sub.Status,
		PriceCents:        sub.PriceCents,
		MessagesLimit:     sub.MessagesLimit,
		ConnectionsLimit:  sub.ConnectionsLimit,
		RetentionDays:     sub.RetentionDays,
		MaxApps:           sub.MaxApps,
		WebhooksEnabled:   sub.WebhooksEnabled,
		VideoCallsEnabled: sub.VideoCallsEnabled,
		ModerationEnabled: sub.ModerationEnabled,
		NeedsPayment:      sub.Status == StatusPastDue,
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
