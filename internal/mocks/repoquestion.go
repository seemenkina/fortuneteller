package mocks

import (
	"context"

	"fortuneteller/internal/data"
)

type QuestionMock []data.Question

func (q *QuestionMock) AddQuestion(ctx context.Context, question data.Question) error {
	*q = append(*q, question)
	return nil
}

func (q *QuestionMock) FindUserQuestion(ctx context.Context, user data.User) ([]data.Question, error) {
	var questions []data.Question
	for _, question := range *q {
		if question.Owner == user.ID {
			questions = append(questions, question)
		}
	}
	return questions, nil
}
