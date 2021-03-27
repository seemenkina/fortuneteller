package service

import (
	"log"
	"testing"

	"fortuneteller/internal/mocks"
	"github.com/stretchr/testify/suite"
)

type userServiceSuite struct {
	suite.Suite

	user    mocks.UserMock
	service UserService
}

func TestUserService(t *testing.T) {
	suite.Run(t, new(userServiceSuite))
}

func (s *userServiceSuite) SetupTest() {
	s.user = make(mocks.UserMock)
	s.service = UserService{
		Repo:  s.user,
		Token: mocks.TokenMock{},
	}
}

func (s *userServiceSuite) TestRegisterUser() {
	username := "testUser"

	token, err := s.service.Register(username)
	s.Require().NoError(err)
	s.Assert().NotEmpty(token)

	s.Assert().Contains(s.user, username)
	s.Assert().Equal(s.user[username].Token, token)
}

func (s *userServiceSuite) TestLoginUser() {
	username := "testUser"

	token, err := s.service.Register(username)
	s.Require().NoError(err)
	s.Assert().NotEmpty(token)

	s.Assert().Contains(s.user, username)
	s.Assert().Equal(s.user[username].Token, token)

	u, err := s.service.Login(token)
	s.Require().NoError(err)
	s.Assert().NotEmpty(u)
}

func (s *userServiceSuite) TestRegisterSameUser() {
	username := "testUser"

	token, err := s.service.Register(username)
	s.Require().NoError(err)
	s.Assert().NotEmpty(token)

	s.Assert().Contains(s.user, username)
	s.Assert().Equal(s.user[username].Token, token)

	_, err = s.service.Register(username)
	s.Require().Error(err)
}

func (s *userServiceSuite) TestListUser() {
	usernames := []string{"testUser", "alpha", "beta"}

	for _, username := range usernames {
		token, err := s.service.Register(username)
		s.Require().NoError(err)
		s.Assert().NotEmpty(token)

		s.Assert().Contains(s.user, username)
		s.Assert().Equal(s.user[username].Token, token)
	}

	users, err := s.service.ListUsers()
	s.Require().NoError(err)
	log.Printf("users : %q", users)
}
