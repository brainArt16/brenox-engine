package apps

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"
	"time"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/brainart16/brenox/internal/origins"
	"github.com/brainart16/brenox/internal/sandbox"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Service struct {
	queries *db.Queries
	billing BillingHook
	origins *origins.Checker
	sandbox sandbox.Config
}

type BillingHook interface {
	OnAppCreated(ctx context.Context, appID int64, planSlug string) error
	CheckCanCreateWebhook(ctx context.Context, appID int64) error
}

func NewService(queries *db.Queries) *Service {
	return &Service{queries: queries, sandbox: sandbox.LoadConfig()}
}

func (s *Service) SetBilling(hook BillingHook) {
	s.billing = hook
}

func (s *Service) SetOriginChecker(checker *origins.Checker) {
	s.origins = checker
}

type AuthenticatedApp struct {
	App    db.App
	APIKey db.ApiKey
}

func (s *Service) CreateApp(ctx context.Context, ownerID int64, req CreateAppRequest) (AppResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return AppResponse{}, ErrNameRequired
	}

	slug, err := normalizeSlug(req.Slug, name)
	if err != nil {
		return AppResponse{}, err
	}

	if _, err := s.queries.GetAppBySlug(ctx, slug); err == nil {
		return AppResponse{}, ErrSlugTaken
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return AppResponse{}, err
	}

	workspace, err := s.queries.CreateWorkspace(ctx, db.CreateWorkspaceParams{
		Name:    name + " App",
		Slug:    "app-" + slug,
		OwnerID: ownerID,
	})
	if err != nil {
		return AppResponse{}, err
	}

	sandboxWorkspace, err := s.queries.CreateWorkspace(ctx, db.CreateWorkspaceParams{
		Name:    name + " Sandbox",
		Slug:    "app-" + slug + "-sandbox",
		OwnerID: ownerID,
	})
	if err != nil {
		return AppResponse{}, err
	}

	if err := s.queries.AddWorkspaceMember(ctx, db.AddWorkspaceMemberParams{
		WorkspaceID: workspace.ID,
		UserID:      ownerID,
		Role:        "owner",
	}); err != nil {
		return AppResponse{}, err
	}

	if err := s.queries.AddWorkspaceMember(ctx, db.AddWorkspaceMemberParams{
		WorkspaceID: sandboxWorkspace.ID,
		UserID:      ownerID,
		Role:        "owner",
	}); err != nil {
		return AppResponse{}, err
	}

	app, err := s.queries.CreateApp(ctx, db.CreateAppParams{
		Name:               name,
		Slug:               slug,
		WorkspaceID:        workspace.ID,
		SandboxWorkspaceID: sandboxWorkspace.ID,
		OwnerID:            ownerID,
	})
	if err != nil {
		return AppResponse{}, err
	}

	if s.billing != nil {
		if err := s.billing.OnAppCreated(ctx, app.ID, req.PlanSlug); err != nil {
			return AppResponse{}, err
		}
	}

	return toAppResponse(app), nil
}

func (s *Service) ListApps(ctx context.Context, ownerID int64) ([]AppResponse, error) {
	rows, err := s.queries.ListAppsByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	items := make([]AppResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAppResponse(row))
	}
	return items, nil
}

func (s *Service) CreateAPIKey(ctx context.Context, appID, ownerID int64, req CreateAPIKeyRequest) (APIKeyCreatedResponse, error) {
	app, err := s.assertAppOwner(ctx, appID, ownerID)
	if err != nil {
		return APIKeyCreatedResponse{}, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "default"
	}

	plain, prefix, hash, err := GenerateAPIKey(req.Sandbox)
	if err != nil {
		return APIKeyCreatedResponse{}, err
	}

	expiresAt := pgtype.Timestamptz{}
	if req.Sandbox && s.sandbox.APIKeyTTL > 0 {
		expiresAt = pgtype.Timestamptz{Time: time.Now().UTC().Add(s.sandbox.APIKeyTTL), Valid: true}
	}

	key, err := s.queries.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		AppID:     app.ID,
		Name:      name,
		KeyPrefix: prefix,
		KeyHash:   hash,
		IsSandbox: req.Sandbox,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return APIKeyCreatedResponse{}, err
	}

	return APIKeyCreatedResponse{
		APIKeyResponse: toAPIKeyResponse(key),
		Secret:         plain,
	}, nil
}

