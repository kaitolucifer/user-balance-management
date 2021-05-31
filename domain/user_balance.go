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
	TransactionType TransactionType
	Amount          int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type TransactionType int

const (
    TypeAddUserBalance TransactionType = iota
    TypeReduceUserBalance
	TypeAddAllUserBalance
)

type UserBalanceRepository interface {
	GetUserBalanceByUserID(string) (UserBalanceModel, error)
	AddUserBalanceByUserID(string, int, string) error
	ReduceUserBalanceByUserID(string, int, string) error
	AddAllUserBalance(int, string) error
}

type UserBalanceUsecase interface {
	AddBalance(string, int, string) error
	ReduceBalance(string, int, string) error
	AddAllUserBalance(int, string) error
	GetBalance(string) (int, error)
}
