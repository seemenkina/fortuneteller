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
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("register post error: %v", err),
		})
		return
	}
	username := r.Form.Get("username")
	if len(username) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "username is empty",
		})
		return
	}

	token, err := usersubrouter.UserService.Register(r.Context(), username)
	log.Printf("user: %s token: %s", username, token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("register error: %v", err),
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "tokencookie",
		Value:   token,
		Expires: time.Now().Add(365 * 24 * time.Hour),
		Path:    "/",
	})
	cooki, err := r.Cookie("tokencookie")
	log.Printf("register tcookie: %v", cooki)
	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// w.WriteHeader()
	json.NewEncoder(w).Encode(map[string]interface{}{
		// "msg":      fmt.Sprintf("Hey, you are register %s, your token : %s\n", username, token),
		"redirect": "/register#" + token,
		"username": username,
	})

}

func (usersubrouter UserSubrouter) HandlerLoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("login post error: %v", err),
		})
		return
	}
	token := r.Form.Get("token")
	if len(token) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "token is empty",
		})
		return
	}
	log.Printf("token: %s\n", token)
	user, err := usersubrouter.UserService.Login(r.Context(), token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("login error: %v", err),
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "tokencookie",
		Value:   token,
		Expires: time.Now().Add(365 * 24 * time.Hour),
		Path:    "/",
	})
	cooki, err := r.Cookie("tokencookie")
	log.Printf("login tcookie: %v", cooki)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"redirect": "/homepage",
		"username": user.Username,
	})

}

func (usersubrouter UserSubrouter) HandlerListsUser(w http.ResponseWriter, r *http.Request) {
	users, err := usersubrouter.UserService.ListUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("list users error: %v", err),
		})
		return
	}
	// TODO: return last n usernames
	names := make([]string, len(users))
	for i, u := range users {
		names[i] = u.Username
	}

	// usernames := []string{"testUser", "alpha", "beta"}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": names,
	})
}

func (usersubrouter UserSubrouter) HandlerUserQuestionsGet(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("users questions error: %v", err),
		})
		return
	}
	name := r.Form.Get("username")
	log.Printf("USERNAME: %s", name)

	cookie, err := r.Cookie("tokencookie")
	if cookie.Value == "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("cookie is empty: %v", err),
		})
		return
	}
	cookiename, err := usersubrouter.UserService.Token.GetUsername(cookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("cant get username by token: %v", err),
		})
		return
	}
	log.Printf("USERNAME FROM COOKIE: %s", cookiename)

	var questions []data.Question
	if name == "" || cookiename == name {
		questions, err = usersubrouter.QuestionService.ListUserDecryptedQuestions(r.Context(), cookiename)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": fmt.Sprintf("list users error: %v", err),
			})
			return
		}
	} else {
		questions, err = usersubrouter.QuestionService.ListUserEncryptedQuestions(r.Context(), name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": fmt.Sprintf("list users error: %v", err),
			})
			return
		}
	}

	answer := make([]struct {
		Question string
		Answer   string
	}, len(questions))

	for i, q := range questions {
		answer[i].Answer = q.Answer
		answer[i].Question = q.Question
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"questions": answer,
	})

}

func (usersubrouter UserSubrouter) HandlerAnswerGet(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("tokencookie")
	if err != nil || cookie.Value == "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("cookie is empty: %v", err),
		})
		return
	}
	name, err := usersubrouter.UserService.Token.GetUsername(cookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("cant get username by token: %v", err),
		})
		return
	}

	log.Print(r.URL.Query().Get("id"))

	question, err := usersubrouter.QuestionService.FindUserQuestionByID(r.Context(),
		r.URL.Query().Get("id"), name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("cant get question by id: %v", err),
		})
		return
	}
	if question == (data.Question{}) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Question": "",
			"Answer":   "",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"Question": question.Question,
		"Answer":   question.Answer,
	})

}

func (usersubrouter UserSubrouter) HandlerAskQuestionPost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("tokencookie")
	if err != nil || cookie.Value == "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("cookie is empty: %v", err),
		})
		return
	}
	user, err := usersubrouter.UserService.Repo.FindUserByToken(r.Context(), cookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("cant get user by token: %v", err),
		})
		return
	}

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("ask post error: %v", err),
		})
		return
	}
	log.Printf(r.Form.Get("question"), r.Form.Get("book"), r.Form.Get("page"))

	row, err := strconv.Atoi(r.Form.Get("page"))
	bookData := data.FromAskData{
		Name: r.Form.Get("book"),
		Row:  row,
	}
	question, err := usersubrouter.QuestionService.AskQuestion(
		r.Context(), r.Form.Get("question")[9:], user, bookData)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("cant ask question: %v", err),
		})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"redirect": "/answer#" + question.ID,
	})

}

func (usersubrouter UserSubrouter) HandlerListBooks(w http.ResponseWriter, r *http.Request) {
	books, err := usersubrouter.QuestionService.Repob.ListBooks(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("list books error: %v", err),
		})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"books": books,
	})
}
