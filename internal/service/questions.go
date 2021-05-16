package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"fortuneteller/internal/data"
	"fortuneteller/internal/logger"
	"fortuneteller/internal/repository"
	"github.com/sirupsen/logrus"

	"github.com/google/uuid"
)

type QuestionService struct {
	UserRepository     repository.User
	BookRepository     repository.Book
	QuestionRepository repository.Question
}

func (qs QuestionService) AskQuestion(ctx context.Context, question string, user data.User, askData data.FromAskData) (data.Question, error) {
	logger.WithFunction().WithField("question", question).Info("starting to ask question")

	book := qs.BookRepository.GetBookKey(askData.Name)

	encryptedQuestion := book.Encrypt([]byte(question))
	logger.WithFunction().WithField(
		"encrypted_question", question).Info("the question is encrypted for writing to the database")

	answer, err := qs.BookRepository.FindRowInBook(askData.Name, askData.Row)
	if err != nil {
		return data.Question{}, fmt.Errorf("cant find answer : %v", err)
	}

	bdata := askData.Name + ":" + strconv.Itoa(askData.Row)
	q := data.Question{
		ID:       uuid.New().String(),
		Question: string(encryptedQuestion),
		Answer:   answer,
		BData:    bdata,
		Owner:    user.ID,
	}
	if err := qs.QuestionRepository.AddQuestion(ctx, q); err != nil {
		return data.Question{}, fmt.Errorf("cant add question : %v", err)
	}
	return q, nil
}

func (qs QuestionService) ListUserEncryptedQuestions(ctx context.Context, username string) ([]data.Question, error) {
	user, err := qs.UserRepository.FindUserByName(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("can't find user by name %s : %v", username, err)
	}
	return qs.QuestionRepository.FindUserQuestion(ctx, user.ID)
}

func (qs QuestionService) ListUserDecryptedQuestions(ctx context.Context, username string) ([]data.Question, error) {
	user, err := qs.UserRepository.FindUserByName(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("can't find user by name %s : %v", username, err)
	}
	questions, err := qs.QuestionRepository.FindUserQuestion(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("can't find user questions: %v", err)
	} else if questions == nil {
		return nil, fmt.Errorf("empty questions")
	}

	for i, question := range questions {
		bookname := strings.Split(question.BData, ":")[0]
		book := qs.BookRepository.GetBookKey(bookname)
		decryptedQuestion, err := book.Decrypt([]byte(question.Question))
		if err != nil {
			return nil, fmt.Errorf("cant decrypt question : %v", err)
		}
		questions[i].Question = string(decryptedQuestion)
	}

	return questions, nil
}

func (qs QuestionService) ListBooks() ([]data.Book, error) {
	return qs.BookRepository.ListBooks()
}

func (qs QuestionService) FindUserQuestionByID(ctx context.Context, id string, username string) (data.Question, error) {
	question, err := qs.QuestionRepository.FindQuestionByID(ctx, id)
	if err != nil {
		return data.Question{}, fmt.Errorf("can't find question: %v", err)
	}
	user, err := qs.UserRepository.FindUserByName(ctx, username)
	if err != nil {
		return data.Question{}, fmt.Errorf("can't find user by name %s : %v", username, err)
	}

	if question.Owner == user.ID {
		bookname := strings.Split(question.BData, ":")[0]
		book := qs.BookRepository.GetBookKey(bookname)
		b, err := book.Decrypt([]byte(question.Question))
		if err != nil {
			return data.Question{}, fmt.Errorf("cant decrypt question : %v", err)
		}
		question.Question = string(b)
		logger.WithFunction().WithField("decrypted_question", string(b)).Info("the question is decrypted")
	}
	return question, nil
}

func (qs QuestionService) AskQuestionFromAnotherBook(ctx context.Context,
	question data.Question, usernameFromCookie, bookname string) (data.Question, error) {

	questionOwner, err := qs.UserRepository.FindUserByID(ctx, question.Owner)
	if err != nil {
		return data.Question{}, fmt.Errorf("can't find user by name %s : %v", questionOwner, err)
	}

	// Find row in new book for question from request
	bookData := strings.Split(question.BData, ":")
	if bookData[0] == bookname {
		logger.WithFunction().WithFields(logrus.Fields{
			"book":     bookData[0],
			"new_book": bookname,
		}).Info("book name and new book name are equal")
		return question, nil
	}
	row, _ := strconv.Atoi(bookData[1])
	answer, err := qs.BookRepository.FindRowInBook(bookname, row)
	if err != nil {
		return data.Question{}, fmt.Errorf("cant find answer : %v", err)
	}

	if usernameFromCookie == questionOwner.Username {
		return data.Question{
			Question: question.Question,
			Answer:   answer,
		}, nil
	} else {
		// return encrypted question on new book key
		book := qs.BookRepository.GetBookKey(bookData[0])
		newBook := qs.BookRepository.GetBookKey(bookname)

		decryptedQuestion, err := book.Decrypt([]byte(question.Question))
		if err != nil {
			return data.Question{}, fmt.Errorf("cant decrypt question : %v", err)
		}

		encryptedQuestion := newBook.Encrypt(decryptedQuestion)

		return data.Question{
			Question: string(encryptedQuestion),
			Answer:   answer,
		}, nil
	}
}
