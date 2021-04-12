package repository

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"fortuneteller/internal/data"
)

type Book interface {
	FindRowInBook(ctx context.Context, book string, row int) (string, error)
	ListBooks(ctx context.Context) ([]data.Book, error)
}

type bookfs struct {
	Path string
}

func NewBookFileSystem(path string) Book {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModeDir)
	}
	return &bookfs{Path: path}
}

func (bfs bookfs) FindRowInBook(ctx context.Context, book string, row int) (string, error) {
	filename := filepath.Join(bfs.Path, book+".txt")
	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("can't open file %s : %v", book, err)
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	result := 0
	str := ""
	for scanner.Scan() {
		if result == row {
			str = scanner.Text()
			break
		}
		result++
	}
	if scanner.Err() != nil {
		return "", fmt.Errorf("scanner error: %v", err)
	}
	return str, f.Close()
}

func (bfs bookfs) ListBooks(ctx context.Context) ([]data.Book, error) {
	files, err := ioutil.ReadDir(bfs.Path)
	if err != nil {
		return nil, fmt.Errorf("can't open books directory: %v", err)
	}
	var books []data.Book
	for _, f := range files {
		// fmt.Println(f.Name())
		fname := f.Name()
		lenf := len(fname) - len(filepath.Ext(fname))
		// fmt.Printf("l: %d", lenf)
		books = append(books, data.Book{
			Name: fname[:lenf],
			Rows: rowsInFile(filepath.Join(bfs.Path, f.Name())),
		})
	}
	return books, err
}

func rowsInFile(filename string) int {
	f, err := os.Open(filename)
	if err != nil {
		return -1
	}
	defer func() {
		_ = f.Close()
	}()
	// Create new Scanner.
	scanner := bufio.NewScanner(f)
	result := 0
	for scanner.Scan() {
		result++
	}
	if scanner.Err() != nil {
		return -1
	}
	_ = f.Close()
	return result
}
