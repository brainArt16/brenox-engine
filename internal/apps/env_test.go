package apps

import (
	"testing"

	db "github.com/brainart16/brenox/internal/db"
)

func TestWorkspaceForEnvironment(t *testing.T) {
	app := db.App{
		WorkspaceID:        10,
		SandboxWorkspaceID: 20,
	}

	if got := WorkspaceForEnvironment(app, EnvLive); got != 10 {
		t.Fatalf("live workspace = %d, want 10", got)
	}
	if got := WorkspaceForEnvironment(app, EnvSandbox); got != 20 {
		t.Fatalf("sandbox workspace = %d, want 20", got)
	}
}

func TestEnvironmentFromSandbox(t *testing.T) {
	if EnvironmentFromSandbox(true) != EnvSandbox {
		t.Fatal("expected sandbox env")
	}
	if EnvironmentFromSandbox(false) != EnvLive {
		t.Fatal("expected live env")
	}
}
