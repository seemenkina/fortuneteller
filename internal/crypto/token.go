package crypto

import "encoding/hex"

type Token interface {
	CreateToken(username string) (string, error)
	GetUsername(token string) (string, error)
}
type MumboJumbo struct {
}

func (mjtoken MumboJumbo) CreateToken(username string) (string, error) {
	b := hex.EncodeToString([]byte(username))
	return b, nil
}

func (mjtoken MumboJumbo) GetUsername(token string) (string, error) {
	b, err := hex.DecodeString(token)
	return string(b), err
}
