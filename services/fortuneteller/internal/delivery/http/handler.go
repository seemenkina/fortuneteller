package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"fortuneteller/internal/data"
	"fortuneteller/internal/logger"
	"fortuneteller/internal/service"
)

type UserSubrouter struct {
	mux.Router
	UserService     *service.UserService
	QuestionService *service.QuestionService
}

// data in { username }
func (usersubrouter UserSubrouter) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse register form: %v", err)
		return
	}
	username := r.Form.Get("username")
	if len(username) == 0 {
		writeError(w, http.StatusBadRequest, "username in request form is empty")
		return
	}

	logger.WithFunction().WithField("username", username).Info("starting the user registration")
	token, err := usersubrouter.UserService.Register(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "can't register user: %v", err)
		return
	}

	logger.WithFunction().WithFields(logrus.Fields{
		"username": username,
		"token":    token,
	}).Info("the user is successfully register")

	setCookie(w, r, "tokencookie", token)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"redirect": "/register#" + token,
		"username": username,
	})
}

// data in { token }
func (usersubrouter UserSubrouter) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse login form: %v", err)
		return
	}

	token := r.Form.Get("token")
	if len(token) == 0 {
		writeError(w, http.StatusBadRequest, "token in request form is empty")
		return
	}

	logger.WithFunction().WithField("token", token).Info("starting the user login")
	user, err := usersubrouter.UserService.Login(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "can't login user: %v", err)
		return
	}
	logger.WithFunction().WithFields(logrus.Fields{
		"username": user.Username,
		"token":    user.Token,
	}).Info("the user is successfully login")

	setCookie(w, r, "tokencookie", token)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"redirect": "/homepage",
		"username": user.Username,
	})
}

// data in { }
func (usersubrouter UserSubrouter) ListUsers(w http.ResponseWriter, r *http.Request) {
	logger.WithFunction().Info("Start list users")
	users, err := usersubrouter.UserService.ListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to return all users: %v", err)
		return
	}
	// TODO: return last n usernames
	usernames := make([]string, len(users))
	for i, u := range users {
		usernames[i] = u.Username
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"users": usernames,
	})
}

// data in { username }
func (usersubrouter UserSubrouter) GetUserQuestions(w http.ResponseWriter, r *http.Request) {
	tokenFromCookie := tokenFromReq(r)

	usernameFromCookie, err := usersubrouter.UserService.Token.GetUsername(tokenFromCookie)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to get username by token: %v", err)
		return
	}
	logger.WithFunction().WithField("username", usernameFromCookie).Info("get username from cookie")

	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse user questions form: %v", err)
		return
	}
	username := r.Form.Get("username")
	logger.WithFunction().WithField("username", username).Info("get username from request form")

	var questions []data.Question
	if username == "" || usernameFromCookie == username {
		questions, err = usersubrouter.QuestionService.ListUserDecryptedQuestions(r.Context(), usernameFromCookie)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "unable to return decrypted user questions: %v", err)
			return
		}
	} else {
		questions, err = usersubrouter.QuestionService.ListUserEncryptedQuestions(r.Context(), username)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "unable to return encrypted user questions: %v", err)
			return
		}
	}

	logger.WithFunction().WithField("user_questions", questions).Info("return list of user questions")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"questions": questions,
	})

}

// data in { question id } (from URL)
func (usersubrouter UserSubrouter) GetAnswerFromBook(w http.ResponseWriter, r *http.Request) {
	tokenFromCookie := tokenFromReq(r)

	usernameFromCookie, err := usersubrouter.UserService.Token.GetUsername(tokenFromCookie)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to get username by token: %v", err)
		return
	}

	questionId := r.URL.Query().Get("id")
	logger.WithFunction().WithField("question_id", questionId).Info("get question id from url query")

	question, err := usersubrouter.QuestionService.FindUserQuestionByID(r.Context(), questionId, usernameFromCookie)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to get question by id: %v", err)
		return
	}

	books, err := usersubrouter.QuestionService.BookRepository.ListBooks()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to return all books: %v", err)
		return
	}

	logger.WithFunction().WithFields(logrus.Fields{
		"question": question.Question,
		"answer":   question.Answer,
		"books":    books,
	}).Info("return the answer to the question and available books")

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"Question": question.Question,
		"Answer":   question.Answer,
		"Books":    books,
	})

}

