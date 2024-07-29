package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

var AuthenticationError = errors.New("invalid authentication")

func LoadUserIdFromToken(token string) (int, error) {
	tokenFromString, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return 0, AuthenticationError
		}

		return []byte(os.Getenv("HMAC")), nil
	})

	if err != nil {
		return 0, err
	}
	if claims, ok := tokenFromString.Claims.(jwt.MapClaims); ok {
		return int(claims["id"].(float64)), nil
	} else {
		return 0, AuthenticationError
	}
}

func GenerateToken(id int) (string, error) {
	hmac := os.Getenv("HMAC")
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"nbf": now.Unix(),
		"exp": now.Add(5 * 24 * time.Hour).Unix(),
		"iat": now.Unix(),
	})

	tokenString, err := token.SignedString([]byte(hmac))
	if err != nil {
		return "", err
	}
	return tokenString, nil

}
