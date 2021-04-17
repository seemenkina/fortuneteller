package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"fortuneteller/internal/data"
	"fortuneteller/internal/service"
	"github.com/gorilla/mux"
)

type UserSubrouter struct {
	mux.Router
	UserService     *service.UserService
	QuestionService *service.QuestionService
}

func (usersubrouter UserSubrouter) HandlerRegisterPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse register form: %v", err)
		return
	}
	username := r.Form.Get("username")
	if len(username) == 0 {
		writeError(w, http.StatusBadRequest, "username is empty")
		return
	}

	token, err := usersubrouter.UserService.Register(r.Context(), username)
	log.Printf("user: %s token: %s", username, token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "can't register user: %v", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "tokencookie",
		Value:   token,
		Expires: time.Now().Add(365 * 24 * time.Hour),
		Path:    "/",
	})
	tokencookie, _ := r.Cookie("tokencookie")
	log.Printf("register token from cookie: %v", tokencookie)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"redirect": "/register#" + token,
		"username": username,
	})

}

func (usersubrouter UserSubrouter) HandlerLoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse login form: %v", err)
		return
	}

	token := r.Form.Get("token")
	if len(token) == 0 {
		writeError(w, http.StatusBadRequest, "token is empty")
		return
	}
	log.Printf("token: %s\n", token)
	user, err := usersubrouter.UserService.Login(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "can't login user: %v", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "tokencookie",
		Value:   token,
		Expires: time.Now().Add(365 * 24 * time.Hour),
		Path:    "/",
	})
	tokencookie, _ := r.Cookie("tokencookie")
	log.Printf("login token from cookie: %v", tokencookie)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"redirect": "/homepage",
		"username": user.Username,
	})

}

func (usersubrouter UserSubrouter) HandlerListsUser(w http.ResponseWriter, r *http.Request) {
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

func (usersubrouter UserSubrouter) HandlerUserQuestionsGet(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse user questions form: %v", err)
		return
	}
	username := r.Form.Get("username")
	log.Printf("USERNAME: %s", username)

	cookie, err := r.Cookie("tokencookie")
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusBadRequest, "cookie is empty: %v", err)
		return
	}

	usernameFromCookie, err := usersubrouter.UserService.Token.GetUsername(cookie.Value)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to get username by token: %v", err)
		return
	}
	log.Printf("USERNAME FROM COOKIE: %s", usernameFromCookie)

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
	//
	// questionData := make([]struct {
	// 	Id string
	// 	Answer   string
	// 	Question string
	// }, len(questions))
	//
	// for i, q := range questions {
	// 	questionData[i].Id = q.ID
	// 	questionData[i].Answer = q.Answer
	// 	questionData[i].Question = q.Question
	// }

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"questions": questions,
	})

}

func (usersubrouter UserSubrouter) HandlerAnswerGet(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("tokencookie")
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusBadRequest, "cookie is empty: %v", err)
		return
	}

	usernameFromCookie, err := usersubrouter.UserService.Token.GetUsername(cookie.Value)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to get username by token: %v", err)
		return
	}

	log.Printf("QUESTION ID : %v", r.URL.Query().Get("id"))

	question, err := usersubrouter.QuestionService.FindUserQuestionByID(r.Context(),
		r.URL.Query().Get("id"), usernameFromCookie)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to get question by id: %v", err)
		return
	}

	books, err := usersubrouter.QuestionService.Repob.ListBooks()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to return all books: %v", err)
		return
	}
	booksData := make([]struct {
		Name string
		Id   int
	}, len(books))
	for i, b := range books {
		booksData[i].Name = b.Name
		booksData[i].Id = i + 1
	}

	if question == (data.Question{}) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"Question": "",
			"Answer":   "",
			"Books":    booksData,
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"Question": question.Question,
		"Answer":   question.Answer,
		"Books":    booksData,
	})

}

func (usersubrouter UserSubrouter) HandlerOtherAnswerGet(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("tokencookie")
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusBadRequest, "cookie is empty: %v", err)
		return
	}

	usernameFromCookie, err := usersubrouter.UserService.Token.GetUsername(cookie.Value)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to get username by token: %v", err)
		return
	}

	log.Printf("OTHER QUESTION ID : %v, %v", r.URL.Query().Get("id"), r.URL.Query().Get("id_book"))

	question, err := usersubrouter.QuestionService.FindUserQuestionByID(r.Context(),
		r.URL.Query().Get("id"), usernameFromCookie)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to get question by id: %v", err)
		return
	}

	log.Print("OTHER "+r.Form.Get("id_book"), r.Form.Get("book"))

	otherAnswer, err := usersubrouter.QuestionService.AskQuestionFromAnotherBook(r.Context(),
		question, usernameFromCookie, r.URL.Query().Get("id_book"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to get question answer from another book: %v", err)
		return
	}

	if otherAnswer == (data.Question{}) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"Question": "",
			"Answer":   "",
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"Question": otherAnswer.Question,
		"Answer":   otherAnswer.Answer,
	})

}

func (usersubrouter UserSubrouter) HandlerAskQuestionPost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("tokencookie")
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusBadRequest, "cookie is empty: %v", err)
		return
	}

	user, err := usersubrouter.UserService.Repo.FindUserByToken(r.Context(), cookie.Value)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to get user by token: %v", err)
		return
	}

	if err = r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse ask questions form: %v", err)
		return
	}
	log.Print(r.Form.Get("question"), r.Form.Get("book"), r.Form.Get("page"))

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

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"redirect": "/answer#" + question.ID,
	})
}

func (usersubrouter UserSubrouter) HandlerListBooks(w http.ResponseWriter, r *http.Request) {
	books, err := usersubrouter.QuestionService.Repob.ListBooks()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to return all books: %v", err)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"books": books,
	})
}

func writeError(w http.ResponseWriter, code int, formatstr string, args ...interface{}) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": fmt.Sprintf(formatstr, args...),
	})
}
