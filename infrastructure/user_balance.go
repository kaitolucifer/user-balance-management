package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/kaitolucifer/user-balance-management/domain"
)

type userBalanceRepository struct {
	Conn DB
}

func NewUserBalanceRepository(db DB) domain.UserBalanceRepository {
	return &userBalanceRepository{db}
}

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
	num, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if num == 0 {
		// 更新するタイミングでユーザーが消滅した場合
		return errors.New("update failed")
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
	num, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if num == 0 {
		// 更新するタイミングでユーザーが消滅または減算後残高が負の場合
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
