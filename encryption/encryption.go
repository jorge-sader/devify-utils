// Package encryption provides utilities for AES-GCM encryption and decryption.
//
// This package offers a simple interface for encrypting and decrypting text using the AES-GCM algorithm.
// It supports 128-bit, 192-bit, and 256-bit keys and uses base64 encoding for ciphertext representation.
// All functions are designed to be secure and easy to use, with proper error handling for invalid inputs
// and cryptographic operations.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Encryption is a type used to manage AES-GCM encryption and decryption operations.
//
// It holds the encryption key and provides methods for encrypting and decrypting data.
// The key must be 16, 24, or 32 bytes long to support AES-128, AES-192, or AES-256, respectively.
type Encryption struct {
	Key []byte
}

// NewEncryption creates a new Encryption instance with key validation.
//
// The key must be 16, 24, or 32 bytes long, corresponding to AES-128, AES-192, or AES-256.
// If the key length is invalid, an error is returned.
//
// Example:
//
//	key := []byte("16-byte-key12345") // 16 bytes for AES-128
//	enc, err := NewEncryption(key)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - key: The encryption key as a byte slice (must be 16, 24, or 32 bytes).
//
// Returns:
//   - *Encryption: A pointer to the initialized Encryption instance.
//   - error: An error if the key length is invalid.
func NewEncryption(key []byte) (*Encryption, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, errors.New("invalid key size: must be 16, 24, or 32 bytes")
	}
	return &Encryption{Key: key}, nil
}

// Encrypt encrypts the given plaintext using AES-GCM and returns the ciphertext as a base64-encoded string.
//
// The plaintext is encrypted using the AES-GCM algorithm, which provides both confidentiality and authenticity.
// A random nonce is generated for each encryption operation, and the resulting ciphertext includes the nonce.
// The output is base64-URL-encoded for safe storage and transmission.
//
// Example:
//
//	enc, _ := NewEncryption([]byte("16-byte-key12345"))
//	ciphertext, err := enc.Encrypt("Hello, World!")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(ciphertext) // Prints base64-encoded ciphertext
//
// Parameters:
//   - text: The plaintext string to encrypt.
//
// Returns:
//   - string: The base64-URL-encoded ciphertext (includes nonce).
//   - error: An error if the encryption process fails (e.g., invalid key or nonce generation failure).
func (e *Encryption) Encrypt(text string) (string, error) {
	plainText := []byte(text)
	block, err := aes.NewCipher(e.Key)
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
	cipherText := gcm.Seal(nonce, nonce, plainText, nil)
	return base64.URLEncoding.EncodeToString(cipherText), nil
}

// Decrypt decrypts a base64-encoded ciphertext using AES-GCM and returns the plaintext.
//
// The input ciphertext must be a base64-URL-encoded string produced by the Encrypt method.
// The function extracts the nonce from the ciphertext and uses it to decrypt the data with AES-GCM.
// If the ciphertext is invalid, too short, or decryption fails (e.g., due to tampering or incorrect key),
// an error is returned.
//
// Example:
//
//	enc, _ := NewEncryption([]byte("16-byte-key12345"))
//	ciphertext, _ := enc.Encrypt("Hello, World!")
//	plaintext, err := enc.Decrypt(ciphertext)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(plaintext) // Prints "Hello, World!"
//
// Parameters:
//   - cipherText: The base64-URL-encoded ciphertext to decrypt.
//
// Returns:
//   - string: The decrypted plaintext string.
//   - error: An error if the ciphertext is invalid, too short, or decryption fails.
func (e *Encryption) Decrypt(cipherText string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce, ct := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plainText, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(plainText), nil
}