// data in { question id, new book id } (from url)
func (usersubrouter UserSubrouter) GetAnswerFromAnotherBook(w http.ResponseWriter, r *http.Request) {
	tokenFromCookie := tokenFromReq(r)

	usernameFromCookie, err := usersubrouter.UserService.Token.GetUsername(tokenFromCookie)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to get username by token: %v", err)
		return
	}

	logger.WithFunction().WithFields(logrus.Fields{
		"question_id": r.URL.Query().Get("id"),
		"book_id":     r.URL.Query().Get("id_book"),
	}).Info("try to ask another book same question")

	question, err := usersubrouter.QuestionService.FindUserQuestionByID(r.Context(),
		r.URL.Query().Get("id"), usernameFromCookie)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to get question by id: %v", err)
		return
	}

	otherAnswer, err := usersubrouter.QuestionService.AskQuestionFromAnotherBook(r.Context(),
		question, usernameFromCookie, r.URL.Query().Get("id_book"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to get question answer from another book: %v", err)
		return
	}

	pk := usersubrouter.QuestionService.BookRepository.GetBookKey(r.URL.Query().Get("id_book"))

	logger.WithFunction().WithFields(logrus.Fields{
		"question": question.Question,
		"answer":   question.Answer,
		"pubKey":   pk,
	}).Info("return the answer to the question and the public key of another book")

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"Question": otherAnswer.Question,
		"Answer":   otherAnswer.Answer,
		"PubKey":   pk,
	})

}

// data in { question, book , page }
func (usersubrouter UserSubrouter) AskQuestion(w http.ResponseWriter, r *http.Request) {
	tokenFromCookie := tokenFromReq(r)
	user, err := usersubrouter.UserService.UserRepository.FindUserByToken(r.Context(), tokenFromCookie)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to get user by token: %v", err)
		return
	}

	if err = r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse ask questions form: %v", err)
		return
	}

	logger.WithFunction().WithFields(logrus.Fields{
		"question": r.Form.Get("question"),
		"book":     r.Form.Get("book"),
		"page":     r.Form.Get("page"),
	}).Info("parse parameters from request")

	row, _ := strconv.Atoi(r.Form.Get("page"))
	question, err := usersubrouter.QuestionService.AskQuestion(
		r.Context(),
		r.Form.Get("question"),
		user,
		data.FromAskData{
			Name: r.Form.Get("book"),
			Row:  row,
		})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cant ask question: %v", err)
		return
	}

	logger.WithFunction().WithField("question", question).Info("new question is successfully asked")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"redirect": "/answer#" + question.ID,
	})
}

func (usersubrouter UserSubrouter) ListBooks(w http.ResponseWriter, r *http.Request) {
	books, err := usersubrouter.QuestionService.BookRepository.ListBooks()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to return all books: %v", err)
		return
	}
	logger.WithFunction().WithField("books", books).Info("return list of all available books")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"books": books,
	})
}

func writeError(w http.ResponseWriter, code int, formatstr string, args ...interface{}) {
	logger.WithFunction().Errorf(formatstr, args...)
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": fmt.Sprintf(formatstr, args...),
	})
}

func setCookie(w http.ResponseWriter, r *http.Request, cookieName, cookieValue string) {
	http.SetCookie(w, &http.Cookie{
		Name:    cookieName,
		Value:   cookieValue,
		Expires: time.Now().Add(365 * 24 * time.Hour),
		Path:    "/",
	})
}
