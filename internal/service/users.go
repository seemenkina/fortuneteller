package service

import (
	"errors"
	"fmt"

	"fortuneteller/internal/crypto"
	"fortuneteller/internal/models"
	"fortuneteller/internal/repository"
)

type UserService struct {
	repo  repository.User
	token crypto.Token
}

func (us UserService) Register(username string) (string, error) {
	_, err := us.repo.FindUserByName(username)
	switch {
	case errors.Is(err, repository.ErrNoSuchUser):
		// all ok
	case err == nil:
		return "", fmt.Errorf("user alredy exists")
	default:
		return "", fmt.Errorf("cant find user : %v", err)
	}

	token, err := us.token.CreateToken(username)
	if err != nil {
		return "", fmt.Errorf("cant create token : %v", err)
	}

	user := models.User{
		Username: username,
		Token:    token,
	}
	if err := us.repo.AddUser(user); err != nil {
		return "", fmt.Errorf("cant add user : %v", err)
	}
	return token, nil
}

func (us UserService) Login(token string) (models.User, error) {
	user, err := us.repo.FindUserByToken(token)
	switch {
	case err == nil:
		return user, nil
	case errors.Is(err, repository.ErrNoSuchUser):
		return models.User{}, fmt.Errorf("cant find user : %v", err)
	default:
		return models.User{}, fmt.Errorf("cant login user : %v", err)
	}
}

func (us UserService) ListUsers() ([]models.User, error) {
	return us.repo.AllUsers()
}
