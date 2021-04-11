package crypto

import (
	"bytes"
	"crypto/aes"
	"fmt"
)

type AwesomeCrypto interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

type IzzyWizzy struct {
	Key []byte
}

func (iwcrypto IzzyWizzy) Encrypt(plaintext []byte) ([]byte, error) {
	cipher, err := aes.NewCipher(iwcrypto.Key)
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
	return ciphertext, nil
}

func (iwcrypto IzzyWizzy) Decrypt(ciphertext []byte) ([]byte, error) {
	cipher, err := aes.NewCipher(iwcrypto.Key)
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
