package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/edpsouza/chatterbox/internal/store"
)

func setupTestStore(t *testing.T) *store.Store {
	dbPath := "test_auth.db"
	t.Cleanup(func() { os.Remove(dbPath) })
	storeInstance, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	return storeInstance
}

func TestRegisterAndLoginHandler(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	storeInstance := setupTestStore(t)

	// Register user
	registerPayload := map[string]string{
		"username":   "testuser",
		"password":   "testpassword",
		"public_key": "testpublickey",
	}
	t.Logf("Register payload: %+v", registerPayload)
	body, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := RegisterHandler(storeInstance)
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", resp.StatusCode)
	}

	// Fetch user and print password hash for debug
	user, err := storeInstance.GetUserByUsername("testuser")
	if err != nil || user == nil {
		t.Fatalf("failed to fetch user after registration: %v", err)
	}
	t.Logf("Registered user password hash: %s", user.Password)

	// Login user
	loginPayload := map[string]string{
		"username": "testuser",
		"password": "testpassword",
	}
	t.Logf("Login payload: %+v", loginPayload)
	body, _ = json.Marshal(loginPayload)
	req = httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	handler = LoginHandler(storeInstance)
	handler(w, req)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}

	// Check response contains a token
	var respBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	if _, ok := respBody["token"]; !ok {
		t.Errorf("login response missing 'token' field: %v", respBody)
	}
}
