package usecase

import (
	"github.com/kaitolucifer/user-balance-management/domain"
)

type userBalanceUsecase struct {
	repo domain.UserBalanceRepository
}

func NewUserBalanceUsecase(repo domain.UserBalanceRepository) domain.UserBalanceUsecase {
	return &userBalanceUsecase{
		repo: repo,
	}
}

func (u *userBalanceUsecase) AddBalance(userID string, amount int) error {
	_, err := u.repo.GetUserBalanceByUserID(userID)
	if err != nil {
		return err
	}

	err = u.repo.AddUserBalanceByUserID(userID, amount)
	if err != nil {
		return err
	}

	return nil
}

func (u *userBalanceUsecase) GetBalance(userID string) (int, error) {
	userBalance, err := u.repo.GetUserBalanceByUserID(userID)

	var balance int

	if err != nil {
		return balance, err
	}

	balance = userBalance.Balance
	return balance, nil
}
