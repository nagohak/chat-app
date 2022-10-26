package main

import (
	"encoding/json"
	"net/http"

	"github.com/nagohak/chat-app/auth"
	"github.com/nagohak/chat-app/repository"
)

type LoginUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Api struct {
	UserRepository *repository.UserRepository
}

func (api *Api) Login(w http.ResponseWriter, r *http.Request) {
	var user LoginUser

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		errorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbUser, err := api.UserRepository.FindUserByUsername(user.Username)
	if err != nil {
		errorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if dbUser == nil {
		errorResponse(w, "Login failed", http.StatusForbidden)
		return
	}

	ok, err := auth.ComparePassword(user.Password, dbUser.Password)
	if !ok || err != nil {
		errorResponse(w, "Login failed", http.StatusForbidden)
		return
	}

	token, err := auth.CreateToken(dbUser)
	if err != nil {
		errorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(token))
}

func errorResponse(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte("{\"error\": \"" + msg + "\"}"))
}
