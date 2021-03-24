package service

import (
	"fmt"

	"fortuneteller/internal/crypto"
	"fortuneteller/internal/models"
	"fortuneteller/internal/repository"
)

type QuestionService struct {
	repoq repository.Question
	repou repository.User
	repob repository.Book
	cryp  crypto.AwesomeCrypto
}

func (qs QuestionService) AskQuestion(question string, user models.User, book models.BookData) (string, error) {
	encryptedQuestion, err := qs.cryp.Encrypt([]byte(question))
	if err != nil {
		return "", fmt.Errorf("cant encrypt question : %v", err)
	}

	answer, err := qs.repob.FindRowInBook(book)
	if err != nil {
		return "", fmt.Errorf("cant find answer : %v", err)
	}

	q := models.Question{
		Question: string(encryptedQuestion),
		Answer:   answer,
		BData:    book,
		Owner:    user,
	}
	if err := qs.repoq.AddQuestion(q); err != nil {
		return "", fmt.Errorf("cant add question : %v", err)
	}
	return answer, nil
}

func (qs QuestionService) ListUserEncryptedQuestions(username string) ([]models.Question, error) {
	user, err := qs.repou.FindUserByName(username)
	if err != nil {
		return nil, fmt.Errorf("cant list user questions : %v", err)
	}
	return qs.repoq.FindUserQuestion(user)
}

func (qs QuestionService) ListUserDecryptedQuestions(username string) ([]models.Question, error) {
	user, err := qs.repou.FindUserByName(username)
	if err != nil {
		return nil, fmt.Errorf("cant list user questions : %v", err)
	}
	questions, err := qs.repoq.FindUserQuestion(user)
	for i, question := range questions {
		decryptedQuestion, err := qs.cryp.Decrypt([]byte(question.Question))
		if err != nil {
			return nil, fmt.Errorf("cant decrypt question : %v", err)
		}
		questions[i].Question = string(decryptedQuestion)
	}

	return questions, nil
}
