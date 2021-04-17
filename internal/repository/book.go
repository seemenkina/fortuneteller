package repository

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"fortuneteller/internal/crypto"
	"fortuneteller/internal/data"
)

type Book interface {
	FindRowInBook(book string, row int) (string, error)
	ListBooks() ([]data.Book, error)
	GetBookKey(book string) (crypto.AwesomeCrypto, error)
}

type bookfs struct {
	Path    string
	BookKey map[string]crypto.AwesomeCrypto
}

func NewBookFileSystem(path string) Book {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil
	}
	books := make(map[string]crypto.AwesomeCrypto, len(files))
	for _, f := range files {
		fname := f.Name()
		lenf := len(fname) - len(filepath.Ext(fname))
		books[fname[:lenf]] = crypto.GenerateKeyPair()
	}
	return &bookfs{Path: path, BookKey: books}
}

func (bfs bookfs) FindRowInBook(book string, row int) (string, error) {
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

func (bfs bookfs) ListBooks() ([]data.Book, error) {
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

func (bfs bookfs) GetBookKey(book string) (crypto.AwesomeCrypto, error) {
	if value, ok := bfs.BookKey[book]; ok {
		return value, nil
	} else {
		return crypto.IzzyWizzy{}, fmt.Errorf("no such book %s", book)
	}
}
