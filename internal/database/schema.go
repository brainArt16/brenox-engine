package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const ExpectedMigrationVersion int64 = 13

type MigrationStatus struct {
	Version int64
	Dirty   bool
	OK      bool
	Message string
}

func CheckMigrations(ctx context.Context, pool *pgxpool.Pool) (MigrationStatus, error) {
	var version int64
	var dirty bool

	err := pool.QueryRow(ctx, `
		SELECT version, dirty
		FROM schema_migrations
		ORDER BY version DESC
		LIMIT 1
	`).Scan(&version, &dirty)

	if err != nil {
		return MigrationStatus{
			OK:      false,
			Message: "schema_migrations table missing — run database migrations before starting the API",
		}, nil
	}

	status := MigrationStatus{
		Version: version,
		Dirty:   dirty,
	}

	switch {
	case dirty:
		status.OK = false
		status.Message = fmt.Sprintf("migration %d is dirty — fix schema_migrations before serving traffic", version)
	case version < ExpectedMigrationVersion:
		status.OK = false
		status.Message = fmt.Sprintf(
			"database schema is at version %d, expected %d — run migrations (missing audit_logs, revoked_tokens, and/or platform admin columns)",
			version,
			ExpectedMigrationVersion,
		)
	default:
		status.OK = true
		status.Message = "ok"
	}

	return status, nil
}
