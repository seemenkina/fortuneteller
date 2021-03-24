package repository

import "fortuneteller/internal/models"

type Question interface {
	AddQuestion(q models.Question) error
	FindUserQuestion(u models.User) ([]models.Question, error)
}
