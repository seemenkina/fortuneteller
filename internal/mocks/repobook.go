package mocks

import (
	"fmt"

	"fortuneteller/internal/models"
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

func (b BookMock) FindRowInBook(book models.BookData) (string, error) {
	strs, ok := b[book.Name]
	if !ok {
		return "", fmt.Errorf("no such book")
	}
	if len(strs) <= book.Row {
		return "", fmt.Errorf("no such row")
	}
	return strs[book.Row], nil
}