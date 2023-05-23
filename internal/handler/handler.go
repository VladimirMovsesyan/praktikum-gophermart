package handler

import (
	"encoding/json"
	"errors"
	"github.com/VladimirMovsesyan/praktikum-gophermart/internal/auth"
	"github.com/VladimirMovsesyan/praktikum-gophermart/internal/model"
	"github.com/VladimirMovsesyan/praktikum-gophermart/internal/repository/postgre"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"strconv"
)

type repository interface {
	Create(user model.User) error
	UpdateOrder(login string, order model.Order) error
	GetUser(login string) (model.User, error)
	GetOrders(login string) ([]model.Order, error)
	Withdraw(login string, withdraw model.Withdraw) error
	GetWithdrawals(login string) ([]model.Withdraw, error)
}

func Register(storage repository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		bytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		user := model.User{}
		err = json.Unmarshal(bytes, &user)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		user.Password = auth.HashPass(user.Password)

		_, err = storage.GetUser(user.Login)
		if err == nil {
			log.Println("Error: user with same login already exist")
			ctx.Writer.WriteHeader(http.StatusConflict)
			return
		}

		token, err := auth.GenerateToken(user.Login)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = storage.Create(user)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx.Writer.Header().Add("Authorization", token)
		ctx.Writer.WriteHeader(http.StatusOK)
	}
}

func Login(storage repository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		bytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		user := model.User{}
		err = json.Unmarshal(bytes, &user)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		user.Password = auth.HashPass(user.Password)

		userDb, err := storage.GetUser(user.Login)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user.Password != userDb.Password {
			log.Println("Error: wrong login/password passed")
			ctx.Writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, err := auth.GenerateToken(user.Login)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx.Writer.Header().Add("Authorization", token)
		ctx.Writer.WriteHeader(http.StatusOK)
	}
}

func luhnValidation(number int) bool {
	return (number%10+checksum(number/10))%10 == 0
}

func checksum(number int) int {
	var luhn int

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 { // even
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}

func UpdateOrder(storage repository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		bytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		number := string(bytes)
		num, err := strconv.Atoi(number)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		if !luhnValidation(num) {
			log.Printf("Luhn's validation of value: %s failed", number)
			ctx.Writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		l, ok := ctx.Get("Login")
		if !ok {
			log.Println("Couldn't get Login value from context")
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		login, ok := l.(string)
		if !ok {
			log.Println("Couldn't transform Login value from context")
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		order := model.Order{
			Number: number,
		}

		err = storage.UpdateOrder(login, order)
		if err != nil {
			log.Println(err)
			if errors.Is(err, postgre.ErrorConflict) {
				ctx.Writer.WriteHeader(http.StatusConflict)
				return
			}
			ctx.Writer.WriteHeader(http.StatusOK)
			return
		}

		ctx.Writer.WriteHeader(http.StatusAccepted)
	}
}

func GetOrders(storage repository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		l, ok := ctx.Get("Login")
		if !ok {
			log.Println("Couldn't get Login value from context")
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		login, ok := l.(string)
		if !ok {
			log.Println("Couldn't transform Login value from context")
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		batch, err := storage.GetOrders(login)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(batch) == 0 {
			log.Println("Error: no data to response")
			ctx.Writer.WriteHeader(http.StatusNoContent)
			return
		}

		bytes, err := json.Marshal(batch)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx.Writer.Header().Add("Content-Type", "application/json")

		_, err = ctx.Writer.Write(bytes)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx.Writer.WriteHeader(http.StatusOK)
	}
}

type balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func calcBalance(storage repository, login string) (balance, error) {
	orders, err := storage.GetOrders(login)
	if err != nil {
		return balance{}, err
	}

	var bal balance

	for _, order := range orders {
		if order.Accrual != nil {
			bal.Current += *order.Accrual
		}
	}

	withdrawals, err := storage.GetWithdrawals(login)
	if err != nil {
		return balance{}, err
	}

	for _, w := range withdrawals {
		bal.Withdrawn += w.Sum
	}

	bal.Current -= bal.Withdrawn

	return bal, nil
}

func GetBalance(storage repository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		l, ok := ctx.Get("Login")
		if !ok {
			log.Println("Couldn't get Login value from context")
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		login, ok := l.(string)
		if !ok {
			log.Println("Couldn't transform Login value from context")
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		bal, err := calcBalance(storage, login)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		bytes, err := json.Marshal(&bal)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx.Writer.Header().Add("Content-Type", "application/json")

		_, err = ctx.Writer.Write(bytes)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx.Writer.WriteHeader(http.StatusOK)
	}
}

func Withdraw(storage repository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		l, ok := ctx.Get("Login")
		if !ok {
			log.Println("Couldn't get Login value from context")
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		login, ok := l.(string)
		if !ok {
			log.Println("Couldn't transform Login value from context")
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		bytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		var w model.Withdraw
		err = json.Unmarshal(bytes, &w)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusBadRequest)
			return
		}

		bal, err := calcBalance(storage, login)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if bal.Current < w.Sum {
			log.Println("Error: withdrawal sum is bigger than current balance")
			ctx.Writer.WriteHeader(http.StatusPaymentRequired)
			return
		}

		orderNum, err := strconv.Atoi(w.Order)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !luhnValidation(orderNum) {
			log.Printf("Luhn's validation of value: %s failed", w.Order)
			ctx.Writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		err = storage.Withdraw(login, w)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx.Writer.WriteHeader(http.StatusOK)
	}
}

func GetWithdrawals(storage repository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		l, ok := ctx.Get("Login")
		if !ok {
			log.Println("Couldn't get Login value from context")
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		login, ok := l.(string)
		if !ok {
			log.Println("Couldn't transform Login value from context")
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		withdrawals, err := storage.GetWithdrawals(login)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(withdrawals) == 0 {
			log.Println("Error: no data to response")
			ctx.Writer.WriteHeader(http.StatusNoContent)
			return
		}

		bytes, err := json.Marshal(withdrawals)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx.Writer.Header().Add("Content-Type", "application/json")

		_, err = ctx.Writer.Write(bytes)
		if err != nil {
			log.Println(err)
			ctx.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx.Writer.WriteHeader(http.StatusOK)
	}
}
