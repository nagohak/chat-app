package auth

import "github.com/nagohak/chat-app/models"

type Auth interface {
	GeneratePassword(password string) (string, error)
	ComparePassword(password, hash string) (bool, error)
	CreateToken(user models.User) (string, error)
	ValidateToken(tokenString string) (models.User, error)
	NewUser(id string, name string) models.User
}

type auth struct{}

func NewAuth() Auth {
	return &auth{}
}
