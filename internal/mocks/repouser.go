package mocks

import (
	"fmt"

	"fortuneteller/internal/models"
	"fortuneteller/internal/repository"
)

type UserMock map[string]models.User

func (u UserMock) AddUser(user models.User) error {
	_, ok := u[user.Username]
	if ok {
		return fmt.Errorf("user exists")
	}
	u[user.Username] = user
	return nil
}

func (u UserMock) FindUserByName(name string) (models.User, error) {
	user, ok := u[name]
	if ok {
		return user, nil
	} else {
		return models.User{}, repository.ErrNoSuchUser
	}
}

func (u UserMock) AllUsers() ([]models.User, error) {
	var users []models.User
	for _, user := range u {
		users = append(users, user)
	}
	return users, nil
}

func (u UserMock) FindUserByToken(token string) (models.User, error) {
	for _, user := range u {
		if user.Token == token {
			return user, nil
		}
	}

	return models.User{}, repository.ErrNoSuchUser
}
