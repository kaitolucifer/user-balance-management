package domain

import (
	"context"
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
	TransactionType_AddUserBalance TransactionType = iota
	TransactionType_ReduceUserBalance
	TransactionType_AddAllUserBalance
)

// UserBalanceRepository ユーザー残高管理repositoryのインタフェース
type UserBalanceRepository interface {
	GetCtxWithTimeout(time.Duration) (context.Context, context.CancelFunc)
	BeginTx(context.Context) error
	Commit() error
	Rollback() error
	InsertTransactionHistory(context.Context, string, string, TransactionType, int) error
	QueryUserBalanceByUserID(context.Context, string) (UserBalanceModel, error)
	AddUserBalanceByUserID(context.Context, string, int) error 
	ReduceUserBalanceByUserID(context.Context, string, int) error
	AddAllUserBalance(context.Context, int) error
}

// UserBalanceUsecase ユーザー残高管理usecaseのインタフェース
type UserBalanceUsecase interface {
	AddBalance(string, int, string) error
	ReduceBalance(string, int, string) error
	AddAllUserBalance(int, string) error
	GetBalance(string) (int, error)
}
