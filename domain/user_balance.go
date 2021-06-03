package domain

import (
	"time"
)

// UserBalanceModel user_balanceテーブルのデータモデル
type UserBalanceModel struct {
	UserID    string
	Balance   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TransactionHistoryModel transaction_historyテーブルのデータモデル
type TransactionHistoryModel struct {
	TransactionID   string
	UserID          string
	TransactionType TransactionType
	Amount          int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TransactionType 取引種類
type TransactionType int

const (
    TypeAddUserBalance TransactionType = iota
    TypeReduceUserBalance
	TypeAddAllUserBalance
)

// UserBalanceRepository ユーザー残高管理repositoryのインタフェース
type UserBalanceRepository interface {
	GetUserBalanceByUserID(string) (UserBalanceModel, error)
	AddUserBalanceByUserID(string, int, string) error
	ReduceUserBalanceByUserID(string, int, string) error
	AddAllUserBalance(int, string) error
}

// UserBalanceUsecase ユーザー残高管理usecaseのインタフェース
type UserBalanceUsecase interface {
	AddBalance(string, int, string) error
	ReduceBalance(string, int, string) error
	AddAllUserBalance(int, string) error
	GetBalance(string) (int, error)
}
