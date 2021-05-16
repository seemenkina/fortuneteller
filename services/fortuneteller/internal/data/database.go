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

	poolconn, err := pgxpool.Connect(ctx, u.String())
	if err != nil {
		return nil, fmt.Errorf("can't connect to pgxpool: %v", poolconn)
	}
	return Migrate(ctx, poolconn)
}