func (s *Service) ListAPIKeys(ctx context.Context, appID, ownerID int64) ([]APIKeyResponse, error) {
	if _, err := s.assertAppOwner(ctx, appID, ownerID); err != nil {
		return nil, err
	}

	rows, err := s.queries.ListAPIKeysByApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	items := make([]APIKeyResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAPIKeyResponse(row))
	}
	return items, nil
}

func (s *Service) RevokeAPIKey(ctx context.Context, appID, keyID, ownerID int64) error {
	if _, err := s.assertAppOwner(ctx, appID, ownerID); err != nil {
		return err
	}

	_, err := s.queries.RevokeAPIKey(ctx, db.RevokeAPIKeyParams{
		ID:    keyID,
		AppID: appID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrKeyNotFound
		}
		return err
	}
	return nil
}

func (s *Service) AuthenticateAPIKey(ctx context.Context, plainKey string) (AuthenticatedApp, error) {
	if !IsAPIKeyToken(plainKey) {
		return AuthenticatedApp{}, ErrInvalidKey
	}

	key, err := s.queries.GetAPIKeyByPrefix(ctx, LookupPrefix(plainKey))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AuthenticatedApp{}, ErrInvalidKey
		}
		return AuthenticatedApp{}, err
	}

	if key.KeyHash != HashAPIKey(plainKey) {
		return AuthenticatedApp{}, ErrInvalidKey
	}
	if key.IsSandbox && key.ExpiresAt.Valid && !key.ExpiresAt.Time.After(time.Now()) {
		return AuthenticatedApp{}, ErrInvalidKey
	}

	app, err := s.queries.GetAppByID(ctx, key.AppID)
	if err != nil {
		return AuthenticatedApp{}, ErrInvalidKey
	}

	_ = s.queries.TouchAPIKeyLastUsed(ctx, key.ID)

	return AuthenticatedApp{App: app, APIKey: key}, nil
}

func (s *Service) CreateWebhook(ctx context.Context, appID, ownerID int64, req CreateWebhookRequest) (WebhookResponse, error) {
	app, err := s.assertAppOwner(ctx, appID, ownerID)
	if err != nil {
		return WebhookResponse{}, err
	}

	url := strings.TrimSpace(req.URL)
	if url == "" {
		return WebhookResponse{}, ErrWebhookURLRequired
	}

	if s.billing != nil {
		if err := s.billing.CheckCanCreateWebhook(ctx, app.ID); err != nil {
			return WebhookResponse{}, err
		}
	}

	events := req.Events
	if len(events) == 0 {
		events = []string{"message.created", "user.provisioned", "channel.created"}
	}

	secret, err := randomSecret()
	if err != nil {
		return WebhookResponse{}, err
	}

	webhook, err := s.queries.CreateWebhook(ctx, db.CreateWebhookParams{
		AppID:  app.ID,
		Url:    url,
		Events: events,
		Secret: secret,
	})
	if err != nil {
		return WebhookResponse{}, err
	}

	resp := toWebhookResponse(webhook)
	resp.Secret = secret
	return resp, nil
}

func (s *Service) GetApp(ctx context.Context, appID int64) (db.App, error) {
	app, err := s.queries.GetAppByID(ctx, appID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.App{}, ErrNotFound
		}
		return db.App{}, err
	}
	return app, nil
}

func (s *Service) GetAppForOwner(ctx context.Context, appID, ownerID int64) (AppResponse, error) {
	app, err := s.assertAppOwner(ctx, appID, ownerID)
	if err != nil {
		return AppResponse{}, err
	}
	return toAppResponse(app), nil
}

