package chat

import "testing"

func TestNormalizeContent(t *testing.T) {
	t.Run("trims whitespace", func(t *testing.T) {
		got, err := normalizeContent("  hello  ", false)
		if err != nil || got != "hello" {
			t.Fatalf("got %q err %v", got, err)
		}
	})

	t.Run("rejects empty", func(t *testing.T) {
		_, err := normalizeContent("   ", false)
		if err != ErrEmptyContent {
			t.Fatalf("expected ErrEmptyContent, got %v", err)
		}
	})

	t.Run("allows empty with attachments", func(t *testing.T) {
		got, err := normalizeContent(" ", true)
		if err != nil || got != "" {
			t.Fatalf("got %q err %v", got, err)
		}
	})

	t.Run("rejects too long", func(t *testing.T) {
		content := make([]byte, MaxMessageLength+1)
		for i := range content {
			content[i] = 'a'
		}
		_, err := normalizeContent(string(content), false)
		if err != ErrMessageTooLong {
			t.Fatalf("expected ErrMessageTooLong, got %v", err)
		}
	})
}
