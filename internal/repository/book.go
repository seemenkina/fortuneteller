package repository

import (
	"context"
	"fmt"

	"fortuneteller/internal/data"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Book interface {
	FindRowInBook(ctx context.Context, book string, row int) (string, error)
	ListBooks(ctx context.Context) ([]data.Book, error)
	AddBook(ctx context.Context, book data.Book) error
}

type bookdb struct {
	*pgxpool.Pool
}

func NewBookInterface(db *pgxpool.Pool) Book {
	return &bookdb{db}
}

func (bdb bookdb) FindRowInBook(ctx context.Context, book string, row int) (string, error) {
	const q = `SELECT book_data FROM Books WHERE book_name = $1`
	var rows []string
	rawData := bdb.QueryRow(ctx, q, book)
	if err := rawData.Scan(&rows); err != nil {
		return "", fmt.Errorf("can't retrieve book data: %v", err)
	}

	if len(rows) < row {
		return "", fmt.Errorf("can't find row in book")
	}
	return rows[row], nil
}

func (bdb bookdb) ListBooks(ctx context.Context) ([]data.Book, error) {
	var books []data.Book
	const q = `SELECT * FROM Books`

	if err := pgxscan.Select(ctx, bdb, &books, q); err != nil {
		return nil, fmt.Errorf("can't get all books from DB : %v", err)
	}

	return books, nil
}

func (bdb bookdb) AddBook(ctx context.Context, book data.Book) error {
	tx, err := bdb.Begin(ctx)
	if err != nil {
		return fmt.Errorf("can't start a transaction: %v", err)
	}

	const q = `INSERT INTO Books (book_id, book_name, book_len, book_data)
				VALUES ($1, $2, $3, $4)`

	_, err = tx.Exec(ctx, q, book.ID, book.Name, book.Rows, book.Data)
	if err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("can't insert new book: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit error: %v", err)
	}
	return nil
}
