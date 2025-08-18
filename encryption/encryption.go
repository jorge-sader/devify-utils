package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Encryption is the type used to instantiate this module.
type Encryption struct {
	Key []byte
}

// NewEncryption creates a new Encryption instance with key validation.
func NewEncryption(key []byte) (*Encryption, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, errors.New("invalid key size: must be 16, 24, or 32 bytes")
	}
	return &Encryption{Key: key}, nil
}

// Encrypt encrypts the given plaintext using AES-GCM.
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

// Decrypt decrypts the given base64-encoded ciphertext using AES-GCM.
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
