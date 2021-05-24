package main

import (
	"context"
	"errors"
	"expvar"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ardanlabs/conf"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"fortuneteller/internal/crypto"
	"fortuneteller/internal/data"
	fthttp "fortuneteller/internal/delivery/http"
	"fortuneteller/internal/logger"
	"fortuneteller/internal/repository"
	"fortuneteller/internal/service"
)

func main() {
	var cfg struct {
		DB struct {
			User         string `conf:"default:admin"`
			Password     string `conf:"default:123456,noprint"`
			Host         string `conf:"default:database"`
			DatabaseName string `conf:"default:pgdb"`
			DisableTLS   string `conf:"default:disable"`
		}
		TokenKey string `conf:"noprint,flag:token"`
	}

	err := conf.Parse(os.Args[1:], "FORTUNETELLER", &cfg)
	switch {
	case err == nil:
		// pass
	case errors.Is(err, conf.ErrHelpWanted):
		if usage, err := conf.Usage("FORTUNETELLER", &cfg); err == nil {
			println(usage)
		} else {
			logger.WithFunction().Fatalf("Can't generate config usage: %v", err)
		}
	default:
		logger.WithFunction().Fatalf("Can't parse config: %v", err)
	}

	// =========================================================================
	// Configuration database

	logger.WithFunction().Infof("starting the database configuration")
	ctx := context.Background()

	rawDBConn, err := data.Open(ctx, data.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		DatabaseName: cfg.DB.DatabaseName,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		logger.WithFunction().Errorf("unable to connect db: %v", err)
		return
	}
	defer func() {
		logger.WithFunction().Info("database stopping")
		rawDBConn.Close()
	}()
	logger.WithFunction().Info("the database is successfully configured")

	// =========================================================================
	// Configuration user interface

	logger.WithFunction().Info("starting the user service configuration")
	var userService = &service.UserService{
		UserRepository: repository.NewUserInterface(rawDBConn),
		Token: crypto.MumboJumbo{
			Key: []byte(cfg.TokenKey),
		},
	}
	logger.WithFunction().Info("the user service is successfully configured")

	// =========================================================================
	// Configuration interface for question
	logger.WithFunction().Info("starting the question service configuration")

	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	logger.WithFunction().Infof("executable directory for books interface: %s", dir)

	var questionService = &service.QuestionService{
		QuestionRepository: repository.NewQuestionInterface(rawDBConn),
		UserRepository:     userService.UserRepository,
		BookRepository:     repository.NewBookFileSystem(filepath.Join(dir, "books"), filepath.Join(dir, "books_keys")),
	}
	logger.WithFunction().Info("the question service is successfully configured")

	// =========================================================================
	// Configuration router and subrouter

	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./assets"))))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/homepage", http.StatusFound)
	})
	router.Handle("/stats/", expvar.Handler())

	router.MatcherFunc(func(r *http.Request, match *mux.RouteMatch) bool {
		return !strings.HasPrefix(r.URL.Path, "/api")
	}).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// redirect to register page if no cookies are set
		if r.URL.Path != "/cuteregister" {

			// Read cookie
			c, err := r.Cookie("tokencookie")
			logger.WithFunction().WithFields(logrus.Fields{
				"url":    r.URL.Path,
				"cookie": c,
			})
			if err != nil || c.Value == "" {
				http.Redirect(w, r, "/cuteregister", http.StatusFound)
				return
			}
		}
		// redirect to correspond html page
		http.ServeFile(w, r, filepath.Join("./assets/html", r.URL.Path[1:]+".html"))
	})

	us := fthttp.UserSubrouter{
		UserService:     userService,
		QuestionService: questionService,
	}
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.HandleFunc("/auth/register", us.Register).Methods("POST")
	apiRouter.HandleFunc("/auth/login", us.Login).Methods("POST")

	users := apiRouter.PathPrefix("/users").Subrouter()
	users.Use(fthttp.AuthenticatedUser)
	users.HandleFunc("", us.ListUsers).Methods("GET")
	users.HandleFunc("/questions", us.GetUserQuestions).Methods("GET")

	users.HandleFunc("/questions/answer", us.GetAnswerFromBook).Methods("GET")
	users.HandleFunc("/questions/otherAnswer", us.GetAnswerFromAnotherBook).Methods("GET")
	users.HandleFunc("/questions/ask", us.ListBooks).Methods("GET")
	users.HandleFunc("/questions/ask", us.AskQuestion).Methods("POST")

	if err = http.ListenAndServe(":8080", router); err != nil {
		logger.WithFunction().WithError(err)
	}
}
