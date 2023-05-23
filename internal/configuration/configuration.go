package configuration

import (
	"context"
	"errors"
	"github.com/VladimirMovsesyan/praktikum-gophermart/internal/handler"
	"github.com/VladimirMovsesyan/praktikum-gophermart/internal/model"
	"github.com/VladimirMovsesyan/praktikum-gophermart/internal/repository/postgre"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"net/http"
	"os"
)

const (
	envAddress    = "RUN_ADDRESS"
	envDsn        = "DATABASE_URI"
	envAccAddress = "ACCRUAL_SYSTEM_ADDRESS"
)

type repository interface {
	Create(user model.User) error
	UpdateOrder(login string, order model.Order) error
	GetOrderOwner(orderNum string) (login string, err error)
	GetUser(login string) (model.User, error)
	GetOrders(login string) ([]model.Order, error)
	GetProcessingOrders() ([]model.Order, error)
	Withdraw(login string, withdraw model.Withdraw) error
	GetWithdrawals(login string) ([]model.Withdraw, error)
}

type configuration struct {
	Address    string
	Dsn        string
	AccAddress string
	DB         *pgx.Conn
	Storage    repository
	Server     *http.Server
}

func NewConfiguration(flAddress, flDsn, flAccAddress *string) (configuration, error) {
	address, err := parseStringVar(flAddress, envAddress)
	if err != nil {
		return configuration{}, err
	}

	dsn, err := parseStringVar(flDsn, envDsn)
	if err != nil {
		return configuration{}, err
	}

	accAddress, err := parseStringVar(flAccAddress, envAccAddress)
	if err != nil {
		return configuration{}, err
	}

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return configuration{}, err
	}

	storage, err := postgre.NewStorage(conn)
	if err != nil {
		return configuration{}, err
	}

	gin.SetMode(gin.ReleaseMode)
	router := newRouter(storage)

	server := &http.Server{
		Addr:    address,
		Handler: router,
	}

	return configuration{
		Address:    address,
		Dsn:        dsn,
		AccAddress: accAddress,
		DB:         conn,
		Storage:    storage,
		Server:     server,
	}, nil
}

func parseStringVar(flag *string, envName string) (string, error) {
	if *flag != "" {
		return *flag, nil
	}

	value := os.Getenv(envName)
	if value == "" {
		return "", errors.New("not enough parameters to run service")
	}
	return value, nil
}

func newRouter(storage repository) *gin.Engine {
	router := gin.New()
	auth := router.Group("/api/user")
	{
		auth.POST("/register", handler.Register(storage))
		auth.POST("/login", handler.Login(storage))
	}

	api := router.Group("/api/user", handler.AuthMiddleware, handler.CompressMiddleware, handler.DecompressMiddleware)
	{
		api.POST("/orders", handler.UpdateOrder(storage))
		api.GET("/orders", handler.GetOrders(storage))
		api.GET("/balance", handler.GetBalance(storage))
		api.POST("/balance/withdraw", handler.Withdraw(storage))
		api.GET("/withdrawals", handler.GetWithdrawals(storage))
	}

	return router
}
