package main

import (
	"context"
	"encoding/hex"
	"expvar"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fortuneteller/internal/crypto"
	"fortuneteller/internal/data"
	fthttp "fortuneteller/internal/delivery/http"
	"fortuneteller/internal/logger"
	"fortuneteller/internal/repository"
	"fortuneteller/internal/service"

	"github.com/ardanlabs/conf"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {

	// =========================================================================
	// Configuration database

	logger.WithFunction().Infof("starting the database configuration")
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
				logger.WithFunction().Errorf("generating config usage: %v", err)
			}
			logger.WithFunction().Info(usage)
			return
		}
		logger.WithFunction().Errorf("parsing config : %v", err)
	}

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
	logger.WithFunction().Info("the database is succesfully configured")

	// =========================================================================
	// Configuration user interface

	logger.WithFunction().Info("starting the user service configuration")
	key := hex.EncodeToString([]byte("~ThisIsMagicKey~"))
	var userService = &service.UserService{
		UserRepository: repository.NewUserInterface(rawDBConn),
		Token: crypto.MumboJumbo{
			Key: []byte(key),
		},
	}
	logger.WithFunction().Info("the user service is succesfully configured")

	// =========================================================================
	// Configuration interface for question
	logger.WithFunction().Info("starting the question service configuration")

	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	logger.WithFunction().Info("executable directory for books interface: %s", dir)

	var questionService = &service.QuestionService{
		QuestionRepository: repository.NewQuestionInterface(rawDBConn),
		UserRepository:     userService.UserRepository,
		BookRepository:     repository.NewBookFileSystem(filepath.Join(dir, "books"), filepath.Join(dir, "books_keys")),
	}
	logger.WithFunction().Info("the question service is succesfully configured")

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

	apiRouter := router.PathPrefix("/api/v1/").Subrouter()

	apiRouter.HandleFunc("/auth/register", us.HandlerRegister).Methods("POST")
	apiRouter.HandleFunc("/auth/login", us.HandlerLogin).Methods("POST")

	apiRouter.HandleFunc("/users", us.HandlerListUsers).Methods("GET")
	apiRouter.HandleFunc("/users/questions", us.HandlerGetUserQuestions).Methods("GET")
	apiRouter.HandleFunc("/users/questions/answer", us.HandlerGetAnswerToQuestion).Methods("GET")
	apiRouter.HandleFunc("/users/questions/otherAnswer", us.HandlerGetAnswerToQuestionFromAnotherBook).Methods("GET")

	apiRouter.HandleFunc("/users/questions/ask", us.HandlerListBooks).Methods("GET")
	apiRouter.HandleFunc("/users/questions/ask", us.HandlerAskQuestion).Methods("POST")

	if err = http.ListenAndServe(":8080", router); err != nil {
		logger.WithFunction().WithError(err)
	}
}
