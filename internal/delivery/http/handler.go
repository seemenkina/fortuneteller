package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"fortuneteller/internal/service"
	"github.com/gorilla/mux"
)

type UserSubrouter struct {
	mux.Router
	UserService *service.UserService
}

func HandlerMain(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "/Users/seemenkina/code/fortuneteller/assets/html/homepage.html")
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

	token, err := usersubrouter.UserService.Register(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("register error: %v", err),
		})
		return
	}

	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// w.WriteHeader()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"msg":      fmt.Sprintf("Hey, you are register %s, your token : %s\n", username, token),
		"redirect": "/register",
	})

	return
}

func (usersubrouter UserSubrouter) HandlerLoginPost(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"redirect": "/homepage",
	})
	return
}
