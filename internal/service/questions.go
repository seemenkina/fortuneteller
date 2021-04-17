package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"fortuneteller/internal/data"
	"fortuneteller/internal/repository"
	"github.com/google/uuid"
)

type QuestionService struct {
	Repoq repository.Question
	Repou repository.User
	Repob repository.Book
}

func (qs QuestionService) AskQuestion(ctx context.Context, question string, user data.User, askData data.FromAskData) (data.Question, error) {
	book, err := qs.Repob.GetBookKey(askData.Name)
	if err != nil {
		return data.Question{}, fmt.Errorf("can't find book %s : %v", askData.Name, err)
	}
	log.Printf("QUESTION: %s ", question)
	encryptedQuestion, err := book.Encrypt([]byte(question))
	if err != nil {
		return data.Question{}, fmt.Errorf("cant encrypt question : %v", err)
	}

	answer, err := qs.Repob.FindRowInBook(askData.Name, askData.Row)
	if err != nil {
		return data.Question{}, fmt.Errorf("cant find answer : %v", err)
	}

	log.Printf("EQUESTION: %s ", encryptedQuestion)
	bdata := askData.Name + ":" + strconv.Itoa(askData.Row)
	q := data.Question{
		ID:       uuid.New().String(),
		Question: string(encryptedQuestion),
		Answer:   answer,
		BData:    bdata,
		Owner:    user.ID,
	}
	if err := qs.Repoq.AddQuestion(ctx, q); err != nil {
		return data.Question{}, fmt.Errorf("cant add question : %v", err)
	}
	return q, nil
}

func (qs QuestionService) ListUserEncryptedQuestions(ctx context.Context, username string) ([]data.Question, error) {
	user, err := qs.Repou.FindUserByName(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("can't find user by name %s : %v", username, err)
	}
	return qs.Repoq.FindUserQuestion(ctx, user.ID)
}

func (qs QuestionService) ListUserDecryptedQuestions(ctx context.Context, username string) ([]data.Question, error) {
	user, err := qs.Repou.FindUserByName(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("can't find user by name %s : %v", username, err)
	}
	questions, err := qs.Repoq.FindUserQuestion(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("can't find user questions: %v", err)
	} else if questions == nil {
		return nil, fmt.Errorf("empty questions")
	}

	for i, question := range questions {
		bookname := strings.Split(question.BData, ":")[0]
		book, err := qs.Repob.GetBookKey(bookname)
		if err != nil {
			return nil, fmt.Errorf("can't find book %s : %v", bookname, err)
		}

		decryptedQuestion, err := book.Decrypt([]byte(question.Question))
		if err != nil {
			return nil, fmt.Errorf("cant decrypt question : %v", err)
		}
		questions[i].Question = string(decryptedQuestion)
	}

	return questions, nil
}

func (qs QuestionService) ListBooks() ([]data.Book, error) {
	return qs.Repob.ListBooks()
}

func (qs QuestionService) FindUserQuestionByID(ctx context.Context, id string, username string) (data.Question, error) {
	question, err := qs.Repoq.FindQuestionByID(ctx, id)
	if err != nil {
		return data.Question{}, fmt.Errorf("can't find question: %v", err)
	}
	user, err := qs.Repou.FindUserByName(ctx, username)
	if err != nil {
		return data.Question{}, fmt.Errorf("can't find user by name %s : %v", username, err)
	}

	if question.Owner == user.ID {
		bookname := strings.Split(question.BData, ":")[0]
		book, err := qs.Repob.GetBookKey(bookname)
		if err != nil {
			return data.Question{}, fmt.Errorf("can't find book %s : %v", bookname, err)
		}

		b, err := book.Decrypt([]byte(question.Question))
		if err != nil {
			return data.Question{}, fmt.Errorf("cant decrypt question : %v", err)
		}
		question.Question = string(b)
		log.Printf("DQUESTION: %v", string(b))
	}
	return question, nil
}

func (qs QuestionService) AskQuestionFromAnotherBook(ctx context.Context,
	question data.Question, usernameFromCookie, bookname string) (data.Question, error) {

	questionOwner, err := qs.Repou.FindUserByID(ctx, question.Owner)
	if err != nil {
		return data.Question{}, fmt.Errorf("can't find user by name %s : %v", questionOwner, err)
	}

	// Find row in new book for question from request
	bookData := strings.Split(question.BData, ":")
	if bookData[0] == bookname {
		return question, nil
	}
	row, _ := strconv.Atoi(bookData[1])
	answer, err := qs.Repob.FindRowInBook(bookname, row)
	if err != nil {
		return data.Question{}, fmt.Errorf("cant find answer : %v", err)
	}

	if usernameFromCookie == questionOwner.Username {
		return data.Question{
			Question: question.Question,
			Answer:   question.Answer,
		}, nil
	} else {
		// return encrypted question
		book, err := qs.Repob.GetBookKey(bookData[0])
		if err != nil {
			return data.Question{}, fmt.Errorf("can't find book %s : %v", bookname, err)
		}
		decryptedQuestion, err := book.Decrypt([]byte(question.Question))
		if err != nil {
			return data.Question{}, fmt.Errorf("cant encrypt question : %v", err)
		}

		newBook, err := qs.Repob.GetBookKey(bookname)
		if err != nil {
			return data.Question{}, fmt.Errorf("can't find new book %s : %v", bookname, err)
		}
		encryptedQuestion, err := newBook.Encrypt(decryptedQuestion)
		if err != nil {
			return data.Question{}, fmt.Errorf("cant encrypt question : %v", err)
		}

		return data.Question{
			Question: string(encryptedQuestion),
			Answer:   answer,
		}, nil
	}

}
