package usecase

import "github.com/kaitolucifer/user-balance-management/domain"

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
