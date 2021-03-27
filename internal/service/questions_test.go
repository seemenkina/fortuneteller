package service

import (
	"testing"

	"fortuneteller/internal/mocks"
	"fortuneteller/internal/models"
	"github.com/stretchr/testify/suite"
)

type questionServiceSuite struct {
	suite.Suite

	question        *mocks.QuestionMock
	questionService QuestionService
	userService     UserService
	user            models.User
}

func TestQuestionService(t *testing.T) {
	suite.Run(t, new(questionServiceSuite))
}

func (s *questionServiceSuite) SetupTest() {
	s.question = &mocks.QuestionMock{}
	users := make(mocks.UserMock)
	s.questionService = QuestionService{
		repoq: s.question,
		repou: users,
		repob: mocks.NewBookMock(),
		cryp:  mocks.AwesomeCryptoMock{},
	}
	s.userService = UserService{
		Repo:  users,
		Token: mocks.TokenMock{},
	}
	username := "testUser"
	token, err := s.userService.Register(username)
	s.Require().NoError(err)
	s.Assert().NotEmpty(token)
	s.user = models.User{
		Username: username,
		Token:    token,
	}
}

func (s *questionServiceSuite) TestListUserQuestion() {
	book := models.BookData{
		Name: mocks.BookName,
		Row:  1,
	}
	question := "how are you?"

	answer, err := s.questionService.AskQuestion(question, models.User{
		Username: s.user.Username,
		Token:    s.user.Token,
	}, book)
	s.Require().NoError(err)
	s.Assert().NotEmpty(answer)

	questions, err := s.questionService.ListUserEncryptedQuestions(s.user.Username)
	s.Require().NoError(err)
	s.Assert().NotEmpty(questions)
	s.T().Logf("%+v\n", questions)

	questionsD, err := s.questionService.ListUserDecryptedQuestions(s.user.Username)
	s.Require().NoError(err)
	s.Assert().NotEmpty(questionsD)
	s.T().Logf("%+v\n", questionsD)
}

func (s *questionServiceSuite) TestAskQuestion() {
	book := models.BookData{
		Name: mocks.BookName,
		Row:  1,
	}
	question := "how are you?"

	answer, err := s.questionService.AskQuestion(question, models.User{
		Username: s.user.Username,
		Token:    s.user.Token,
	}, book)
	s.Require().NoError(err)
	s.Assert().NotEmpty(answer)
}
