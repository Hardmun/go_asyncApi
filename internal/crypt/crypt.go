package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
)

// generateKey generates a 32-byte key from a secret word using SHA-256.
func generateKey(secret string) []byte {
	hash := sha256.Sum256([]byte(secret))
	return hash[:]
}

// EncryptToken encrypts a token using AES-GCM with the provided key.
func EncryptToken(token string, secret string) (string, error) {
	var aesGCM cipher.AEAD

	key := generateKey(secret)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 12) // 12 bytes is standard for GCM nonce
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	aesGCM, err = cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nil, nonce, []byte(token), nil)
	fullCiphertext := append(nonce, ciphertext...)

	return base64.StdEncoding.EncodeToString(fullCiphertext), nil
}

// DecryptToken decrypts an encrypted token using AES-GCM with the provided key.
func DecryptToken(encryptedToken string, secret string) (string, error) {
	var (
		aesGCM    cipher.AEAD
		block     cipher.Block
		plaintext []byte
	)

	key := generateKey(secret)

	data, err := base64.StdEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", err
	}

	block, err = aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	nonce := data[:12] // Extract the nonce
	ciphertext := data[12:]

	aesGCM, err = cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err = aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
