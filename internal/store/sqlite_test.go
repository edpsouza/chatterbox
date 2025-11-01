package store

import (
	"os"
	"testing"

	"github.com/edpsouza/chatterbox/internal/models"
)

func TestStore_CreateAndFetchMessage(t *testing.T) {
	// Use a temporary database file for testing
	dbPath := "test_chatterbox.db"
	defer os.Remove(dbPath)

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create test user
	user := &models.User{
		Username:  "testuser",
		Password:  "hashedpassword",
		PublicKey: "testpublickey",
	}
	err = store.CreateUser(user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Fetch user to get ID
	fetchedUser, err := store.GetUserByUsername("testuser")
	if err != nil || fetchedUser == nil {
		t.Fatalf("failed to fetch user: %v", err)
	}

	// Create a message
	err = store.CreateMessage(fetchedUser.ID, "testuser", "recipientuser", "ciphertext123")
	if err != nil {
		t.Fatalf("failed to create message: %v", err)
	}

	// Fetch messages between testuser and recipientuser
	messages, err := store.GetMessagesBetween("testuser", "recipientuser")
	if err != nil {
		t.Fatalf("failed to fetch messages: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	if messages[0].Content != "ciphertext123" {
		t.Errorf("expected message content 'ciphertext123', got '%s'", messages[0].Content)
	}
	if messages[0].Username != "testuser" || messages[0].Recipient != "recipientuser" {
		t.Errorf("unexpected sender/recipient: got %s -> %s", messages[0].Username, messages[0].Recipient)
	}
}
