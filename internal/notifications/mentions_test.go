package notifications

import "testing"

func TestParseMentions(t *testing.T) {
	got := ParseMentions("hey @alice and @bob @alice")
	if len(got) != 2 {
		t.Fatalf("expected 2 mentions, got %d", len(got))
	}
	if got[0] != "alice" || got[1] != "bob" {
		t.Fatalf("unexpected mentions: %#v", got)
	}
}

func TestParseMentionsEmpty(t *testing.T) {
	if ParseMentions("no mentions here") != nil {
		t.Fatal("expected nil")
	}
}
