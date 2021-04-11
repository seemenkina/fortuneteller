package mocks

import (
	"context"
	"fmt"

	"fortuneteller/internal/data"
	"fortuneteller/internal/repository"
)

type UserMock map[string]data.User

func (u UserMock) AddUser(ctx context.Context, user data.User) error {
	_, ok := u[user.Username]
	if ok {
		return fmt.Errorf("user exists")
	}
	u[user.Username] = user
	return nil
}

func (u UserMock) FindUserByName(ctx context.Context, name string) (data.User, error) {
	user, ok := u[name]
	if ok {
		return user, nil
	} else {
		return data.User{}, repository.ErrNoSuchUser
	}
}

func (u UserMock) AllUsers(ctx context.Context) ([]data.User, error) {
	var users []data.User
	for _, user := range u {
		users = append(users, user)
	}
	return users, nil
}

func (u UserMock) FindUserByToken(ctx context.Context, token string) (data.User, error) {
	for _, user := range u {
		if user.Token == token {
			return user, nil
		}
	}

	return data.User{}, repository.ErrNoSuchUser
}
