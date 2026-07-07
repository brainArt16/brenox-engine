package sandbox

import (
	"context"
	"log/slog"
	"time"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type Cleaner struct {
	queries *db.Queries
	cfg     Config
}

func NewCleaner(queries *db.Queries, cfg Config) *Cleaner {
	return &Cleaner{queries: queries, cfg: cfg}
}

func (c *Cleaner) RunOnce(ctx context.Context) error {
	if c.cfg.DataTTL <= 0 {
		return nil
	}

	cutoff := time.Now().UTC().Add(-c.cfg.DataTTL)
	cutoffParam := pgtype.Timestamptz{Time: cutoff, Valid: true}
	messages, err := c.queries.DeleteExpiredSandboxMessages(ctx, cutoffParam)
	if err != nil {
		return err
	}
	channels, err := c.queries.DeleteExpiredEmptySandboxChannels(ctx, cutoffParam)
	if err != nil {
		return err
	}

	if messages > 0 || channels > 0 {
		slog.Info("sandbox cleanup completed", "messages", messages, "channels", channels, "cutoff", cutoff.Format(time.RFC3339))
	}
	return nil
}

func (c *Cleaner) Start(ctx context.Context) {
	if c.cfg.DataTTL <= 0 || c.cfg.CleanupInterval <= 0 {
		return
	}

	go func() {
		if err := c.RunOnce(ctx); err != nil {
			slog.Warn("sandbox cleanup failed", "error", err)
		}

		ticker := time.NewTicker(c.cfg.CleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.RunOnce(ctx); err != nil {
					slog.Warn("sandbox cleanup failed", "error", err)
				}
			}
		}
	}()
}
