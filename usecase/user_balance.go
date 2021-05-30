package usecase

import (
	"database/sql"
	"errors"

	"github.com/kaitolucifer/user-balance-management/domain"
)

type userBalanceUsecase struct {
	repo domain.UserBalanceRepository
}

func NewUserBalanceUsecase(repo domain.UserBalanceRepository) *userBalanceUsecase {
	return &userBalanceUsecase{
		repo: repo,
	}
}

func (u *userBalanceUsecase) AddBalance(amount int) error {
	return nil
}

func (u *userBalanceUsecase) GetBalance(userID string) (int, error) {
	userBalance, err := u.repo.GetUserBalanceByUserId(userID)

	var balance int

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return balance, errors.New("userID not found")
		default:
			return balance, errors.New("something went wrong")
		}
	}

	balance = userBalance.Balance
	return balance, nil
}
