package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

// Encrypt encrypts plaintext using AES-256-GCM with the given hex-encoded key.
// Returns hex-encoded ciphertext (nonce + encrypted data).
func Encrypt(key, plaintext string) (string, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return "", errors.New("invalid encryption key: must be hex-encoded")
	}
	if len(keyBytes) != 32 {
		return "", errors.New("invalid encryption key: must be 32 bytes (64 hex chars)")
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt decrypts hex-encoded ciphertext using AES-256-GCM.
func Decrypt(key, ciphertextHex string) (string, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return "", errors.New("invalid encryption key")
	}
	if len(keyBytes) != 32 {
		return "", errors.New("invalid encryption key length")
	}

	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", errors.New("invalid ciphertext")
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("decryption failed: invalid key or corrupted data")
	}

	return string(plaintext), nil
}

// GenerateKey generates a random 32-byte AES-256 key as hex string.
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}
