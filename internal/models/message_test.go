package models

import (
	"testing"
)

func TestMessageFields(t *testing.T) {
	msg := Message{
		ID:        1,
		UserID:    42,
		Username:  "alice",
		Recipient: "bob",
		Content:   "encrypted-ciphertext",
		CreatedAt: "2024-06-10T12:34:56Z",
	}

	if msg.ID != 1 {
		t.Errorf("expected ID 1, got %d", msg.ID)
	}
	if msg.UserID != 42 {
		t.Errorf("expected UserID 42, got %d", msg.UserID)
	}
	if msg.Username != "alice" {
		t.Errorf("expected Username 'alice', got '%s'", msg.Username)
	}
	if msg.Recipient != "bob" {
		t.Errorf("expected Recipient 'bob', got '%s'", msg.Recipient)
	}
	if msg.Content != "encrypted-ciphertext" {
		t.Errorf("expected Content 'encrypted-ciphertext', got '%s'", msg.Content)
	}
	if msg.CreatedAt != "2024-06-10T12:34:56Z" {
		t.Errorf("expected CreatedAt '2024-06-10T12:34:56Z', got '%s'", msg.CreatedAt)
	}
}
