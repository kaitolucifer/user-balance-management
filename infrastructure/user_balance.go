package infrastructure

import (
	"context"
	"time"

	"github.com/kaitolucifer/user-balance-management/domain"
)

type userBalanceRepository struct {
	Conn DB
}

func NewUserBalanceRepository(db DB) domain.UserBalanceRepository {
	return &userBalanceRepository{db}
}

func (repo *userBalanceRepository) GetUserBalanceByUserId(userId string) (domain.UserBalanceModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var userBalances domain.UserBalanceModel

	row := repo.Conn.DB.QueryRowContext(ctx, "SELECT * FROM user_balance WHERE user_id = $1", userId)
	err := row.Scan(
		&userBalances.UserID,
		&userBalances.Balance,
		&userBalances.CreatedAt,
		&userBalances.UpdatedAt,
	)
	if err != nil {
		return userBalances, err
	}

	return userBalances, nil
}
