package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/edpsouza/chatterbox/internal/models"
	"github.com/edpsouza/chatterbox/internal/store"
	"github.com/golang-jwt/jwt/v5"
)

// UserRequest for registration/login
type UserRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
}

// JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// RegisterHandler handles user registration using Store
func RegisterHandler(storeInstance *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req UserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if req.Username == "" || req.Password == "" || req.PublicKey == "" {
			http.Error(w, "Username, password, and public key required", http.StatusBadRequest)
			return
		}
		// Hash password
		hashed, err := models.HashPassword(req.Password)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		user := &models.User{
			Username:  req.Username,
			Password:  hashed,
			PublicKey: req.PublicKey,
		}
		err = storeInstance.CreateUser(user)
		if err != nil {
			if err.Error() == "username already exists" {
				http.Error(w, "Username already taken", http.StatusConflict)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"public_key": user.PublicKey,
		})
	}
}

// LoginHandler handles user login and JWT issuance using Store
func LoginHandler(storeInstance *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req UserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if req.Username == "" || req.Password == "" {
			http.Error(w, "Username and password required", http.StatusBadRequest)
			return
		}
		// Get user
		user, err := storeInstance.GetUserByUsername(req.Username)
		if err != nil || user == nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		ok, err := models.VerifyPassword(user.Password, req.Password)
		if err != nil || !ok {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		// Issue JWT
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			http.Error(w, "JWT secret not set", http.StatusInternalServerError)
			return
		}
		claims := Claims{
			UserID:   strconv.FormatInt(user.ID, 10),
			Username: user.Username,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte(secret))
		if err != nil {
			http.Error(w, "JWT error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"token": signed})
	}
}

// PublicKeyHandler handles GET /users/:username/public_key requests
func PublicKeyHandler(storeInstance *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract username from URL path
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 || parts[1] != "users" || parts[3] != "public_key" {
			http.Error(w, "Invalid endpoint", http.StatusBadRequest)
			return
		}
		username := parts[2]
		if username == "" {
			http.Error(w, "Username required", http.StatusBadRequest)
			return
		}
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
}
