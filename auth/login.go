package auth

import "github.com/nagohak/chat-app/models"

type contextKey string

const UserContextKey = contextKey("user")

type NewUser struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (user *NewUser) GetID() string {
	return user.Id
}

func (user *NewUser) GetName() string {
	return user.Name
}

func (a *auth) NewUser(id string, name string) models.User {
	return &NewUser{Id: id, Name: name}
}
