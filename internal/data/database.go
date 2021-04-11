package data

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Config struct {
	User         string
	Password     string
	Host         string
	DatabaseName string
	DisableTLS   string
}

func Open(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	q := make(url.Values)
	q.Set("sslmode", cfg.DisableTLS)
	q.Set("timezone", "utc")
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.DatabaseName,
		RawQuery: q.Encode(),
	}

	return pgxpool.Connect(ctx, u.String())
}

func createUserTable(ctx context.Context, conn *pgxpool.Pool) (*pgxpool.Pool, error) {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't start a transaction: %v", err)
	}

	const q = `CREATE TABLE if not exists Users (
	user_id integer PRIMARY KEY, 
	user_token TEXT,
	user_name TEXT );`

	if _, err := tx.Exec(ctx, q); err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("can't create users database: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit error: %v", err)
	}
	return conn, err
}
