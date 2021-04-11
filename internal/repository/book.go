package repository

import (
	"fortuneteller/internal/data"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Book interface {
	FindRowInBook(b data.BookData) (string, error)
	ListBooks() ([]data.BookData, error)
}

type bookdb struct {
	*pgxpool.Pool
}

func NewBookInterface(db *pgxpool.Pool) Book {
	return &bookdb{db}
}

func (bdb bookdb) FindRowInBook(book data.BookData) (string, error) {
	panic("implement me")
}

func (bdb bookdb) ListBooks() ([]data.BookData, error) {
	panic("implement me")
}
