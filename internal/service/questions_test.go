package service

import (
	"context"
	"encoding/hex"
	"testing"

	"fortuneteller/internal/crypto"
	"fortuneteller/internal/data"
	"fortuneteller/internal/mocks"
	"github.com/stretchr/testify/suite"
)

type questionServiceSuite struct {
	suite.Suite

	question        *mocks.QuestionMock
	questionService QuestionService
	userService     UserService
	user            data.User
	ctx             context.Context
}

func TestQuestionService(t *testing.T) {
	suite.Run(t, new(questionServiceSuite))
}

func (s *questionServiceSuite) SetupTest() {
	s.question = &mocks.QuestionMock{}
	users := make(mocks.UserMock)
	s.ctx = context.Background()
	key := hex.EncodeToString([]byte("~ThisIsMagicKey~"))
	s.Assert().NotEmpty(key)
	s.questionService = QuestionService{
		Repoq: s.question,
		Repou: users,
		Repob: mocks.NewBookMock(),
		Cryp: crypto.IzzyWizzy{
			Key: []byte(key),
		},
	}
	s.userService = UserService{
		Repo:  users,
		Token: mocks.TokenMock{},
	}
	username := "testUser"
	token, err := s.userService.Register(s.ctx, username)
	s.Require().NoError(err)
	s.Assert().NotEmpty(token)
	s.user = data.User{
		Username: username,
		Token:    token,
	}
}

func (s *questionServiceSuite) TestListUserQuestion() {
	book := data.BookData{
		Name: mocks.BookName,
		Row:  1,
	}
	question := "how are you?"

	answer, err := s.questionService.AskQuestion(s.ctx, question, data.User{
		Username: s.user.Username,
		Token:    s.user.Token,
	}, book)
	s.Require().NoError(err)
	s.Assert().NotEmpty(answer)

	questions, err := s.questionService.ListUserEncryptedQuestions(s.ctx, s.user.Username)
	s.Require().NoError(err)
	s.Assert().NotEmpty(questions)
	s.T().Logf("%+v\n", questions)

	questionsD, err := s.questionService.ListUserDecryptedQuestions(s.ctx, s.user.Username)
	s.Require().NoError(err)
	s.Assert().NotEmpty(questionsD)
	s.T().Logf("%+v\n", questionsD)
}

func (s *questionServiceSuite) TestAskQuestion() {
	book := data.BookData{
		Name: mocks.BookName,
		Row:  1,
	}
	question := "how are you?"

	answer, err := s.questionService.AskQuestion(s.ctx, question, data.User{
		Username: s.user.Username,
		Token:    s.user.Token,
	}, book)
	s.Require().NoError(err)
	s.Assert().NotEmpty(answer)
}
