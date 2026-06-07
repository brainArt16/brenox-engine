package channels

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsDuplicateChannelName(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23505"}
	if !isDuplicateChannelName(pgErr) {
		t.Fatal("expected duplicate detection")
	}
	if isDuplicateChannelName(errors.New("other error")) {
		t.Fatal("unexpected duplicate detection")
	}
}
