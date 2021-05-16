package crypto

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"fmt"
)

type Token interface {
	CreateToken(username string) (string, error)
	GetUsername(token string) (string, error)
}
type MumboJumbo struct {
	Key []byte
}

const padd = "cryptomagic=aes:awesomeusername="

func (mjtoken MumboJumbo) CreateToken(username string) (string, error) {
	readyToken, err := Encrypt([]byte(padd+username), mjtoken.Key)
	if err != nil {
		return "", err
	}
	return string(readyToken), nil
}

func (mjtoken MumboJumbo) GetUsername(token string) (string, error) {
	encryptedToken, err := Decrypt([]byte(token), mjtoken.Key)
	if err != nil {
		return "", err
	}
	username := string(encryptedToken)[len(padd):]
	return username, err
}

func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create new cipher: %s", err)
	}
	plaintext = AddPadding(plaintext)
	ciphertext := make([]byte, len(plaintext))
	if len(plaintext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("need a multiple of the blocksize for encrypt")
	}
	for bs, be := 0, aes.BlockSize; bs < len(plaintext); bs, be = bs+aes.BlockSize, be+aes.BlockSize {
		cipher.Encrypt(ciphertext[bs:be], plaintext[bs:be])
	}

	return []byte(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(string(ciphertext))
	if err != nil {
		return nil, fmt.Errorf("can't decode ciphertext: %v", err)
	}
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create new cipher: %s", err)
	}
	plaintext := make([]byte, len(ciphertext))
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("need a multiple of the blocksize for decrypt")
	}
	for bs, be := 0, aes.BlockSize; bs < len(ciphertext); bs, be = bs+aes.BlockSize, be+aes.BlockSize {
		cipher.Decrypt(plaintext[bs:be], ciphertext[bs:be])
	}
	return RemovePadding(plaintext)
}
func AddPadding(raw []byte) []byte {
	padLen := aes.BlockSize - (len(raw) % aes.BlockSize)
	var padding []byte
	padding = append(padding, byte(padLen))
	padding = bytes.Repeat(padding, padLen)
	raw = append(raw, padding...)
	return raw
}

func RemovePadding(raw []byte) ([]byte, error) {
	var rawLen = len(raw)
	if rawLen%aes.BlockSize != 0 {
		return nil, fmt.Errorf("data's length isn't a multiple of blockSize")
	}
	padBlock := raw[rawLen-aes.BlockSize:]
	if ok, padLen := ValidatePadding(padBlock); ok {
		return raw[:rawLen-padLen], nil
	} else {
		return nil, fmt.Errorf("incorrect padding in last block")
	}
}

func ValidatePadding(block []byte) (bool, int) {
	padCharacter := block[len(block)-1]
	padSize := int(padCharacter)
	if padSize > aes.BlockSize || padSize == 0 {
		return false, 0
	}
	for i := len(block) - padSize; i < len(block); i++ {
		if block[i] != padCharacter {
			return false, 0
		}
	}
	return true, padSize
}
