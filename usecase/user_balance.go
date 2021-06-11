package usecase

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgconn"
	"github.com/kaitolucifer/user-balance-management/domain"
)

// userBalanceUsecase repositoryを格納
type userBalanceUsecase struct {
	repo domain.UserBalanceRepository
}

// NewUserBalanceUsecase 新しいusecaseを作成
func NewUserBalanceUsecase(repo domain.UserBalanceRepository) domain.UserBalanceUsecase {
	return &userBalanceUsecase{
		repo: repo,
	}
}

// AddBalance ユーザーIDでユーザー残高を加算
func (u *userBalanceUsecase) AddBalance(userID string, amount int, transactionID string) error {
	ctx, cancel := u.repo.GetCtxWithTimeout(3 * time.Second)
	defer cancel()
	if err := u.repo.BeginTx(ctx); err != nil {
		return errors.New("database error")
	}

	err := u.repo.AddUserBalanceByUserID(ctx, userID, amount)
	if err != nil {
		if err := u.repo.Rollback(); err != nil {
			return errors.New("database error")
		}

		var pgErr *pgconn.PgError
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		} else if errors.As(err, &pgErr) {
			return errors.New("database error")
		}

		return err
	}

	err = u.repo.InsertTransactionHistory(ctx, transactionID, userID, domain.TransactionType_AddUserBalance, amount)
	if err != nil {
		if err := u.repo.Rollback(); err != nil {
			return errors.New("database error")
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return errors.New("transaction_id must be unique")
			default:
				return errors.New("database error")
			}
		}

		return err
	}

	if err := u.repo.Commit(); err != nil {
		return errors.New("database error")
	}

	return nil
}

// ReduceBalance ユーザーIDでユーザー残高を減算
func (u *userBalanceUsecase) ReduceBalance(userID string, amount int, transactionID string) error {
	ctx, cancel := u.repo.GetCtxWithTimeout(3 * time.Second)
	defer cancel()

	userBalance, err := u.repo.QueryUserBalanceByUserID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		return err
	}

	if userBalance.Balance-amount < 0 {
		return errors.New("balance insufficient")
	}

	if err := u.repo.BeginTx(ctx); err != nil {
		return errors.New("database error")
	}
	
	err = u.repo.ReduceUserBalanceByUserID(ctx, userID, amount)
	if err != nil {
		if err := u.repo.Rollback(); err != nil {
			return errors.New("database error")
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return errors.New("database error")
		}

		return err
	}

	err = u.repo.InsertTransactionHistory(ctx, transactionID, userID, domain.TransactionType_ReduceUserBalance, amount)
	if err != nil {
		if err := u.repo.Rollback(); err != nil {
			return errors.New("database error")
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return errors.New("transaction_id must be unique")
			default:
				return errors.New("database error")
			}
		}

		return err
	}

	if err := u.repo.Commit(); err != nil {
		return errors.New("database error")
	}

	return nil
}

// ユーザー残高を一斉に加算
func (u *userBalanceUsecase) AddAllUserBalance(amount int, transactionID string) error {
	ctx, cancel := u.repo.GetCtxWithTimeout(3 * time.Second)
	defer cancel()
	if err := u.repo.BeginTx(ctx); err != nil {
		return errors.New("database error")
	}

	err := u.repo.AddAllUserBalance(ctx, amount)
	if err != nil {
		if err := u.repo.Rollback(); err != nil {
			return errors.New("database error")
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return errors.New("database error")
		}

		return err
	}

	err = u.repo.InsertTransactionHistory(ctx, transactionID, "", domain.TransactionType_AddAllUserBalance, amount)
	if err != nil {
		if err := u.repo.Rollback(); err != nil {
			return errors.New("database error")
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return errors.New("transaction_id must be unique")
			default:
				return errors.New("database error")
			}
		}

		return err
	}

	if err := u.repo.Commit(); err != nil {
		return errors.New("database error")
	}

	return nil
}

func (u *userBalanceUsecase) GetBalance(userID string) (int, error) {
	ctx, cancel := u.repo.GetCtxWithTimeout(3 * time.Second)
	defer cancel()
	
	userBalance, err := u.repo.QueryUserBalanceByUserID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("user not found")
		}
		
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return 0, errors.New("database error")
		}

		return 0, err
	}

	return userBalance.Balance, nil
}
