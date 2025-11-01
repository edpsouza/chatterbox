package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/edpsouza/chatterbox/internal/store"
)

// MessageHistoryHandler serves encrypted message history between authenticated user and another user.
// Endpoint: GET /messages/:with_user
func MessageHistoryHandler(storeInstance *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple auth: get username from header (replace with JWT/session in production)
		username := r.Header.Get("X-Username")
		if username == "" {
			http.Error(w, "Unauthorized: missing X-Username header", http.StatusUnauthorized)
			return
		}

		// Extract with_user from URL path: /messages/:with_user
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 || parts[2] == "" {
			http.Error(w, "Missing with_user in path", http.StatusBadRequest)
			return
		}
		withUser := parts[2]

		// Fetch messages between username and withUser
		messages, err := storeInstance.GetMessagesBetween(username, withUser)
		if err != nil {
			http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
			return
		}

		// Return as JSON, including recipient field
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	}
}
