package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
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

func AuthMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, tok := r.URL.Query()["bearer"]
		name, nok := r.URL.Query()["name"]

		if tok && len(token) == 1 {
			user, err := auth.ValidateToken(token[0])
			if err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
			} else {
				ctx := context.WithValue(r.Context(), auth.UserContextKey, user)
				f(w, r.WithContext(ctx))
			}
		} else if nok && len(name) == 1 {
			user := auth.NewUser{Id: uuid.New().String(), Name: name[0]}
			ctx := context.WithValue(r.Context(), auth.UserContextKey, &user)
			f(w, r.WithContext(ctx))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Please login or provide name"))
		}
	}
}

func errorResponse(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte("{\"error\": \"" + msg + "\"}"))
}
