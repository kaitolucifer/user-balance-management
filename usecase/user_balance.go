package usecase

import (
	"errors"

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
	return u.repo.AddUserBalanceByUserID(userID, amount, transactionID)
}

// ReduceBalance ユーザーIDでユーザー残高を減算
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

// ユーザー残高を一斉に加算
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
