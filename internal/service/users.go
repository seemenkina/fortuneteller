package service

import (
	"context"
	"errors"
	"fmt"

	"fortuneteller/internal/crypto"
	"fortuneteller/internal/data"
	"fortuneteller/internal/repository"

	"github.com/google/uuid"
)

type UserService struct {
	UserRepository repository.User
	Token          crypto.Token
}

func (us UserService) Register(ctx context.Context, username string) (string, error) {
	_, err := us.UserRepository.FindUserByName(ctx, username)
	switch {
	case errors.Is(err, repository.ErrNoSuchUser):
		// all ok
	case err == nil:
		return "", fmt.Errorf("user alredy exists")
	default:
		return "", fmt.Errorf("cant find user : %v", err)
	}

	token, err := us.Token.CreateToken(username)
	if err != nil {
		return "", fmt.Errorf("cant create Token : %v", err)
	}

	user := data.User{
		Username: username,
		Token:    token,
		ID:       uuid.New().String(),
	}
	if err := us.UserRepository.AddUser(ctx, user); err != nil {
		return "", fmt.Errorf("cant add user : %v", err)
	}
	return token, nil
}

func (us UserService) Login(ctx context.Context, token string) (data.User, error) {
	user, err := us.UserRepository.FindUserByToken(ctx, token)
	switch {
	case err == nil:
		return user, nil
	case errors.Is(err, repository.ErrNoSuchUser):
		return data.User{}, fmt.Errorf("cant find user : %v", err)
	default:
		return data.User{}, fmt.Errorf("cant login user : %v", err)
	}
}

func (us UserService) ListUsers(ctx context.Context) ([]data.User, error) {
	return us.UserRepository.AllUsers(ctx)
}
