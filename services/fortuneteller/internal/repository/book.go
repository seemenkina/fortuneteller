package repository

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"fortuneteller/internal/crypto"
	"fortuneteller/internal/data"
	"fortuneteller/internal/logger"
)

type Book interface {
	ListBooks() ([]data.Book, error)
	GetBookKey(book string) crypto.AwesomeCrypto
	FindRowInBook(book string, row int) (string, error)
}

type bookfs struct {
	Path    string
	BookKey map[string]crypto.IzzyWizzy
}

func NewBookFileSystem(path, keyPath string) Book {
	bookFiles, err := ioutil.ReadDir(path)
	if err != nil {
		logger.WithFunction().Errorf("unable to read books file system directory: %v", err)
		return nil
	}

	keysFiles, err := ioutil.ReadDir(keyPath)
	if err != nil {
		logger.WithFunction().Errorf("unable to read keys file system directory: %v", err)
		return nil
	}

	books := make(map[string]crypto.IzzyWizzy, len(bookFiles))
	for _, f := range bookFiles {
		fname := f.Name()
		logger.WithFunction().WithField("book_name", fname).Infof("starting to create book's key pair")
		if len(keysFiles) == 0 {
			if err := generateNewKey(books, fname, keyPath); err != nil {
				return nil
			}
		} else {
			key := crypto.LoadKeyFromFile(filepath.Join(keyPath, fname) + "_key")
			if (crypto.IzzyWizzy{} == key) {
				if err := generateNewKey(books, fname, keyPath); err != nil {
					return nil
				}
			} else {
				logger.WithFunction().WithField("file_name", fname).Info("key for book is load")
				books[fname] = key
			}
		}
	}

	return &bookfs{Path: path, BookKey: books}
}

func (bfs bookfs) FindRowInBook(book string, row int) (string, error) {
	filename := filepath.Join(bfs.Path, book)
	logger.WithFunction().WithFields(logrus.Fields{
		"book": filename,
		"row":  row,
	}).Info("starting to find row in book")

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
	logger.WithFunction().WithFields(logrus.Fields{
		"book":   filename,
		"result": str,
	}).Info("the row in book is successfully find")

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
		fname := f.Name()
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

func generateNewKey(books map[string]crypto.IzzyWizzy, fname, keyPath string) error {
	books[fname] = crypto.GenerateKeyPair()
	logger.WithFunction().WithFields(logrus.Fields{
		"public_key":  books[fname].PublicKey,
		"private_key": books[fname].PrivateKey,
	}).Info("new key pair are generated")

	if err := books[fname].SaveKeyOnFile(filepath.Join(keyPath, fname+"_key")); err != nil {
		logger.WithFunction().WithFields(logrus.Fields{
			"error":     err,
			"file_name": fname,
		}).Errorf("error: save book key for file")
		return err
	}
	return nil
}
