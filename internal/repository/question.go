package repository

import (
	"context"
	"fmt"

	"fortuneteller/internal/data"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Question interface {
	AddQuestion(ctx context.Context, q data.Question) error
	FindUserQuestion(ctx context.Context, u data.User) ([]data.Question, error)
}

type questiondb struct {
	*pgxpool.Pool
}

func NewQuestionInterface(db *pgxpool.Pool) Question {
	return &questiondb{db}
}

func (qdb questiondb) AddQuestion(ctx context.Context, question data.Question) error {

	tx, err := qdb.Begin(ctx)
	if err != nil {
		return fmt.Errorf("can't start a transaction: %v", err)
	}

	const q = `INSERT INTO Questions (question_id, question_data, question_answer, question_book, question_owner)
				VALUES ($1, $2, $3, $4, $5)`

	_, err = tx.Exec(ctx, q, question.ID, question.Question, question.Answer, question.BData, question.Owner)
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("can't insert new question: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit error: %v", err)
	}
	return nil
}

func (qdb questiondb) FindUserQuestion(ctx context.Context, user data.User) ([]data.Question, error) {
	panic("implement me")
}
