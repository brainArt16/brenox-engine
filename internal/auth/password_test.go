package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckPassword("secret123", hash); err != nil {
		t.Fatalf("expected password match: %v", err)
	}
	if err := CheckPassword("wrong", hash); err == nil {
		t.Fatal("expected password mismatch")
	}
}
