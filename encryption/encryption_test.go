package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"
)

// TestNewEncryption tests the NewEncryption constructor for valid and invalid key sizes.
func TestNewEncryption(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		wantErr bool
	}{
		{
			name:    "Valid AES-128 key (16 bytes)",
			key:     make([]byte, 16),
			wantErr: false,
		},
		{
			name:    "Valid AES-192 key (24 bytes)",
			key:     make([]byte, 24),
			wantErr: false,
		},
		{
			name:    "Valid AES-256 key (32 bytes)",
			key:     make([]byte, 32),
			wantErr: false,
		},
		{
			name:    "Invalid key size (15 bytes)",
			key:     make([]byte, 15),
			wantErr: true,
		},
		{
			name:    "Invalid key size (0 bytes)",
			key:     []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEncryption(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEncryption() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestEncryptDecrypt tests the Encrypt and Decrypt methods for round-trip correctness and error cases.
func TestEncryptDecrypt(t *testing.T) {
	// Generate a valid 32-byte key for AES-256
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal("Failed to generate key:", err)
	}

	enc, err := NewEncryption(key)
	if err != nil {
		t.Fatal("Failed to create Encryption:", err)
	}

	tests := []struct {
		name      string
		plainText string
		wantErr   bool
	}{
		{
			name:      "Normal text",
			plainText: "Hello, world!",
			wantErr:   false,
		},
		{
			name:      "Empty validate",
			plainText: "",
			wantErr:   false,
		},
		{
			name:      "Long text",
			plainText: strings.Repeat("a", 1000),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test encryption
			cipherText, err := enc.Encrypt(tt.plainText)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Test decryption
			decrypted, err := enc.Decrypt(cipherText)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if decrypted != tt.plainText {
				t.Errorf("Decrypt() = %q, want %q", decrypted, tt.plainText)
			}
		})
	}
}

// TestDecryptErrorCases tests Decrypt with invalid inputs.
func TestDecryptErrorCases(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal("Failed to generate key:", err)
	}

	enc, err := NewEncryption(key)
	if err != nil {
		t.Fatal("Failed to create Encryption:", err)
	}

	tests := []struct {
		name       string
		cipherText string
		wantErr    bool
	}{
		{
			name:       "Invalid base64",
			cipherText: "invalid-base64-!",
			wantErr:    true,
		},
		{
			name:       "Ciphertext too short",
			cipherText: base64.URLEncoding.EncodeToString([]byte("short")),
			wantErr:    true,
		},
		{
			name:       "Tampered ciphertext",
			cipherText: base64.URLEncoding.EncodeToString(append(make([]byte, 12), []byte("tampered")...)),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := enc.Decrypt(tt.cipherText)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestInvalidKeyEncryption tests encryption with an invalid key.
func TestInvalidKeyEncryption(t *testing.T) {
	enc := &Encryption{Key: []byte("invalid")} // Invalid key size
	_, err := enc.Encrypt("test")
	if err == nil {
		t.Error("Encrypt() should fail with invalid key size")
	}
}

// TestInvalidKeyDecryption tests decryption with an invalid key.
func TestInvalidKeyDecryption(t *testing.T) {
	enc := &Encryption{Key: []byte("invalid")} // Invalid key size
	_, err := enc.Decrypt(base64.URLEncoding.EncodeToString([]byte("test")))
	if err == nil {
		t.Error("Decrypt() should fail with invalid key size")
	}
}

// TestDifferentKeyDecrypt tests decryption with a different key than encryption.
func TestDifferentKeyDecrypt(t *testing.T) {
	key1 := make([]byte, 32)
	if _, err := rand.Read(key1); err != nil {
		t.Fatal("Failed to generate key1:", err)
	}
	key2 := make([]byte, 32)
	if _, err := rand.Read(key2); err != nil {
		t.Fatal("Failed to generate key2:", err)
	}

	enc1, err := NewEncryption(key1)
	if err != nil {
		t.Fatal("Failed to create enc1:", err)
	}
	enc2, err := NewEncryption(key2)
	if err != nil {
		t.Fatal("Failed to create enc2:", err)
	}

	cipherText, err := enc1.Encrypt("test")
	if err != nil {
		t.Fatal("Failed to encrypt:", err)
	}

	_, err = enc2.Decrypt(cipherText)
	if err == nil {
		t.Error("Decrypt() should fail with different key")
	}
}
