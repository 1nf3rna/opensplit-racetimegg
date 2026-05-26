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
	"log"
	"os"
	"time"

	"golang.org/x/oauth2"
)

func logSecureStore(format string, v ...any) {
	log.Printf("[SECURESTORE] "+format, v...)
}

func encrypt(data []byte, key []byte) (string, error) {
	logSecureStore("encrypting token payload (%d bytes)", len(data))

	block, err := aes.NewCipher(key)
	if err != nil {
		logSecureStore("aes.NewCipher failed: %v", err)
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logSecureStore("cipher.NewGCM failed: %v", err)
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		logSecureStore("nonce generation failed: %v", err)
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	logSecureStore("encryption successful")

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(enc string, key []byte) ([]byte, error) {
	logSecureStore("decrypting token payload")

	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		logSecureStore("base64 decode failed: %v", err)
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		logSecureStore("aes.NewCipher failed: %v", err)
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logSecureStore("cipher.NewGCM failed: %v", err)
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		err := errors.New("ciphertext too short")
		logSecureStore("decrypt failed: %v", err)
		return nil, err
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		logSecureStore("gcm.Open failed: %v", err)
		return nil, err
	}

	logSecureStore("decryption successful")

	return plain, nil
}

func SaveToken(path string, token oauth2.Token, key []byte) error {
	logSecureStore(
		"saving token file path=%s expiry=%s refresh_present=%v",
		path,
		token.Expiry.Format(time.RFC3339),
		token.RefreshToken != "",
	)

	data, err := json.Marshal(token)
	if err != nil {
		logSecureStore("json.Marshal failed: %v", err)
		return err
	}

	enc, err := encrypt(data, key)
	if err != nil {
		logSecureStore("encrypt failed: %v", err)
		return err
	}

	err = os.WriteFile(path, []byte(enc), 0600)
	if err != nil {
		logSecureStore("WriteFile failed: %v", err)
		return err
	}

	logSecureStore("token saved successfully")

	return nil
}

func LoadToken(path string, key []byte) (*oauth2.Token, error) {
	logSecureStore("loading token file path=%s", path)

	enc, err := os.ReadFile(path)
	if err != nil {
		logSecureStore("ReadFile failed: %v", err)
		return nil, err
	}

	data, err := decrypt(string(enc), key)
	if err != nil {
		logSecureStore("decrypt failed: %v", err)
		return nil, err
	}

	var token oauth2.Token

	if err := json.Unmarshal(data, &token); err != nil {
		logSecureStore("json.Unmarshal failed: %v", err)
		return nil, err
	}

	logSecureStore(
		"token loaded expiry=%s valid=%v refresh_present=%v",
		token.Expiry.Format(time.RFC3339),
		token.Valid(),
		token.RefreshToken != "",
	)

	return &token, nil
}

func DeleteToken(path string) error {
	logSecureStore("deleting token file path=%s", path)

	err := os.Remove(path)
	if err != nil {
		logSecureStore("token delete failed: %v", err)
		return err
	}

	logSecureStore("token deleted successfully")

	return nil
}

func KeyFromEnv(secret string) []byte {
	logSecureStore("deriving encryption key from secret")

	hash := sha256.Sum256([]byte(secret))

	logSecureStore("key derivation complete")

	return hash[:]
}
