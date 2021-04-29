package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
)

type AwesomeCrypto interface {
	Encrypt(plaintext []byte) []byte
	Decrypt(ciphertext []byte) ([]byte, error)
}

type PublicKey struct {
	Exp *big.Int
	Mod *big.Int
}

type PrivateKey struct {
	Exp *big.Int
	Mod *big.Int
}

type IzzyWizzy struct {
	PublicKey  PublicKey
	PrivateKey PrivateKey
}

const bits = 1024

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
		PrivateKey: PrivateKey{
			Exp: d,
			Mod: N,
		},
	}
}

func (iwcrypto IzzyWizzy) Encrypt(plaintext []byte) []byte {
	if iwcrypto.PublicKey.Exp == nil {
		log.Printf("RETURN PLAINTEXT: %s", plaintext)
		return plaintext
	}

	bytePT := new(big.Int).SetBytes(plaintext)
	encrypted := new(big.Int).Exp(bytePT, iwcrypto.PublicKey.Exp, iwcrypto.PublicKey.Mod)

	return []byte(base64.StdEncoding.EncodeToString(encrypted.Bytes()))
}

func (iwcrypto IzzyWizzy) Decrypt(ciphertext []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(string(ciphertext))
	if err != nil {
		return nil, fmt.Errorf("can't decode ciphertext: %v", err)
	}

	if iwcrypto.PrivateKey.Exp == nil {
		log.Printf("RETURN CIPHERTEXT: %s", ciphertext)
		return ciphertext, nil
	}
	byteCT := new(big.Int).SetBytes(ciphertext)
	decrypted := new(big.Int).Exp(byteCT, iwcrypto.PrivateKey.Exp, iwcrypto.PublicKey.Mod)

	return decrypted.Bytes(), nil
}

func (iwcrypto IzzyWizzy) SaveKeyOnFile(filename string) error {
	var f *os.File
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		f, err = os.Create(filename)

		if err != nil {
			return err
		}
		log.Printf("CREATE FILE: %s", filename)
	} else {
		f, err = os.Open(filename)
		if err != nil {
			return err
		}
		log.Printf("OPEN FILE: %s", filename)
	}
	defer func() {
		_ = f.Close()
	}()

	js, err := json.Marshal(iwcrypto)
	if err != nil {
		return err
	}

	if _, err = f.Write(js); err != nil {
		return err
	}
	return nil
}

func LoadKeyFromFile(filename string) IzzyWizzy {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("ERROR LOAD: %v", err)
		return IzzyWizzy{}
	}

	key := IzzyWizzy{}
	if err := json.Unmarshal(f, &key); err != nil {
		log.Printf("ERROR LOAD: %v", err)
		return IzzyWizzy{}
	}

	return key
}
