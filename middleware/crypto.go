package middleware

import "apollo-backend/util"

// DecryptAES256 decrypts a base64-encoded AES-256-GCM ciphertext.
// This is a convenience re-export from the util package.
func DecryptAES256(encryptedBase64 string, key []byte) (string, error) {
	return util.DecryptAES256(encryptedBase64, key)
}
