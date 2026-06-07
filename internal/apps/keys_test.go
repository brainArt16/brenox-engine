package apps

import (
	"strings"
	"testing"
)

func TestGenerateAPIKey(t *testing.T) {
	plain, prefix, hash, err := GenerateAPIKey(false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(plain, "bx_live_") {
		t.Fatalf("expected live prefix, got %q", plain)
	}
	if prefix != plain[:16] {
		t.Fatalf("prefix mismatch")
	}
	if HashAPIKey(plain) != hash {
		t.Fatal("hash mismatch")
	}
}
