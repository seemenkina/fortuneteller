package mocks

import (
	"encoding/hex"
)

type TokenMock struct {
}

func (t TokenMock) CreateToken(username string) (string, error) {
	b := hex.EncodeToString([]byte(username))
	return b, nil
}

type AwesomeCryptoMock struct {
}

func (acr AwesomeCryptoMock) Encrypt(plaintext []byte) ([]byte, error) {
	return []byte(hex.EncodeToString(plaintext)), nil
}

func (acr AwesomeCryptoMock) Decrypt(ciphertext []byte) ([]byte, error) {
	return hex.DecodeString(string(ciphertext))
}
