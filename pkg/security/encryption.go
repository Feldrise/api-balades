package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

type EncryptionService interface {
	Encrypt(plainText string) (string, error)
	Decrypt(cipherText string) (string, error)
}

type encryptionService struct {
	key []byte
}

func NewEncryptionService(secretKey string) EncryptionService {
	// Use first 32 bytes of the secret key for AES-256
	key := []byte(secretKey)
	if len(key) > 32 {
		key = key[:32]
	} else if len(key) < 32 {
		// Pad the key if it's too short
		padding := make([]byte, 32-len(key))
		key = append(key, padding...)
	}

	return &encryptionService{
		key: key,
	}
}

func (e *encryptionService) Encrypt(plainText string) (string, error) {
	if plainText == "" {
		return "", nil
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	// Create a new GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the plaintext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *encryptionService) Decrypt(cipherText string) (string, error) {
	if cipherText == "" {
		return "", nil
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	// Create a new GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Get the nonce size
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Split the nonce and ciphertext
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt the ciphertext
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
