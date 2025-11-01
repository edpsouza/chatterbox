package models

import (
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	password := "supersecret123"

	// Hash the password
	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hashed == "" {
		t.Fatal("HashPassword returned empty string")
	}
	t.Logf("DEBUG: Hashed password: %s", hashed)

	// Verify correct password
	ok, err := VerifyPassword(hashed, password)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !ok {
		t.Error("VerifyPassword did not verify correct password")
	}

	// Verify incorrect password
	ok, err = VerifyPassword(hashed, "wrongpassword")
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if ok {
		t.Error("VerifyPassword verified incorrect password")
	}
}
