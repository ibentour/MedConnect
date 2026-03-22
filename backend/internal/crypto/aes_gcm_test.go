package crypto

import (
	"encoding/hex"
	"strings"
	"testing"
)

// testKey is a valid 256-bit key (64 hex chars) used across tests.
const testKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// ─── Test: Roundtrip Encrypt → Decrypt ──────────────────────────────

func TestEncryptDecrypt(t *testing.T) {
	crypto, err := NewAESCrypto(testKey)
	if err != nil {
		t.Fatalf("NewAESCrypto failed: %v", err)
	}

	testCases := []string{
		"Patient: Mohammed El Amrani",
		"CIN: AB123456",
		"Symptoms: Severe chest pain radiating to left arm, shortness of breath",
		"مريض يعاني من صداع شديد", // Arabic text
		"Douleur abdominale aiguë",  // French with accents
		"A",                         // Single character
		strings.Repeat("X", 10000), // Large payload
	}

	for _, plaintext := range testCases {
		t.Run(plaintext[:min(len(plaintext), 30)], func(t *testing.T) {
			encrypted, err := crypto.Encrypt(plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// Verify format: should contain exactly one ":"
			parts := strings.SplitN(encrypted, ":", 2)
			if len(parts) != 2 {
				t.Fatalf("Encrypted output missing ':' separator: %s", encrypted)
			}

			decrypted, err := crypto.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if decrypted != plaintext {
				t.Errorf("Roundtrip mismatch:\n  got:  %q\n  want: %q", decrypted, plaintext)
			}
		})
	}
}

// ─── Test: Empty Plaintext ──────────────────────────────────────────

func TestEmptyPlaintext(t *testing.T) {
	crypto, err := NewAESCrypto(testKey)
	if err != nil {
		t.Fatalf("NewAESCrypto failed: %v", err)
	}

	encrypted, err := crypto.Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt of empty string failed: %v", err)
	}

	decrypted, err := crypto.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt of empty string failed: %v", err)
	}

	if decrypted != "" {
		t.Errorf("Expected empty string, got: %q", decrypted)
	}
}

// ─── Test: Nonce Uniqueness ─────────────────────────────────────────

func TestNonceUniqueness(t *testing.T) {
	crypto, err := NewAESCrypto(testKey)
	if err != nil {
		t.Fatalf("NewAESCrypto failed: %v", err)
	}

	plaintext := "Same input, different output"
	seen := make(map[string]bool)

	for i := 0; i < 100; i++ {
		encrypted, err := crypto.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("Encrypt iteration %d failed: %v", i, err)
		}

		if seen[encrypted] {
			t.Fatalf("Duplicate ciphertext detected at iteration %d — nonce reuse!", i)
		}
		seen[encrypted] = true
	}
}

// ─── Test: Invalid Key Lengths ──────────────────────────────────────

func TestInvalidKeyLength(t *testing.T) {
	invalidKeys := []struct {
		name string
		key  string
	}{
		{"too short (32 chars)", "0123456789abcdef0123456789abcdef"},
		{"too long (128 chars)", strings.Repeat("ab", 64)},
		{"empty", ""},
		{"128-bit key (32 chars)", "0123456789abcdef0123456789abcdef"},
	}

	for _, tc := range invalidKeys {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewAESCrypto(tc.key)
			if err == nil {
				t.Error("Expected error for invalid key length, got nil")
			}
		})
	}
}

// ─── Test: Invalid Hex Key ──────────────────────────────────────────

func TestInvalidKeyHex(t *testing.T) {
	// 64 chars but contains invalid hex characters
	badHexKey := "ZZZZ456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	_, err := NewAESCrypto(badHexKey)
	if err == nil {
		t.Error("Expected error for invalid hex key, got nil")
	}
}

// ─── Test: Tampered Ciphertext ──────────────────────────────────────

func TestTamperedCiphertext(t *testing.T) {
	crypto, err := NewAESCrypto(testKey)
	if err != nil {
		t.Fatalf("NewAESCrypto failed: %v", err)
	}

	encrypted, err := crypto.Encrypt("Sensitive patient data")
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Tamper with the ciphertext portion (after the ":")
	parts := strings.SplitN(encrypted, ":", 2)

	// Flip some bytes in the ciphertext
	ciphertextBytes, _ := hex.DecodeString(parts[1])
	if len(ciphertextBytes) > 0 {
		ciphertextBytes[0] ^= 0xFF // Flip all bits of first byte
	}
	tampered := parts[0] + ":" + hex.EncodeToString(ciphertextBytes)

	_, err = crypto.Decrypt(tampered)
	if err == nil {
		t.Error("Expected decryption error for tampered ciphertext, got nil")
	}
}

// ─── Test: Malformed Ciphertext Formats ─────────────────────────────

func TestMalformedCiphertext(t *testing.T) {
	crypto, err := NewAESCrypto(testKey)
	if err != nil {
		t.Fatalf("NewAESCrypto failed: %v", err)
	}

	malformed := []struct {
		name  string
		input string
	}{
		{"no separator", "abcdef1234567890"},
		{"empty string", ""},
		{"only separator", ":"},
		{"invalid hex nonce", "ZZZZ:abcdef"},
		{"invalid hex ciphertext", "abcdef1234567890abcdef12:ZZZZ"},
	}

	for _, tc := range malformed {
		t.Run(tc.name, func(t *testing.T) {
			_, err := crypto.Decrypt(tc.input)
			if err == nil {
				t.Error("Expected error for malformed ciphertext, got nil")
			}
		})
	}
}

// ─── Test: Different Keys Cannot Decrypt ────────────────────────────

func TestWrongKeyCannotDecrypt(t *testing.T) {
	key1 := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	key2 := "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"

	crypto1, _ := NewAESCrypto(key1)
	crypto2, _ := NewAESCrypto(key2)

	encrypted, err := crypto1.Encrypt("Top secret patient data")
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = crypto2.Decrypt(encrypted)
	if err == nil {
		t.Error("Expected error when decrypting with wrong key, got nil")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
