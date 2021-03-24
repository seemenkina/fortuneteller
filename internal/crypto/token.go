package crypto

type Token interface {
	CreateToken(username string) (string, error)
}
