package mocks

import (
	"fmt"

	"fortuneteller/internal/data"
)

const BookName = "Great Book"

type BookMock map[string][]string

func NewBookMock() BookMock {
	return BookMock{
		BookName: []string{
			"- Hello!",
			"Said Alice",
			"-How are you?",
		},
	}
}

func (b BookMock) FindRowInBook(book data.BookData) (string, error) {
	strs, ok := b[book.Name]
	if !ok {
		return "", fmt.Errorf("no such book")
	}
	if len(strs) <= book.Row {
		return "", fmt.Errorf("no such row")
	}
	return strs[book.Row], nil
}

func (b BookMock) ListBooks() ([]data.BookData, error) {
	var books []data.BookData
	for name, row := range b {
		books = append(books, data.BookData{
			Name: name,
			Row:  len(row),
		})
	}
	return books, nil
}
