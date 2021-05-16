package repository

import (
	"context"
	"errors"
	"fmt"

	"fortuneteller/internal/data"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
)

var ErrNoSuchUser = errors.New("no such user")

type User interface {
	AddUser(ctx context.Context, u data.User) error
	AllUsers(ctx context.Context) ([]data.User, error)
	FindUserByID(ctx context.Context, id string) (data.User, error)
	FindUserByName(ctx context.Context, name string) (data.User, error)
	FindUserByToken(ctx context.Context, token string) (data.User, error)
}

type userdb struct {
	*pgxpool.Pool
}

func NewUserInterface(db *pgxpool.Pool) User {
	return &userdb{db}
}

func (udb userdb) AddUser(ctx context.Context, user data.User) error {
	tx, err := udb.Begin(ctx)
	if err != nil {
		return fmt.Errorf("can't start a transaction: %v", err)
	}

	const q = `INSERT INTO Users (user_id, user_token, user_name)
				VALUES ($1, $2, $3)`

	_, err = tx.Exec(ctx, q, user.ID, user.Token, user.Username)
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("can't insert new user: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit error: %v", err)
	}
	return nil
}

func (udb userdb) AllUsers(ctx context.Context) ([]data.User, error) {
	var users []data.User
	const q = `SELECT * FROM Users`

	if err := pgxscan.Select(ctx, udb, &users, q); err != nil {
		return nil, fmt.Errorf("can't get all users from DB : %v", err)
	}
	return users, nil
}

func (udb userdb) FindUserByName(ctx context.Context, name string) (data.User, error) {
	const q = `SELECT user_id, user_name, user_token FROM Users
				WHERE user_name = $1`

	var user data.User
	row := udb.QueryRow(ctx, q, name)
	if err := row.Scan(&user.ID, &user.Username, &user.Token); err != nil {
		return data.User{}, ErrNoSuchUser
	}

	return user, nil
}

func (udb userdb) FindUserByToken(ctx context.Context, token string) (data.User, error) {
	const q = `SELECT user_id, user_name, user_token FROM Users
				WHERE user_token = $1`

	var user data.User
	row := udb.QueryRow(ctx, q, token)
	if err := row.Scan(&user.ID, &user.Username, &user.Token); err != nil {
		return data.User{}, ErrNoSuchUser
	}

	return user, nil
}

func (udb userdb) FindUserByID(ctx context.Context, id string) (data.User, error) {
	const q = `SELECT user_id, user_name, user_token FROM Users
				WHERE user_id = $1`

	var user data.User
	row := udb.QueryRow(ctx, q, id)
	if err := row.Scan(&user.ID, &user.Username, &user.Token); err != nil {
		return data.User{}, ErrNoSuchUser
	}

	return user, nil
}
