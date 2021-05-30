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
	GetUserBalanceByUserId(int) (UserBalanceModel, error)
}

type UserBalanceUsecase interface {
	AddBalance(int) error
}
