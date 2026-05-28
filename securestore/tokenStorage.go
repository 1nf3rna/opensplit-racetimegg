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
	"time"

	"opensplit-racetimegg/logging"

	"golang.org/x/oauth2"
)

const component = "SECURESTORE"

var logger = logging.NewLogger(true)

func encrypt(data []byte, key []byte) (string, error) {
	logger.Debug(component, "encrypting token payload (%d bytes)", len(data))

	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Error(component, "aes.NewCipher failed: %v", err)
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Error(component, "cipher.NewGCM failed: %v", err)
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		logger.Error(component, "nonce generation failed: %v", err)
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	logger.Debug(component, "encryption successful")

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(enc string, key []byte) ([]byte, error) {
	logger.Debug(component, "decrypting token payload")

	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		logger.Error(component, "base64 decode failed: %v", err)
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Error(component, "aes.NewCipher failed: %v", err)
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Error(component, "cipher.NewGCM failed: %v", err)
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		err := errors.New("ciphertext too short")
		logger.Error(component, "decrypt failed: %v", err)
		return nil, err
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		logger.Error(component, "gcm.Open failed: %v", err)
		return nil, err
	}

	logger.Debug(component, "decryption successful")

	return plain, nil
}

func SaveToken(path string, token oauth2.Token, key []byte) error {
	logger.Info(
		component,
		"saving token file path=%s expiry=%s refresh_present=%v",
		path,
		token.Expiry.Format(time.RFC3339),
		token.RefreshToken != "",
	)

	data, err := json.Marshal(token)
	if err != nil {
		logger.Error(component, "json.Marshal failed: %v", err)
		return err
	}

	enc, err := encrypt(data, key)
	if err != nil {
		logger.Error(component, "encrypt failed: %v", err)
		return err
	}

	err = os.WriteFile(path, []byte(enc), 0600)
	if err != nil {
		logger.Error(component, "WriteFile failed: %v", err)
		return err
	}

	logger.Info(component, "token saved successfully")

	return nil
}

func LoadToken(path string, key []byte) (*oauth2.Token, error) {
	logger.Info(component, "loading token file path=%s", path)

	enc, err := os.ReadFile(path)
	if err != nil {
		logger.Error(component, "ReadFile failed: %v", err)
		return nil, err
	}

	data, err := decrypt(string(enc), key)
	if err != nil {
		logger.Error(component, "decrypt failed: %v", err)
		return nil, err
	}

	var token oauth2.Token

	if err := json.Unmarshal(data, &token); err != nil {
		logger.Error(component, "json.Unmarshal failed: %v", err)
		return nil, err
	}

	logger.Info(
		component,
		"token loaded expiry=%s valid=%v refresh_present=%v",
		token.Expiry.Format(time.RFC3339),
		token.Valid(),
		token.RefreshToken != "",
	)

	return &token, nil
}

func DeleteToken(path string) error {
	logger.Info(component, "deleting token file path=%s", path)

	err := os.Remove(path)
	if err != nil {
		logger.Error(component, "token delete failed: %v", err)
		return err
	}

	logger.Info(component, "token deleted successfully")

	return nil
}

func KeyFromEnv(secret string) []byte {
	logger.Debug(component, "deriving encryption key from secret")

	hash := sha256.Sum256([]byte(secret))

	logger.Debug(component, "key derivation complete")

	return hash[:]
}
