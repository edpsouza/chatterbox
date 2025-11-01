package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// GetEncryptionKey retrieves the AES key from the environment variable.
// The key must be a 32-byte base64-encoded string for AES-256.
func GetEncryptionKey() ([]byte, error) {
	keyB64 := os.Getenv("MESSAGE_ENCRYPTION_KEY")
	if keyB64 == "" {
		return nil, errors.New("MESSAGE_ENCRYPTION_KEY not set")
	}
	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return nil, errors.New("MESSAGE_ENCRYPTION_KEY must be base64 encoded")
	}
	if len(key) != 32 {
		return nil, errors.New("MESSAGE_ENCRYPTION_KEY must be 32 bytes (AES-256)")
	}
	return key, nil
}

// EncryptMessage encrypts the plaintext message using AES-GCM.
// Returns base64-encoded ciphertext.
func EncryptMessage(plaintext string) (string, error) {
	key, err := GetEncryptionKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptMessage decrypts a base64-encoded ciphertext using AES-GCM.
// Returns the plaintext message.
func DecryptMessage(ciphertextB64 string) (string, error) {
	key, err := GetEncryptionKey()
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
