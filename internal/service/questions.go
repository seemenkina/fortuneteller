package service

import (
	"context"
	"fmt"
	"strconv"

	"fortuneteller/internal/crypto"
	"fortuneteller/internal/data"
	"fortuneteller/internal/repository"
	"github.com/google/uuid"
)

type QuestionService struct {
	Repoq repository.Question
	Repou repository.User
	Repob repository.Book
	Cryp  crypto.AwesomeCrypto
}

func (qs QuestionService) AskQuestion(ctx context.Context, question string, user data.User, book data.FromAskData) (data.Question, error) {
	encryptedQuestion, err := qs.Cryp.Encrypt([]byte(question))
	if err != nil {
		return data.Question{}, fmt.Errorf("cant encrypt question : %v", err)
	}

	answer, err := qs.Repob.FindRowInBook(ctx, book.Name, book.Row)
	if err != nil {
		return data.Question{}, fmt.Errorf("cant find answer : %v", err)
	}

	bdata := book.Name + ":" + strconv.Itoa(book.Row)
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
	if questions == nil {
		return nil, fmt.Errorf("empty questions")
	}
	for i, question := range questions {
		decryptedQuestion, err := qs.Cryp.Decrypt([]byte(question.Question))
		if err != nil {
			return nil, fmt.Errorf("cant decrypt question : %v", err)
		}
		questions[i].Question = string(decryptedQuestion)
	}

	return questions, nil
}

func (qs QuestionService) ListBooks(ctx context.Context) ([]data.Book, error) {
	return qs.Repob.ListBooks(ctx)
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
		b, err := qs.Cryp.Decrypt([]byte(question.Question))
		if err != nil {
			return data.Question{}, fmt.Errorf("cant decrypt question : %v", err)
		}
		question.Question = string(b)
	}
	return question, nil
}
