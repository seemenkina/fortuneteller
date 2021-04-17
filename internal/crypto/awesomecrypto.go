package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
)

type AwesomeCrypto interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

type PublicKey struct {
	Exp *big.Int
	Mod *big.Int
}

type PrivateKey struct {
	exp *big.Int
	mod *big.Int
}

type IzzyWizzy struct {
	PublicKey  PublicKey
	privateKey PrivateKey
}

const bits = 200

func GenerateKeyPair() IzzyWizzy {
	e := big.NewInt(3)
	p, _ := rand.Prime(rand.Reader, bits)
	q, _ := rand.Prime(rand.Reader, bits)
	N := new(big.Int).Mul(p, q)
	phi := new(big.Int).Mul(new(big.Int).Sub(p, big.NewInt(1)), new(big.Int).Sub(q, big.NewInt(1)))
	gcd := new(big.Int).GCD(nil, nil, phi, e)
	for gcd.Cmp(big.NewInt(1)) != 0 {
		p, _ = rand.Prime(rand.Reader, bits)
		q, _ = rand.Prime(rand.Reader, bits)
		N = new(big.Int).Mul(p, q)
		phi = new(big.Int).Mul(new(big.Int).Sub(p, big.NewInt(1)), new(big.Int).Sub(q, big.NewInt(1)))
		gcd = new(big.Int).GCD(nil, nil, phi, e)
	}

	d := new(big.Int).ModInverse(e, phi)
	return IzzyWizzy{
		PublicKey: PublicKey{
			Exp: e,
			Mod: N,
		},
		privateKey: PrivateKey{
			exp: d,
			mod: N,
		},
	}
}

func (iwcrypto IzzyWizzy) Encrypt(plaintext []byte) ([]byte, error) {
	bytePT := new(big.Int).SetBytes(plaintext)
	encrypted := new(big.Int).Exp(bytePT, iwcrypto.PublicKey.Exp, iwcrypto.PublicKey.Mod)

	return []byte(hex.EncodeToString(encrypted.Bytes())), nil
}

func (iwcrypto IzzyWizzy) Decrypt(ciphertext []byte) ([]byte, error) {
	ciphertext, err := hex.DecodeString(string(ciphertext))
	if err != nil {
		return nil, fmt.Errorf("can't decode ciphertext: %v", err)
	}
	byteCT := new(big.Int).SetBytes(ciphertext)
	decrypted := new(big.Int).Exp(byteCT, iwcrypto.privateKey.exp, iwcrypto.PublicKey.Mod)

	return decrypted.Bytes(), nil
}
