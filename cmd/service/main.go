package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fortuneteller/internal/crypto"
	"fortuneteller/internal/data"
	fthttp "fortuneteller/internal/delivery/http"
	"fortuneteller/internal/repository"
	"fortuneteller/internal/service"

	"github.com/ardanlabs/conf"
	"github.com/gorilla/mux"
)

func main() {
	// =========================================================================
	// Configuration database
	log.Printf("START CONFIGURATE DATABASE")
	ctx := context.Background()

	var cfg struct {
		DB struct {
			User         string `conf:"default:admin"`
			Password     string `conf:"default:123456,noprint"`
			Host         string `conf:"default:database"`
			DatabaseName string `conf:"default:pgdb"`
			DisableTLS   string `conf:"default:disable"`
		}
	}

	if err := conf.Parse(os.Args[1:], "FORTUNETELLER", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("FORTUNETELLER", &cfg)
			if err != nil {
				log.Printf("generating config usage: %v", err)
			}
			fmt.Println(usage)
			return
		}
		log.Printf("error: parsing config : %v", err)
	}

	rawDBConn, err := data.Open(ctx, data.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		DatabaseName: cfg.DB.DatabaseName,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		log.Printf("unable to connect db: %v", err)
		return
	}
	defer func() {
		log.Printf("database stopping")
		rawDBConn.Close()
	}()
	log.Printf("SUCCES CONFIGURATE DATABASE")

	// =========================================================================
	// Configuration user interface

	log.Printf("START CONFIGURATE USER SERVICE")
	key := hex.EncodeToString([]byte("~ThisIsMagicKey~"))
	var userService = &service.UserService{
		Repo: repository.NewUserInterface(rawDBConn),
		Token: crypto.MumboJumbo{
			Key: []byte(key),
		},
	}
	log.Printf("SUCCES CONFIGURATE USER SERVICE")

	// =========================================================================
	// Configuration interface for question

	exe, err := os.Executable()
	dir := filepath.Dir(exe)
	log.Printf("EXECUTABLE DIR: %s", dir)

	log.Printf("START CONFIGURATE QUESTION SERVICE")
	var questionService = &service.QuestionService{
		Repoq: repository.NewQuestionInterface(rawDBConn),
		Repou: userService.Repo,
		Repob: repository.NewBookFileSystem(filepath.Join(dir, "books"), filepath.Join(dir, "books_keys")),
	}
	log.Printf("SUCCES CONFIGURATE QUESTION SERVICE")

	// =========================================================================
	// Configuration router and subrouter

	us := fthttp.UserSubrouter{
		Router:          mux.Router{},
		UserService:     userService,
		QuestionService: questionService,
	}

	router := mux.NewRouter()
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
	apiRouter.HandleFunc("/users/questions/ask", us.HandlerListBooks).Methods("GET")
	apiRouter.HandleFunc("/users/questions/ask", us.HandlerAskQuestionPost).Methods("POST")

	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
