package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"fortuneteller/internal/api"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/{name}", api.HandlerHello)
	router.HandleFunc("", nil)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
