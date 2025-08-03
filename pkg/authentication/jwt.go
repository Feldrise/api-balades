package authentication

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

func GenerateToken(secret string, id uint) (string, error) {
	secretKey := []byte(secret)
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = id
	claims["exp"] = time.Now().Add(time.Hour * 24 * 365).Unix()

	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(secret string, tokenString string) (uint, error) {
	secretKey := []byte(secret)
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["id"] == nil || claims["id"] == "" {
			return 0, errors.New("invalid token")
		}

		id := claims["id"].(float64)

		return uint(id), nil
	} else {
		return 0, err
	}
}
