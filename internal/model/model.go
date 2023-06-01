package model

import (
	"time"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

const (
	OrderStatusNew        = "NEW"
	OrderStatusRegistered = "REGISTERED"
	OrderStatusProcessing = "PROCESSING"
)

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Withdraw struct {
	Order       string     `json:"order"`
	Sum         float64    `json:"sum"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}
