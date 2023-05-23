package handler

import (
	"github.com/VladimirMovsesyan/praktikum-gophermart/internal/auth"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func AuthMiddleware(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")

	login, err := auth.ParseToken(authHeader)
	if err != nil {
		log.Println("wrong authorization header provided")
		ctx.Writer.WriteHeader(http.StatusUnauthorized)
		ctx.Abort()
		return
	}

	ctx.Set("Login", login)
}
