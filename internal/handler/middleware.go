package handler

import (
	"compress/gzip"
	"github.com/VladimirMovsesyan/praktikum-gophermart/internal/auth"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"strings"
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

type gzipWriter struct {
	gin.ResponseWriter
	writer io.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	g.Header().Del("Content-Length")
	return g.writer.Write(data)
}

func CompressMiddleware(ctx *gin.Context) {
	if !strings.Contains(ctx.Request.Header.Get("Accept-Encoding"), "gzip") {
		ctx.Next()
		return
	}

	gWriter := gzip.NewWriter(ctx.Writer)
	defer gWriter.Close()

	ctx.Writer = &gzipWriter{
		ResponseWriter: ctx.Writer,
		writer:         gWriter,
	}

	ctx.Writer.Header().Set("Content-Encoding", "gzip")
	ctx.Next()
}

func DecompressMiddleware(ctx *gin.Context) {
	if !strings.Contains(ctx.Request.Header.Get("Content-Encoding"), "gzip") {
		ctx.Next()
		return
	}

	gReader, err := gzip.NewReader(ctx.Request.Body)
	if err != nil {
		log.Println(err)
		ctx.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer gReader.Close()

	ctx.Request.Body = gReader
	ctx.Next()
}
