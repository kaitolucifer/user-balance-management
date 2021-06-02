package usecase

import (
	"errors"

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

func (u *userBalanceUsecase) AddBalance(userID string, amount int, transactionID string) error {
	return u.repo.AddUserBalanceByUserID(userID, amount, transactionID)
}

func (u *userBalanceUsecase) ReduceBalance(userID string, amount int, transactionID string) error {
	userBalance, err := u.repo.GetUserBalanceByUserID(userID)
	if err != nil {
		return err
	}

	if userBalance.Balance-amount < 0 {
		return errors.New("balance insufficient")
	}

	err = u.repo.ReduceUserBalanceByUserID(userID, amount, transactionID)
	if err != nil {
		return err
	}

	return nil
}

func (u *userBalanceUsecase) AddAllUserBalance(amount int, transactionID string) error {
	return u.repo.AddAllUserBalance(amount, transactionID)
}

func (u *userBalanceUsecase) GetBalance(userID string) (int, error) {
	userBalance, err := u.repo.GetUserBalanceByUserID(userID)
	if err != nil {
		return 0, err
	}

	return userBalance.Balance, nil
}
