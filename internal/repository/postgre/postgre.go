package postgre

import (
	"context"
	"errors"
	"github.com/VladimirMovsesyan/praktikum-gophermart/internal/model"
	"github.com/jackc/pgx/v5"
	"log"
	"time"
)

const (
	userTable = `CREATE TABLE IF NOT EXISTS users (
    login VARCHAR(100) PRIMARY KEY,
    password VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL
);`

	orderTable = `CREATE TABLE IF NOT EXISTS orders (
    number VARCHAR(100) PRIMARY KEY,
	login VARCHAR(100) NOT NULL,
	status VARCHAR(15) NOT NULL,
	accrual DOUBLE PRECISION,
    uploaded_at TIMESTAMP NOT NULL,
	FOREIGN KEY (login) REFERENCES users (login)
);`

	withdrawTable = `CREATE TABLE IF NOT EXISTS withdrawals (
    number VARCHAR(100) NOT NULL,
	login VARCHAR(100) NOT NULL,
	sum DOUBLE PRECISION,
    processed_at TIMESTAMP NOT NULL,
	FOREIGN KEY (login) REFERENCES users (login)
);`
)

var (
	ErrorConflict = errors.New("error: login don't match")
	ErrorOk       = errors.New("error: already was uploaded")
)

type Storage struct {
	conn *pgx.Conn
}

func NewStorage(conn *pgx.Conn) (*Storage, error) {
	s := &Storage{
		conn: conn,
	}

	err := s.ensureTablesExist()
	if err != nil {
		return &Storage{}, err
	}

	return s, nil
}

func (s *Storage) ensureTablesExist() error {
	_, err := s.conn.Exec(context.Background(), userTable)
	if err != nil {
		return err
	}

	_, err = s.conn.Exec(context.Background(), orderTable)
	if err != nil {
		return err
	}

	_, err = s.conn.Exec(context.Background(), withdrawTable)
	return err
}

func (s *Storage) Create(user model.User) error {
	query := `INSERT INTO users VALUES ($1, $2, $3)`
	_, err := s.conn.Exec(context.Background(), query, user.Login, user.Password, time.Now())
	return err
}

func (s *Storage) UpdateOrder(login string, order model.Order) error {
	loginDB, err := s.GetOrderOwner(order.Number)
	if err != nil {
		log.Println(err)
		creationQuery := `INSERT INTO orders (number, login, status, uploaded_at) VALUES($1, $2, $3, $4)`

		_, err := s.conn.Exec(
			context.Background(),
			creationQuery,
			order.Number,
			login,
			"NEW",
			time.Now(),
		)
		return err
	}

	if login != loginDB {
		return ErrorConflict
	}

	if order.Status == "" {
		return ErrorOk
	}

	updateQuery := `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`

	_, err = s.conn.Exec(
		context.Background(),
		updateQuery,
		order.Status,
		order.Accrual,
		order.Number,
	)
	return err
}

func (s *Storage) GetOrderOwner(orderNum string) (login string, err error) {
	selectQuery := `SELECT (login) FROM orders WHERE number = $1`
	row := s.conn.QueryRow(context.Background(), selectQuery, orderNum)

	err = row.Scan(&login)
	if err != nil {
		return "", err
	}

	return login, nil
}

func (s *Storage) GetUser(login string) (model.User, error) {
	query := `SELECT (login, password) FROM users WHERE login = $1`
	row := s.conn.QueryRow(context.Background(), query, login)

	user := model.User{}

	err := row.Scan(&user)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (s *Storage) GetOrders(login string) ([]model.Order, error) {
	countQuery := `SELECT COUNT(*) FROM orders WHERE login = $1`
	selectQuery := `SELECT (number, status, accrual, uploaded_at) FROM orders WHERE login = $1`

	var cnt int
	row := s.conn.QueryRow(context.Background(), countQuery, login)

	err := row.Scan(&cnt)
	if err != nil {
		return nil, err
	}

	rows, err := s.conn.Query(context.Background(), selectQuery, login)
	if err != nil {
		return nil, err
	}

	orders := make([]model.Order, 0, cnt)

	for rows.Next() {
		var order model.Order

		err := rows.Scan(&order)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (s *Storage) GetProcessingOrders() ([]model.Order, error) {
	countQuery := `SELECT COUNT(*) FROM orders WHERE status IN ($1, $2)`
	selectQuery := `SELECT (number, status, accrual, uploaded_at) FROM orders WHERE status IN ($1, $2)`

	var cnt int
	row := s.conn.QueryRow(context.Background(), countQuery, model.OrderStatusNew, model.OrderStatusProcessing)

	err := row.Scan(&cnt)
	if err != nil {
		return nil, err
	}

	if cnt == 0 {
		return nil, errors.New("error: no content to return")
	}

	rows, err := s.conn.Query(context.Background(), selectQuery, model.OrderStatusNew, model.OrderStatusProcessing)
	if err != nil {
		return nil, err
	}

	orders := make([]model.Order, 0, cnt)

	for rows.Next() {
		var order model.Order

		err := rows.Scan(&order)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (s *Storage) Withdraw(login string, withdraw model.Withdraw) error {
	query := `INSERT INTO withdrawals VALUES($1, $2, $3, $4)`
	_, err := s.conn.Exec(
		context.Background(),
		query,
		withdraw.Order,
		login,
		withdraw.Sum,
		time.Now(),
	)

	return err
}

func (s *Storage) GetWithdrawals(login string) ([]model.Withdraw, error) {
	countQuery := `SELECT COUNT(*) FROM withdrawals WHERE login = $1`
	selectQuery := `SELECT (number, sum, processed_at) FROM withdrawals WHERE login = $1`

	var cnt int
	row := s.conn.QueryRow(context.Background(), countQuery, login)

	err := row.Scan(&cnt)
	if err != nil {
		return nil, err
	}

	withdrawals := make([]model.Withdraw, 0, cnt)

	rows, err := s.conn.Query(context.Background(), selectQuery, login)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var w model.Withdraw

		err := rows.Scan(&w)
		if err != nil {
			return nil, err
		}

		withdrawals = append(withdrawals, w)
	}

	return withdrawals, nil
}
