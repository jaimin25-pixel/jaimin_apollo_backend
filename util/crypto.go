package util

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

// DecryptAES256 decrypts a base64-encoded AES-256-GCM ciphertext.
// Format: base64(nonce + ciphertext)
// Key must be 32 bytes for AES-256
func DecryptAES256(encryptedBase64 string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", errors.New("invalid base64 encoding")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.New("invalid encryption key")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("decryption failed")
	}
	return string(plaintext), nil
}
