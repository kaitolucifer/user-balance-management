package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/kaitolucifer/user-balance-management/domain"
)

// userBalanceRepository DB接続を格納
type userBalanceRepository struct {
	Conn DB
	Tx   TX
}

// NewUserBalanceRepository 新しいrepositoryを作成
func NewUserBalanceRepository(db DB) domain.UserBalanceRepository {
	return &userBalanceRepository{Conn: db}
}

// GetCtxWithTimeout タイムアウト付きのコンテキストを取得
func (repo *userBalanceRepository) GetCtxWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// BeginTx トランザクションを開始
func (repo *userBalanceRepository) BeginTx(ctx context.Context) error {
	tx, err := repo.Conn.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	repo.Tx = TX{tx}
	return nil
}

// Commit トランザクションをコミット
func (repo *userBalanceRepository) Commit() error {
	return repo.Tx.Commit()
}

// Rollback トランザクションをロールバック
func (repo *userBalanceRepository) Rollback() error {
	return repo.Tx.Rollback()
}

// InsertTransactionHistory 取引履歴を挿入
func (repo *userBalanceRepository) InsertTransactionHistory(ctx context.Context, transactionID string, userID string, transactionType domain.TransactionType, amount int) error {
	var query string
	if userID == "" {
		query = `INSERT INTO transaction_history (transaction_id, transaction_type, amount, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)`
		_, err := repo.Tx.ExecContext(ctx, query, transactionID, transactionType, amount, time.Now(), time.Now())
		return err
	} else {
		query = `INSERT INTO transaction_history (transaction_id, user_id, transaction_type, amount, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6)`
		_, err := repo.Tx.ExecContext(ctx, query, transactionID, userID, transactionType, amount, time.Now(), time.Now())
		return err
	}
}

// QueryUserBalanceByUserID ユーザーIDでユーザー残高情報を取得
func (repo *userBalanceRepository) QueryUserBalanceByUserID(ctx context.Context, userID string) (domain.UserBalanceModel, error) {
	var userBalance domain.UserBalanceModel

	row := repo.Conn.DB.QueryRowContext(ctx, "SELECT * FROM user_balance WHERE user_id = $1", userID)
	err := row.Scan(
		&userBalance.UserID,
		&userBalance.Balance,
		&userBalance.CreatedAt,
		&userBalance.UpdatedAt,
	)

	return userBalance, err
}

// AddUserBalanceByUserID ユーザーIDでユーザー残高を加算
func (repo *userBalanceRepository) AddUserBalanceByUserID(ctx context.Context, userID string, amount int) error {
	if (repo.Tx == TX{nil}) {
		return errors.New("current thread is not associated with a transaction")
	}

	query := `UPDATE user_balance SET balance = balance + $1, updated_at = $2 WHERE user_id = $3`
	res, err := repo.Tx.ExecContext(ctx, query, amount, time.Now(), userID)
	if err != nil {
		return err
	}

	numRow, err := res.RowsAffected()
	if err != nil {
		return err
	} else if numRow == 0 {
		// 更新する時点でユーザーが存在しない場合
		return sql.ErrNoRows
	}

	return nil
}

// ReduceUserBalanceByUserID ユーザーIDでユーザー残高を減算
func (repo *userBalanceRepository) ReduceUserBalanceByUserID(ctx context.Context, userID string, amount int) error {
	if (repo.Tx == TX{nil}) {
		return errors.New("current thread is not associated with a transaction")
	}

	query := `UPDATE user_balance SET balance = balance - $1, updated_at = $2 WHERE user_id = $3 AND balance - $1 >= 0`
	res, err := repo.Tx.ExecContext(ctx, query, amount, time.Now(), userID)
	if err != nil {
		return err
	}

	numRow, err := res.RowsAffected()
	if err != nil {
		return err
	} else if numRow == 0 {
		// 更新する時点でユーザーが存在しないまたは減算後残高が負の場合
		return errors.New("update failed")
	}

	return nil
}

// AddAllUserBalance ユーザー残高を一斉に加算
func (repo *userBalanceRepository) AddAllUserBalance(ctx context.Context, amount int) error {
	if (repo.Tx == TX{nil}) {
		return errors.New("current thread is not associated with a transaction")
	}

	query := `UPDATE user_balance SET balance = balance + $1, updated_at = $2`
	_, err := repo.Tx.ExecContext(ctx, query, amount, time.Now())
	return err
}
