package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fortuneteller/internal/crypto"
	"fortuneteller/internal/data"
	"fortuneteller/internal/repository"
	"github.com/gorilla/mux"

	fthttp "fortuneteller/internal/delivery/http"
	"fortuneteller/internal/service"
)

func main() {
	exe, err := os.Executable()
	dir := filepath.Dir(exe)

	router := mux.NewRouter()
	// key := hex.EncodeToString([]byte("~ThisIsMagicKey~"))

	ctx := context.Background()

	rawDBConn, err := data.Open(ctx, data.Config{
		User:         "admin",
		Password:     "123456",
		Host:         "database",
		DatabaseName: "pgdb",
		DisableTLS:   "disable",
	})
	if err != nil {
		log.Printf("unable to connect db: %v", err)
		return
	}
	defer func() {
		log.Printf("database stopping")
		rawDBConn.Close()
	}()

	var userService = &service.UserService{
		Repo:  repository.NewUserInterface(rawDBConn),
		Token: crypto.MumboJumbo{},
	}
	var questionService = &service.QuestionService{
		Repoq: repository.NewQuestionInterface(rawDBConn),
		Repou: userService.Repo,
		Repob: repository.NewBookFileSystem(filepath.Join(dir, "books")),
	}

	us := fthttp.UserSubrouter{
		Router:          mux.Router{},
		UserService:     userService,
		QuestionService: questionService,
	}

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./assets"))))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/homepage", http.StatusFound)
	})

	router.MatcherFunc(func(r *http.Request, match *mux.RouteMatch) bool {
		return !strings.HasPrefix(r.URL.Path, "/api")
	}).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("url: %s", r.URL.Path)
		if r.URL.Path == "/cuteregister" {

		} else {
			// Read cookie
			c, err := r.Cookie("tokencookie")
			log.Printf("url: %s cookie: %v\n", r.URL.Path, c)
			if err != nil || c.Value == "" {
				http.Redirect(w, r, "/cuteregister", http.StatusFound)
				return
			}
		}
		http.ServeFile(w, r, filepath.Join("./assets/html", r.URL.Path[1:]+".html"))
	})

	apiRouter := router.PathPrefix("/api/v1/").Subrouter()

	apiRouter.HandleFunc("/auth/register", us.HandlerRegisterPost).Methods("POST")
	apiRouter.HandleFunc("/auth/login", us.HandlerLoginPost).Methods("POST")

	apiRouter.HandleFunc("/users", us.HandlerListsUser).Methods("GET")
	apiRouter.HandleFunc("/users/questions", us.HandlerUserQuestionsGet).Methods("GET")
	apiRouter.HandleFunc("/users/questions/answer", us.HandlerAnswerGet).Methods("GET")
	apiRouter.HandleFunc("/users/questions/otherAnswer", us.HandlerOtherAnswerGet).Methods("GET")
	apiRouter.HandleFunc("/users/questions/ask", us.HandlerAskQuestionPost).Methods("POST")
	apiRouter.HandleFunc("/users/questions/ask", us.HandlerListBooks).Methods("GET")

	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
