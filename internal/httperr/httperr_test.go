package httperr

import (
	"errors"
	"testing"
)

func TestIsSensitive(t *testing.T) {
	tests := []struct {
		message string
		want    bool
	}{
		{
			message: "invalid credentials",
			want:    false,
		},
		{
			message: "email already exists",
			want:    false,
		},
		{
			message: `failed to connect to user=brenox_admin database=brenox_db: hostname resolving error`,
			want:    true,
		},
		{
			message: "lookup jz8tgpcuopib9pctb29o8eai on 127.0.0.11:53: server misbehaving",
			want:    true,
		},
		{
			message: "dial tcp 127.0.0.1:5432: connection refused",
			want:    true,
		},
	}

	for _, tt := range tests {
		if got := IsSensitive(tt.message); got != tt.want {
			t.Fatalf("IsSensitive(%q) = %v, want %v", tt.message, got, tt.want)
		}
	}
}

func TestClientMessageKnownSentinel(t *testing.T) {
	err := errors.New("channel not found")
	got := ClientMessage(err, err)
	if got != "channel not found" {
		t.Fatalf("ClientMessage() = %q, want channel not found", got)
	}
}

func TestClientMessageUnknownError(t *testing.T) {
	err := errors.New(`failed to connect to user=admin database=app`)
	got := ClientMessage(err)
	if got != InternalMessage {
		t.Fatalf("ClientMessage() = %q, want %q", got, InternalMessage)
	}
}
