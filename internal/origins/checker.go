package origins

import (
	"context"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	db "github.com/brainart16/brenox/internal/db"
	brenoxjwt "github.com/brainart16/brenox/pkg/jwt"
)

const (
	maxOriginsPerApp = 20
	cacheTTL         = 60 * time.Second
)

type Hints struct {
	AppID       int64
	WorkspaceID int64
	Preflight   bool
}

type Checker struct {
	queries  *db.Queries
	platform []string
	mu       sync.RWMutex
	cache    snapshot
	loadedAt time.Time
}

type snapshot struct {
	byApp       map[int64][]string
	byWorkspace map[int64][]string
	union       []string
}

func NewChecker(queries *db.Queries) *Checker {
	return &Checker{
		queries:  queries,
		platform: loadPlatformOrigins(),
	}
}

func loadPlatformOrigins() []string {
	cors := ParseList(os.Getenv("CORS_ALLOWED_ORIGINS"))
	ws := ParseList(os.Getenv("WS_ALLOWED_ORIGINS"))
	return uniqueOrigins(append(cors, ws...))
}

func ParseList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "*" {
		return nil
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := Normalize(strings.TrimSpace(part))
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func Normalize(origin string) string {
	origin = strings.TrimSpace(origin)
	origin = strings.TrimRight(origin, "/")
	return origin
}

func Validate(origin string) error {
	origin = Normalize(origin)
	if origin == "" {
		return ErrOriginRequired
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return ErrInvalidOrigin
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidOrigin
	}
	if parsed.Host == "" {
		return ErrInvalidOrigin
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return ErrInvalidOrigin
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return ErrInvalidOrigin
	}
	return nil
}

func NormalizeList(origins []string) ([]string, error) {
	if len(origins) > maxOriginsPerApp {
		return nil, ErrTooManyOrigins
	}

	seen := make(map[string]struct{}, len(origins))
	out := make([]string, 0, len(origins))
	for _, item := range origins {
		normalized := Normalize(item)
		if err := Validate(normalized); err != nil {
			return nil, err
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out, nil
}

func (c *Checker) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loadedAt = time.Time{}
}

func (c *Checker) IsAllowed(ctx context.Context, origin string, hints Hints) bool {
	origin = Normalize(origin)
	if origin == "" {
		return true
	}
	if c.platformAllowed(origin) {
		return true
	}

	snap := c.snapshot(ctx)
	if hints.AppID > 0 {
		return contains(snap.byApp[hints.AppID], origin)
	}
	if hints.WorkspaceID > 0 {
		return contains(snap.byWorkspace[hints.WorkspaceID], origin)
	}
	if hints.Preflight {
		return contains(snap.union, origin)
	}
	return false
}

func (c *Checker) platformAllowed(origin string) bool {
	return contains(c.platform, origin)
}

func (c *Checker) snapshot(ctx context.Context) snapshot {
	c.mu.RLock()
	if time.Since(c.loadedAt) < cacheTTL && c.cache.byApp != nil {
		defer c.mu.RUnlock()
		return c.cache
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Since(c.loadedAt) < cacheTTL && c.cache.byApp != nil {
		return c.cache
	}

	rows, err := c.queries.ListAppOriginEntries(ctx)
	if err != nil {
		return c.cache
	}

	next := snapshot{
		byApp:       make(map[int64][]string, len(rows)),
		byWorkspace: make(map[int64][]string, len(rows)),
	}
	unionSet := make(map[string]struct{})
	for _, row := range rows {
		origins := normalizeStored(row.AllowedOrigins)
		next.byApp[row.ID] = origins
		next.byWorkspace[row.WorkspaceID] = origins
		for _, origin := range origins {
			unionSet[origin] = struct{}{}
		}
	}
	next.union = keys(unionSet)
	c.cache = next
	c.loadedAt = time.Now()
	return c.cache
}

func normalizeStored(origins []string) []string {
	out := make([]string, 0, len(origins))
	for _, origin := range origins {
		normalized := Normalize(origin)
		if normalized != "" {
			out = append(out, normalized)
		}
	}
	return out
}

func contains(list []string, origin string) bool {
	for _, item := range list {
		if item == origin {
			return true
		}
	}
	return false
}

func uniqueOrigins(origins []string) []string {
	set := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		set[origin] = struct{}{}
	}
	return keys(set)
}

func keys(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for origin := range set {
		out = append(out, origin)
	}
	return out
}

func HintsFromRequest(method, path, query string, appID int64) Hints {
	hints := Hints{AppID: appID}
	if method == "OPTIONS" {
		hints.Preflight = true
	}
	if workspaceID := workspaceIDFromPath(path); workspaceID > 0 {
		hints.WorkspaceID = workspaceID
	}
	if hints.WorkspaceID == 0 {
		hints.WorkspaceID = workspaceIDFromQuery(query)
	}
	return hints
}

func workspaceIDFromPath(path string) int64 {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "workspaces" {
			id, err := strconv.ParseInt(parts[i+1], 10, 64)
			if err == nil {
				return id
			}
		}
	}
	return 0
}

func workspaceIDFromQuery(query string) int64 {
	values, err := url.ParseQuery(query)
	if err != nil {
		return 0
	}
	raw := strings.TrimSpace(values.Get("workspace_id"))
	if raw == "" {
		return 0
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

// AppIDFromToken extracts app_id from a bearer or query token without DB checks.
func AppIDFromToken(authHeader, queryToken string) int64 {
	tokenString := strings.TrimSpace(queryToken)
	if tokenString == "" {
		const prefix = "Bearer "
		if strings.HasPrefix(authHeader, prefix) {
			tokenString = strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
		}
	}
	if tokenString == "" {
		return 0
	}

	claims, err := brenoxjwt.ValidateToken(tokenString)
	if err != nil {
		return 0
	}
	return claims.AppID
}
