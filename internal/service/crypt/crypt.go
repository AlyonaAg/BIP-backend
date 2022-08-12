package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
)

func GenerateRandKey(length int) ([]byte, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func EncryptAES(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	plainText := []byte(plaintext)
	var iv [aes.BlockSize]byte
	cfb := cipher.NewCFBEncrypter(block, iv[:])
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return Encode(cipherText), nil
}

func Decode(s string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func DecryptAES(key []byte, ct string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	cipherText, err := Decode(ct)
	if err != nil {
		return "", err
	}

	var iv [aes.BlockSize]byte
	cfb := cipher.NewCFBDecrypter(block, iv[:])

	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}
