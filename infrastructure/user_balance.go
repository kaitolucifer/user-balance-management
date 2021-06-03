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
}

// NewUserBalanceRepository 新しいrepositoryを作成
func NewUserBalanceRepository(db DB) domain.UserBalanceRepository {
	return &userBalanceRepository{db}
}

// GetUserBalanceByUserID ユーザーIDでユーザー残高情報を取得
func (repo *userBalanceRepository) GetUserBalanceByUserID(userID string) (domain.UserBalanceModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var userBalance domain.UserBalanceModel

	row := repo.Conn.DB.QueryRowContext(ctx, "SELECT * FROM user_balance WHERE user_id = $1", userID)
	err := row.Scan(
		&userBalance.UserID,
		&userBalance.Balance,
		&userBalance.CreatedAt,
		&userBalance.UpdatedAt,
	)
	if err != nil {
		return userBalance, err
	}

	return userBalance, nil
}

// AddUserBalanceByUserID ユーザーIDでユーザー残高を加算
func (repo *userBalanceRepository) AddUserBalanceByUserID(userID string, amount int, transactionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := repo.Conn.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
		}
	}()

	update_query := `UPDATE user_balance SET balance = balance + $1, updated_at = $2 WHERE user_id = $3`
	res, err := tx.ExecContext(ctx, update_query, amount, time.Now(), userID)
	if err != nil {
		tx.Rollback()
		return err
	}
	numRow, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if numRow == 0 {
		// 更新する時点でユーザーが存在しない場合
		return sql.ErrNoRows
	}

	insert_query := `INSERT INTO transaction_history (transaction_id, user_id, transaction_type, amount, created_at, updated_at)
					 VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = tx.ExecContext(ctx, insert_query,
		transactionID,
		userID,
		domain.TypeAddUserBalance,
		amount,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()

	return err
}

// ReduceUserBalanceByUserID ユーザーIDでユーザー残高を減算
func (repo *userBalanceRepository) ReduceUserBalanceByUserID(userID string, amount int, transactionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := repo.Conn.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
		}
	}()

	update_query := `UPDATE user_balance SET balance = balance - $1, updated_at = $2 WHERE user_id = $3 AND balance - $1 > 0`
	res, err := tx.ExecContext(ctx, update_query, amount, time.Now(), userID)
	if err != nil {
		tx.Rollback()
		return err
	}
	numRow, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if numRow == 0 {
		// 更新する時点でユーザーが存在しないまたは減算後残高が負の場合
		return errors.New("update failed")
	}

	insert_query := `INSERT INTO transaction_history (transaction_id, user_id, transaction_type, amount, created_at, updated_at)
					 VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = tx.ExecContext(ctx, insert_query,
		transactionID,
		userID,
		domain.TypeReduceUserBalance,
		amount,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()

	return err
}

// AddAllUserBalance ユーザー残高を一斉に加算
func (repo *userBalanceRepository) AddAllUserBalance(amount int, transactionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := repo.Conn.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
		}
	}()

	update_query := `UPDATE user_balance SET balance = balance + $1, updated_at = $2`
	_, err = tx.ExecContext(ctx, update_query, amount, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	insert_query := `INSERT INTO transaction_history (transaction_id, transaction_type, amount, created_at, updated_at)
					 VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.ExecContext(ctx, insert_query,
		transactionID,
		domain.TypeAddAllUserBalance,
		amount,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()

	return err
}
