package data

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

func Migrate(ctx context.Context, conn *pgxpool.Pool) (*pgxpool.Pool, error) {
	var err error
	conn, err = createUserTable(ctx, conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("can't create user table: %v", err)
	}

	conn, err = createQuestionsTable(ctx, conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("can't create questions table: %v", err)
	}
	return conn, err
}

func createUserTable(ctx context.Context, conn *pgxpool.Pool) (*pgxpool.Pool, error) {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't start a transaction: %v", err)
	}

	const q = `CREATE TABLE if not exists Users (
	user_id UUID NOT NULL PRIMARY KEY, 
	user_token TEXT,
	user_name TEXT );`

	if _, err := tx.Exec(ctx, q); err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("can't create users table: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit error: %v", err)
	}
	return conn, err
}

func createQuestionsTable(ctx context.Context, conn *pgxpool.Pool) (*pgxpool.Pool, error) {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't start a transaction: %v", err)
	}

	const q = `CREATE TABLE if not exists Questions (
	question_id UUID NOT NULL PRIMARY KEY, 
	question_data TEXT,
	question_answer TEXT,
	question_book TEXT
	question_owner UUID,
	CONSTRAINT fk_owner
		FOREIGN KEY (question_owner)
		REFERENCES Users(user_id) ON DELETE RESTRICT);`

	if _, err := tx.Exec(ctx, q); err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("can't create questions table: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit error: %v", err)
	}
	return conn, err
}
