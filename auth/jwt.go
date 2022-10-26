package auth

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/nagohak/chat-app/models"
)

const secret = "BFJWFwFtQgXL4JGE"
const expireTime = 604800 // one week

type Claims struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	jwt.StandardClaims
}

func (c *Claims) GetID() string {
	return c.ID
}

func (c *Claims) GetName() string {
	return c.Name
}

func CreateToken(user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Id":        user.GetID(),
		"Name":      user.GetName(),
		"ExpiresAt": time.Now().Unix() + expireTime,
	})
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString string) (models.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])

		}

		return []byte(secret), nil
	})

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}
