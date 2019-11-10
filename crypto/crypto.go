package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
)

func Encrypt(key, plain []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ret := gcm.Seal(nil, nonce, plain, nil)
	ret = append(nonce, ret...)
	return ret, nil
}

func Decrypt(key, cipherText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to new gcm: %w", err)
	}
	if len(cipherText) < gcm.NonceSize() {
		return nil, errors.New("crypto: invalid cipher text size")
	}
	nonce := cipherText[:gcm.NonceSize()]
	ret, err := gcm.Open(nil, nonce, cipherText[gcm.NonceSize():], nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open gcm: %w", err)
	}
	return ret, nil
}
