package models

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/argon2"
)

// User represents a chat user.
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"-"`          // Stored as Argon2id hash
	PublicKey string `json:"public_key"` // ECC public key (base64 or hex encoded)
	Status    string `json:"status"`
	LastSeen  string `json:"last_seen"`
}

// Argon2id parameters
const (
	ArgonTime    uint32 = 1
	ArgonMemory  uint32 = 64 * 1024 // 64 MB
	ArgonThreads uint8  = 2
	ArgonKeyLen  uint32 = 32
	ArgonSaltLen        = 16
)

// HashPassword hashes a plain password using Argon2id.
func HashPassword(password string) (string, error) {
	salt := make([]byte, ArgonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	// Format: $argon2id$v=19$m=65536,t=1,p=2$<salt>$<hash>
	return "$argon2id$v=19$m=65536,t=1,p=2$" + b64Salt + "$" + b64Hash, nil
}

// VerifyPassword checks if the provided password matches the Argon2id hash.
func VerifyPassword(hash, password string) (bool, error) {
	parts := splitHash(hash)
	if parts == nil {
		return false, errors.New("invalid hash format")
	}
	if len(parts) != 2 {
		return false, errors.New("invalid hash format")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, err
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}
	computed := argon2.IDKey([]byte(password), salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeyLen)
	return subtleCompare(computed, expectedHash), nil
}

// splitHash splits the hash string into its components.
func splitHash(hash string) []string {
	// Format: $argon2id$v=19$m=65536,t=1,p=2$<salt>$<hash>
	hash = strings.TrimSpace(hash)
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return nil
	}
	// parts[4] = salt, parts[5] = hash
	return []string{parts[4], parts[5]}
}

// subtleCompare compares two byte slices for equality without leaking timing info.
func subtleCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// splitN splits s into n parts using sep.
func splitN(s, sep string, n int) []string {
	out := make([]string, 0, n)
	for i := 0; i < n-1; i++ {
		idx := indexOf(s, sep)
		if idx < 0 {
			break
		}
		out = append(out, s[:idx])
		s = s[idx+len(sep):]
	}
	out = append(out, s)
	return out
}

// indexOf returns the index of the first occurrence of sep in s, or -1.
func indexOf(s, sep string) int {
	for i := 0; i+len(sep) <= len(s); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}