func (s *Service) UpdateAllowedOrigins(ctx context.Context, appID, ownerID int64, req UpdateAllowedOriginsRequest) (AppResponse, error) {
	if _, err := s.assertAppOwner(ctx, appID, ownerID); err != nil {
		return AppResponse{}, err
	}

	normalized, err := origins.NormalizeList(req.AllowedOrigins)
	if err != nil {
		switch {
		case errors.Is(err, origins.ErrInvalidOrigin):
			return AppResponse{}, ErrInvalidOrigin
		case errors.Is(err, origins.ErrTooManyOrigins):
			return AppResponse{}, ErrTooManyOrigins
		default:
			return AppResponse{}, err
		}
	}

	app, err := s.queries.UpdateAppAllowedOrigins(ctx, db.UpdateAppAllowedOriginsParams{
		ID:             appID,
		AllowedOrigins: normalized,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AppResponse{}, ErrNotFound
		}
		return AppResponse{}, err
	}

	if s.origins != nil {
		s.origins.Invalidate()
	}

	return toAppResponse(app), nil
}

func (s *Service) ListWebhooks(ctx context.Context, appID, ownerID int64) ([]WebhookResponse, error) {
	if _, err := s.assertAppOwner(ctx, appID, ownerID); err != nil {
		return nil, err
	}

	rows, err := s.queries.ListWebhooksByApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	items := make([]WebhookResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toWebhookResponse(row))
	}
	return items, nil
}

func (s *Service) DisableWebhook(ctx context.Context, appID, webhookID, ownerID int64) error {
	if _, err := s.assertAppOwner(ctx, appID, ownerID); err != nil {
		return err
	}

	rows, err := s.queries.DisableWebhook(ctx, db.DisableWebhookParams{
		ID:    webhookID,
		AppID: appID,
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrWebhookNotFound
	}
	return nil
}

func (s *Service) assertAppOwner(ctx context.Context, appID, ownerID int64) (db.App, error) {
	app, err := s.queries.GetAppByID(ctx, appID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.App{}, ErrNotFound
		}
		return db.App{}, err
	}
	if app.OwnerID != ownerID {
		return db.App{}, ErrForbidden
	}
	return app, nil
}

func normalizeSlug(raw, name string) (string, error) {
	slug := strings.TrimSpace(raw)
	if slug == "" {
		slug = strings.ToLower(name)
		slug = strings.ReplaceAll(slug, " ", "-")
	}
	slug = strings.ToLower(slug)
	if !slugPattern.MatchString(slug) {
		return "", ErrInvalidSlug
	}
	return slug, nil
}

func randomSecret() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func toAppResponse(app db.App) AppResponse {
	allowed := app.AllowedOrigins
	if allowed == nil {
		allowed = []string{}
	}
	return AppResponse{
		ID:                 app.ID,
		Name:               app.Name,
		Slug:               app.Slug,
		WorkspaceID:        app.WorkspaceID,
		SandboxWorkspaceID: app.SandboxWorkspaceID,
		OwnerID:            app.OwnerID,
		CreatedAt:          formatTime(app.CreatedAt),
		AllowedOrigins:     allowed,
	}
}

func toAPIKeyResponse(key db.ApiKey) APIKeyResponse {
	resp := APIKeyResponse{
		ID:        key.ID,
		AppID:     key.AppID,
		Name:      key.Name,
		KeyPrefix: key.KeyPrefix,
		IsSandbox: key.IsSandbox,
		CreatedAt: formatTime(key.CreatedAt),
	}
	if key.RevokedAt.Valid {
		resp.RevokedAt = formatTime(key.RevokedAt)
	}
	if key.LastUsedAt.Valid {
		resp.LastUsed = formatTime(key.LastUsedAt)
	}
	if key.ExpiresAt.Valid {
		resp.ExpiresAt = formatTime(key.ExpiresAt)
	}
	return resp
}

func toWebhookResponse(webhook db.Webhook) WebhookResponse {
	return WebhookResponse{
		ID:        webhook.ID,
		AppID:     webhook.AppID,
		URL:       webhook.Url,
		Events:    webhook.Events,
		CreatedAt: formatTime(webhook.CreatedAt),
	}
}

func formatTime(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.UTC().Format(time.RFC3339Nano)
}
