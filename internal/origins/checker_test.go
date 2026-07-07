package origins

import "testing"

func TestNormalizeList(t *testing.T) {
	t.Parallel()

	got, err := NormalizeList([]string{
		"https://app.example.com/",
		"http://localhost:3000",
		"https://app.example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 origins, got %d", len(got))
	}
}

func TestValidateRejectsPath(t *testing.T) {
	t.Parallel()

	if err := Validate("https://app.example.com/chat"); err == nil {
		t.Fatal("expected path origin to be rejected")
	}
}

func TestHintsFromRequest(t *testing.T) {
	t.Parallel()

	hints := HintsFromRequest(
		"GET",
		"/api/workspaces/42/channels/7/messages",
		"",
		9,
	)
	if hints.WorkspaceID != 42 {
		t.Fatalf("workspace_id = %d, want 42", hints.WorkspaceID)
	}
	if hints.AppID != 9 {
		t.Fatalf("app_id = %d, want 9", hints.AppID)
	}
}

func TestWorkspaceIDFromQuery(t *testing.T) {
	t.Parallel()

	if got := workspaceIDFromQuery("workspace_id=15&channel_id=3"); got != 15 {
		t.Fatalf("workspace_id = %d, want 15", got)
	}
}
