package domain

import (
	"time"
)

type UserBalanceModel struct {
	UserID    string
	Balance   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TransactionHistoryModel struct {
	TransactionID   string
	UserID          string
	TransactionType int
	Amount          int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type UserBalanceRepository interface {
	GetUserBalanceByUserID(string) (UserBalanceModel, error)
	AddUserBalanceByUserID(string, int) error
}

type UserBalanceUsecase interface {
	AddBalance(int) error
	GetBalance(string) (int, error)
}
