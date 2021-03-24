package repository

import "fortuneteller/internal/models"

type Book interface {
	FindRowInBook(b models.BookData) (string, error)
}
