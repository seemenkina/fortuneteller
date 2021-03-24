package mocks

import (
	"fortuneteller/internal/models"
)

type QuestionMock []models.Question

func (q *QuestionMock) AddQuestion(question models.Question) error {
	*q = append(*q, question)
	return nil
}

func (q *QuestionMock) FindUserQuestion(user models.User) ([]models.Question, error) {
	var questions []models.Question
	for _, question := range *q {
		if question.Owner == user {
			questions = append(questions, question)
		}
	}
	return questions, nil
}
