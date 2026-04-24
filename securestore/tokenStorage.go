package securestore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"

	"golang.org/x/oauth2"
)

func encrypt(data []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
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

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(enc string, key []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	return gcm.Open(nil, nonce, ciphertext, nil)
}

func SaveToken(path string, token oauth2.Token, key []byte) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	enc, err := encrypt(data, key)
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(enc), 0600)
}

func LoadToken(path string, key []byte) (*oauth2.Token, error) {
	enc, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	data, err := decrypt(string(enc), key)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

func KeyFromEnv(secret string) []byte {
	hash := sha256.Sum256([]byte(secret))
	return hash[:] // 32 bytes for AES-256
}
