package main

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"fortuneteller/internal/mocks"
	"fortuneteller/internal/service"
	"github.com/gorilla/mux"

	fthttp "fortuneteller/internal/delivery/http"
)

func main() {
	router := mux.NewRouter()

	var userService = &service.UserService{
		Repo:  make(mocks.UserMock),
		Token: mocks.TokenMock{},
	}
	us := fthttp.UserSubrouter{
		Router:      mux.Router{},
		UserService: userService,
	}

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./assets"))))
	router.HandleFunc("/", fthttp.HandlerMain)
	router.MatcherFunc(func(r *http.Request, match *mux.RouteMatch) bool {
		return !strings.HasPrefix(r.URL.Path, "/api")
	}).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("./assets/html", r.URL.Path[1:]+".html"))
	})

	apiRouter := router.PathPrefix("/api/v1/").Subrouter()

	apiRouter.HandleFunc("/auth/register", us.HandlerRegisterPost).Methods("POST")
	apiRouter.HandleFunc("/auth/login", us.HandlerLoginPost).Methods("POST")

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
