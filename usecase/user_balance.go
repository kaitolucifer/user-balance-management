package usecase

import (
	"database/sql"
	"errors"

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
	err := u.repo.AddUserBalanceByUserID(userID, amount, transactionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found")
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

	return nil
}

// ReduceBalance ユーザーIDでユーザー残高を減算
func (u *userBalanceUsecase) ReduceBalance(userID string, amount int, transactionID string) error {
	userBalance, err := u.repo.GetUserBalanceByUserID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		return err
	}

	if userBalance.Balance-amount < 0 {
		return errors.New("balance insufficient")
	}

	err = u.repo.ReduceUserBalanceByUserID(userID, amount, transactionID)
	if err != nil {
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

	return nil
}

// ユーザー残高を一斉に加算
func (u *userBalanceUsecase) AddAllUserBalance(amount int, transactionID string) error {
	err := u.repo.AddAllUserBalance(amount, transactionID)
	if err != nil {
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
	return nil
}

func (u *userBalanceUsecase) GetBalance(userID string) (int, error) {
	userBalance, err := u.repo.GetUserBalanceByUserID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("user not found")
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return 0, errors.New("transaction_id must be unique")
			default:
				return 0, errors.New("database error")
			}
		}
		return 0, err
	}

	return userBalance.Balance, nil
}
