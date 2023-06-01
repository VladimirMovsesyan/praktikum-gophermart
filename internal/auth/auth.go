package auth

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

const (
	passKey  = "492gl12bACtAT1My"
	tokenKey = "Ce45L1mMzQw5w09z"
)

func HashPass(password string) string {
	h := sha1.New()
	h.Write([]byte(password))

	return fmt.Sprintf("%x", h.Sum([]byte(passKey)))
}

type claims struct {
	jwt.MapClaims
	Login string `json:"login"`
}

func GenerateToken(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Login: login,
	})

	signedToken, err := token.SignedString([]byte(tokenKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ParseToken(rawToken string) (string, error) {
	token, err := jwt.ParseWithClaims(rawToken, &claims{}, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("wrong signing method")
		}

		return []byte(tokenKey), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*claims)
	if !ok {
		return "", errors.New("wrong claims type provided")
	}

	return claims.Login, nil
}
