package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/edpsouza/chatterbox/internal/store"
)

// UserHandler dispatches /users/:username/public_key and /users/:username/presence endpoints.
func UserHandler(storeInstance *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Path: /users/:username/public_key or /users/:username/presence
		parts := strings.FieldsFunc(r.URL.Path, func(r rune) bool { return r == '/' })
		if len(parts) < 3 || parts[1] == "" {
			http.Error(w, "Invalid path. Use /users/:username/public_key or /users/:username/presence", http.StatusBadRequest)
			return
		}
		username := parts[1]
		action := parts[2]

		switch action {
		case "public_key":
			handlePublicKey(storeInstance, w, r, username)
		case "presence":
			handlePresence(storeInstance, w, r, username)
		default:
			http.Error(w, "Unknown action. Use /public_key or /presence", http.StatusNotFound)
		}
	}
}

// handlePublicKey serves the user's public key.
func handlePublicKey(storeInstance *store.Store, w http.ResponseWriter, r *http.Request, username string) {
	user, err := storeInstance.GetUserByUsername(username)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	resp := map[string]string{
		"username":   user.Username,
		"public_key": user.PublicKey,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// PresenceResponse represents the user's presence info.
type PresenceResponse struct {
	Username string `json:"username"`
	Status   string `json:"status"`
	LastSeen string `json:"last_seen,omitempty"`
}

// handlePresence serves user presence info (status and last_seen).
func handlePresence(storeInstance *store.Store, w http.ResponseWriter, r *http.Request, username string) {
	user, err := storeInstance.GetUserByUsername(username)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	resp := PresenceResponse{
		Username: user.Username,
		Status:   user.Status,
		LastSeen: user.LastSeen,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
