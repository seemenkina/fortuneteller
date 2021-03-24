package repository

import (
	"errors"

	"fortuneteller/internal/models"
)

var ErrNoSuchUser = errors.New("no such user")

type User interface {
	AddUser(u models.User) error
	AllUsers() ([]models.User, error)
	FindUserByName(name string) (models.User, error)
	FindUserByToken(token string) (models.User, error)
}
