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
	"opensplit-racetimegg/logger"
	"os"
	"time"

	"golang.org/x/oauth2"
)

var log = logger.Module("securestore/tokenstorage").SetLevel(logger.ErrorLevel)

func encrypt(data []byte, key []byte) (string, error) {
	log.Debug("encrypting token payload (%d bytes)", len(data))

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Error("aes.NewCipher failed: %v", err)
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Error("cipher.NewGCM failed: %v", err)
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Error("nonce generation failed: %v", err)
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	log.Debug("encryption successful")

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(enc string, key []byte) ([]byte, error) {
	log.Debug("decrypting token payload")

	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		log.Error("base64 decode failed: %v", err)
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Error("aes.NewCipher failed: %v", err)
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Error("cipher.NewGCM failed: %v", err)
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		err := errors.New("ciphertext too short")
		log.Error("decrypt failed: %v", err)
		return nil, err
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Error("gcm.Open failed: %v", err)
		return nil, err
	}

	log.Debug("decryption successful")

	return plain, nil
}

func SaveToken(path string, token oauth2.Token, key []byte) error {
	log.Debug(
		"saving token file path=%s expiry=%s refresh_present=%v",
		path,
		token.Expiry.Format(time.RFC3339),
		token.RefreshToken != "",
	)

	data, err := json.Marshal(token)
	if err != nil {
		log.Error("json.Marshal failed: %v", err)
		return err
	}

	enc, err := encrypt(data, key)
	if err != nil {
		log.Error("encrypt failed: %v", err)
		return err
	}

	err = os.WriteFile(path, []byte(enc), 0600)
	if err != nil {
		log.Error("WriteFile failed: %v", err)
		return err
	}

	log.Info("token saved successfully")

	return nil
}

func LoadToken(path string, key []byte) (*oauth2.Token, error) {
	log.Info("loading token file path=%s", path)

	enc, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Info("token file does not exist")
		} else {
			log.Error("ReadFile failed: %v", err)
		}

		return nil, err
	}

	data, err := decrypt(string(enc), key)
	if err != nil {
		log.Error("decrypt failed: %v", err)
		return nil, err
	}

	var token oauth2.Token

	if err := json.Unmarshal(data, &token); err != nil {
		log.Error("json.Unmarshal failed: %v", err)
		return nil, err
	}

	log.Debug(
		"token loaded expiry=%s valid=%v refresh_present=%v",
		token.Expiry.Format(time.RFC3339),
		token.Valid(),
		token.RefreshToken != "",
	)

	return &token, nil
}

func DeleteToken(path string) error {
	log.Info("deleting token file path=%s", path)

	err := os.Remove(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Debug("token file already deleted")
			return nil
		}

		log.Error("token delete failed: %v", err)
		return err
	}

	log.Info("token deleted successfully")

	return nil
}

func KeyFromEnv(secret string) []byte {
	log.Debug("deriving encryption key from secret")

	hash := sha256.Sum256([]byte(secret))

	log.Debug("key derivation complete")

	return hash[:]
}
