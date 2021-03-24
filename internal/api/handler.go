package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func HandlerHello(w http.ResponseWriter, req *http.Request) {
	args := mux.Vars(req)
	fmt.Fprintf(w, "Hello, %s!", args)
}
