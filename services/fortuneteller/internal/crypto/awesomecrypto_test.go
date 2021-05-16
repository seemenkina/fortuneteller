package crypto

import (
	"crypto/rand"
	"math/big"
	"testing"
)

func TestRSA(t *testing.T) {
	p, _ := rand.Prime(rand.Reader, 51)
	t.Log("p : ", p.Text(10))
	t.Log("Is prime : ", p.ProbablyPrime(1))
	q, _ := rand.Prime(rand.Reader, 51)
	t.Log("q : ", q.Text(10))
	t.Log("Is prime : ", q.ProbablyPrime(1))

	N := new(big.Int).Mul(p, q)
	t.Log("N : ", N.Text(10))

	phi := new(big.Int).Mul(new(big.Int).Sub(p, big.NewInt(1)), new(big.Int).Sub(q, big.NewInt(1)))
	t.Log("phi(N) : ", phi.Text(10))

	e, _ := rand.Int(rand.Reader, phi)
	t.Log("e : ", e.Text(10))
	gcd := new(big.Int).GCD(nil, nil, phi, e)
	t.Log("gcd : ", gcd.Text(10))
	for gcd.Cmp(big.NewInt(1)) != 0 {
		e, _ = rand.Int(rand.Reader, phi)
		// t.Log("e : ", e.Text(10))
		gcd = new(big.Int).GCD(nil, nil, phi, e)
		// t.Log("gcd : ", gcd.Text(10))
	}
	t.Log("N : ", N.Text(10))
	t.Log("phi(N) : ", phi.Text(10))
	t.Log("e : ", e.Text(10))
	d := new(big.Int).ModInverse(e, phi)
	t.Log("d : ", d.Text(10))

	m := big.NewInt(39)
	t.Log("message : ", m.Text(10))
	c := new(big.Int).Exp(m, e, N)
	t.Log("encrypted : ", c.Text(10))
	pl := new(big.Int).Exp(c, d, N)
	t.Log("decrypted : ", pl.Text(10))
}

func TestRSASmallE(t *testing.T) {

	e := big.NewInt(3)
	t.Log("e : ", e.Text(10))

	p, _ := rand.Prime(rand.Reader, 200)
	q, _ := rand.Prime(rand.Reader, 200)
	N := new(big.Int).Mul(p, q)
	phi := new(big.Int).Mul(new(big.Int).Sub(p, big.NewInt(1)), new(big.Int).Sub(q, big.NewInt(1)))
	gcd := new(big.Int).GCD(nil, nil, phi, e)
	for gcd.Cmp(big.NewInt(1)) != 0 {
		p, _ = rand.Prime(rand.Reader, 200)
		q, _ = rand.Prime(rand.Reader, 200)
		N = new(big.Int).Mul(p, q)
		phi = new(big.Int).Mul(new(big.Int).Sub(p, big.NewInt(1)), new(big.Int).Sub(q, big.NewInt(1)))
		gcd = new(big.Int).GCD(nil, nil, phi, e)
	}

	t.Log("p : ", p.Text(10))
	t.Log("Is prime : ", p.ProbablyPrime(1))
	t.Log("q : ", q.Text(10))
	t.Log("Is prime : ", q.ProbablyPrime(1))
	t.Log("N : ", N.Text(10))
	t.Log("phi(N) : ", phi.Text(10))

	d := new(big.Int).ModInverse(e, phi)
	t.Log("d : ", d.Text(10))

	m := new(big.Int).SetBytes([]byte("FY0363251JDF9IC02BPFX245C3FCD66="))
	t.Log("message : ", string(m.Bytes()))
	c := new(big.Int).Exp(m, e, N)

	t.Log("encrypted : ", c)

	pl := new(big.Int).Exp(c, d, N)

	t.Log("decrypted : ", string(pl.Bytes()))
}
