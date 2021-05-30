package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/kaitolucifer/user-balance-management/domain"
)

type userBalanceRepository struct {
	Conn DB
}

func NewUserBalanceRepository(db DB) domain.UserBalanceRepository {
	return &userBalanceRepository{db}
}

func (repo *userBalanceRepository) GetUserBalanceByUserID(userId string) (domain.UserBalanceModel, error) {
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

func (repo *userBalanceRepository) AddUserBalanceByUserID(userID string, amount int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := fmt.Sprintf(`UPDATE user_balance SET balance = balance + %d, updated_at = $1 WHERE user_id = $2`, amount)
	_, err := repo.Conn.DB.ExecContext(ctx, query, userID, time.Now())
	if err != nil {
		return err
	}

	return nil
}
