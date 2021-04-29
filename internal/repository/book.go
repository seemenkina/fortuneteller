package repository

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"fortuneteller/internal/crypto"
	"fortuneteller/internal/data"
)

type Book interface {
	FindRowInBook(book string, row int) (string, error)
	ListBooks() ([]data.Book, error)
	GetBookKey(book string) crypto.AwesomeCrypto
}

type bookfs struct {
	Path    string
	BookKey map[string]crypto.IzzyWizzy
}

func NewBookFileSystem(path, keyPath string) Book {
	bookFiles, err := ioutil.ReadDir(path)
	if err != nil {
		return nil
	}

	keysFiles, err := ioutil.ReadDir(keyPath)
	if err != nil {
		return nil
	}

	books := make(map[string]crypto.IzzyWizzy, len(bookFiles))
	for _, f := range bookFiles {
		fname := f.Name()
		log.Printf("START CREATE KEY FOR BOOK %s: ", fname)
		if len(keysFiles) == 0 {
			books[fname] = crypto.GenerateKeyPair()
			if err := books[fname].SaveKeyOnFile(filepath.Join(keyPath, fname+"_key")); err != nil {
				log.Printf("ERROR SAVE: %v", err)
				return nil
			}
		} else {
			key := crypto.LoadKeyFromFile(filepath.Join(keyPath, fname) + "_key")
			if (crypto.IzzyWizzy{} == key) {
				books[fname] = crypto.GenerateKeyPair()
				if err := books[fname].SaveKeyOnFile(filepath.Join(keyPath, fname+"_key")); err != nil {
					log.Printf("ERROR SAVE: %v", err)
					return nil
				}
			} else {
				log.Printf("KEY IS LOAD")
				books[fname] = key
			}
		}
	}

	return &bookfs{Path: path, BookKey: books}
}

func (bfs bookfs) FindRowInBook(book string, row int) (string, error) {
	filename := filepath.Join(bfs.Path, book)
	log.Printf("FIND ROW %d in %s", row, filename)
	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("can't open file %s : %v", book, err)
	}
	defer func() {
		_ = f.Close()
	}()

	if rowsInFile(filename) < row {
		return "", fmt.Errorf("row in books must be greater or equal then find row")
	}
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
	log.Printf("ROW %d IN %s : %s", result, filename, str)
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
		// fmt.Printf("l: %d", lenf)
		books = append(books, data.Book{
			Name: fname,
			Rows: rowsInFile(filepath.Join(bfs.Path, fname)),
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

func (bfs bookfs) GetBookKey(book string) crypto.AwesomeCrypto {
	return bfs.BookKey[book]
}
