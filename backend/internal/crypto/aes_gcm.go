// Package crypto provides AES-256-GCM encryption and decryption utilities
// for protecting sensitive patient data (CIN, Full Name, Symptoms) at rest.
//
// Security Design:
//   - 256-bit key from hex-encoded environment variable
//   - Random 12-byte nonce per encryption (never reused)
//   - GCM provides authenticated encryption (confidentiality + integrity)
//   - Output format: hex(nonce) + ":" + hex(ciphertext+tag)
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

// AESCrypto holds the parsed AES key and provides Encrypt/Decrypt methods.
type AESCrypto struct {
	key []byte
}

var (
	ErrInvalidKeyLength      = errors.New("crypto: AES key must be exactly 64 hex characters (256 bits)")
	ErrInvalidKeyHex         = errors.New("crypto: AES key contains invalid hex characters")
	ErrInvalidCiphertextFmt  = errors.New("crypto: ciphertext must be in format 'nonce:ciphertext' (hex-encoded)")
	ErrCiphertextTooShort    = errors.New("crypto: ciphertext is too short to contain a valid GCM tag")
	ErrDecryptionFailed      = errors.New("crypto: decryption failed — data may be tampered or key is wrong")
)

// NewAESCrypto creates a new AESCrypto instance from a 64-character hex-encoded key.
// Returns an error if the key is not exactly 256 bits or contains invalid hex.
func NewAESCrypto(hexKey string) (*AESCrypto, error) {
	hexKey = strings.TrimSpace(hexKey)

	if len(hexKey) != 64 {
		return nil, fmt.Errorf("%w: got %d characters", ErrInvalidKeyLength, len(hexKey))
	}

	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidKeyHex, err)
	}

	return &AESCrypto{key: key}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with a random nonce.
// Returns the result as "hex(nonce):hex(ciphertext+tag)".
// Each call produces a different output due to the random nonce,
// even for identical plaintext inputs.
func (a *AESCrypto) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", fmt.Errorf("crypto: failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: failed to create GCM: %w", err)
	}

	// Generate a random 12-byte nonce (standard GCM nonce size)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("crypto: failed to generate nonce: %w", err)
	}

	// Seal encrypts and authenticates the plaintext
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// Encode as hex: "nonce:ciphertext"
	encoded := hex.EncodeToString(nonce) + ":" + hex.EncodeToString(ciphertext)
	return encoded, nil
}

// Decrypt decrypts ciphertext that was produced by Encrypt.
// Expects input in the format "hex(nonce):hex(ciphertext+tag)".
// Returns the original plaintext or an error if authentication fails.
func (a *AESCrypto) Decrypt(encoded string) (string, error) {
	parts := strings.SplitN(encoded, ":", 2)
	if len(parts) != 2 {
		return "", ErrInvalidCiphertextFmt
	}

	nonce, err := hex.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("%w: invalid nonce hex: %v", ErrInvalidCiphertextFmt, err)
	}

	ciphertext, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("%w: invalid ciphertext hex: %v", ErrInvalidCiphertextFmt, err)
	}

	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", fmt.Errorf("crypto: failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: failed to create GCM: %w", err)
	}

	if len(nonce) != gcm.NonceSize() {
		return "", fmt.Errorf("%w: nonce length %d, expected %d", ErrInvalidCiphertextFmt, len(nonce), gcm.NonceSize())
	}

	if len(ciphertext) < gcm.Overhead() {
		return "", ErrCiphertextTooShort
	}

	// Open decrypts and verifies the authentication tag
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// MustNewAESCrypto is like NewAESCrypto but panics on error.
// Use this only during application startup where a missing key is fatal.
func MustNewAESCrypto(hexKey string) *AESCrypto {
	c, err := NewAESCrypto(hexKey)
	if err != nil {
		panic(fmt.Sprintf("FATAL: Cannot initialize encryption: %v", err))
	}
	return c
}
