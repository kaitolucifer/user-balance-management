package usecase

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/kaitolucifer/user-balance-management/domain"
)

type mockRepository struct {
	userBalance        []domain.UserBalanceModel
	transactionHistory []domain.TransactionHistoryModel
}

func NewMockRepository() domain.UserBalanceRepository {
	userBalances := []domain.UserBalanceModel{
		{UserID: "test_user1", Balance: 10000, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{UserID: "test_user2", Balance: 20000, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{UserID: "test_user3", Balance: 30000, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{UserID: "test_user4", Balance: 40000, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{UserID: "test_user5", Balance: 50000, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	transactionHistory := []domain.TransactionHistoryModel{
		{
			TransactionID:   "b8eb7ccc-6bc3-4be3-b7f8-e2701bf19a6b",
			UserID:          "test_user1",
			TransactionType: domain.TransactionType_AddUserBalance,
			Amount:          5000,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	return &mockRepository{
		userBalance:        userBalances,
		transactionHistory: transactionHistory,
	}
}

func (repo *mockRepository) GetCtxWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

func (repo *mockRepository) BeginTx(ctx context.Context) error {
	return nil
}

func (repo *mockRepository) Commit() error {
	return nil
}

func (repo *mockRepository) Rollback() error {
	return nil
}

func (repo *mockRepository) InsertTransactionHistory(ctx context.Context, transactionID string, userID string, transactionType domain.TransactionType, amount int) error {
	for _, th := range repo.transactionHistory {
		if th.TransactionID == transactionID {
			pgErr := &pgconn.PgError{
				Code: "23505",
			}
			return pgErr
		}
	}

	return nil
}

func (repo *mockRepository) QueryUserBalanceByUserID(ctx context.Context, userID string) (domain.UserBalanceModel, error) {
	var userBalance domain.UserBalanceModel

	for _, ub := range repo.userBalance {
		if ub.UserID == userID {
			return ub, nil
		}
	}

	return userBalance, sql.ErrNoRows
}

func (repo *mockRepository) AddUserBalanceByUserID(ctx context.Context, userID string, amount int) error {
	userExist := false
	for _, ub := range repo.userBalance {
		if ub.UserID == userID {
			userExist = true
		}
	}
	if !userExist {
		return sql.ErrNoRows
	}

	return nil
}

func (repo *mockRepository) ReduceUserBalanceByUserID(ctx context.Context, userID string, amount int) error {
	for _, ub := range repo.userBalance {
		if ub.UserID == userID && ub.Balance-amount < 0 {
			return errors.New("update failed")
		}
	}

	return nil
}

func (repo *mockRepository) AddAllUserBalance(ctx context.Context, amount int) error {
	return nil
}

var repo domain.UserBalanceRepository
var usecase domain.UserBalanceUsecase

func TestMain(m *testing.M) {
	repo = NewMockRepository()
	usecase = NewUserBalanceUsecase(repo)
	code := m.Run()
	os.Exit(code)
}

func TestAddBalance(t *testing.T) {
	cases := []struct {
		Name           string
		UserID         string
		Amount         int
		TransactionID  string
		ExpectedErrMsg string
	}{
		{"existent user", "test_user1", 10000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", ""},
		{"transaction_id must be unique", "test_user5", 50000, "b8eb7ccc-6bc3-4be3-b7f8-e2701bf19a6b", "transaction_id must be unique"},
		{"nonexistent user", "unknown", 10000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "user not found"},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			err := usecase.AddBalance(c.UserID, c.Amount, c.TransactionID)
			if err != nil {
				if c.ExpectedErrMsg == "" {
					t.Errorf("expect no error but got [%s]", err)
				} else if err.Error() != c.ExpectedErrMsg {
					t.Errorf("expect error [%s], got [%s]", c.ExpectedErrMsg, err)
				}
			} else {
				if c.ExpectedErrMsg != "" {
					t.Errorf("expect error [%s] but got no one", c.ExpectedErrMsg)
				}
			}
		})
	}
}

func TestReduceBalance(t *testing.T) {
	cases := []struct {
		Name           string
		UserID         string
		Amount         int
		TransactionID  string
		ExpectedErrMsg string
	}{
		{"existent user", "test_user1", 10000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", ""},
		{"transaction_id must be unique", "test_user5", 50000, "b8eb7ccc-6bc3-4be3-b7f8-e2701bf19a6b", "transaction_id must be unique"},
		{"insufficient balance", "test_user5", 60000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "balance insufficient"},
		{"nonexistent user", "unknown", 10000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "user not found"},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			err := usecase.ReduceBalance(c.UserID, c.Amount, c.TransactionID)
			if err != nil {
				if c.ExpectedErrMsg == "" {
					t.Errorf("expect no error but got [%s]", err)
				} else if err.Error() != c.ExpectedErrMsg {
					t.Errorf("expect error [%s], got [%s]", c.ExpectedErrMsg, err)
				}
			} else {
				if c.ExpectedErrMsg != "" {
					t.Errorf("expect error [%s] but got no one", c.ExpectedErrMsg)
				}
			}
		})
	}
}

func TestAddAllUserBalance(t *testing.T) {
	cases := []struct {
		Name           string
		Amount         int
		TransactionID  string
		ExpectedErrMsg string
	}{
		{"normal case", 10000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", ""},
		{"transaction_id must be unique", 50000, "b8eb7ccc-6bc3-4be3-b7f8-e2701bf19a6b", "transaction_id must be unique"},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			err := usecase.AddAllUserBalance(c.Amount, c.TransactionID)
			if err != nil {
				if c.ExpectedErrMsg == "" {
					t.Errorf("expect no error but got [%s]", err)
				} else if err.Error() != c.ExpectedErrMsg {
					t.Errorf("expect error [%s], got [%s]", c.ExpectedErrMsg, err)
				}
			} else {
				if c.ExpectedErrMsg != "" {
					t.Errorf("expect error [%s] but got no one", c.ExpectedErrMsg)
				}
			}
		})
	}
}

func TestGetBalance(t *testing.T) {
	cases := []struct {
		Name            string
		UserID          string
		ExpectedBalance int
		ExpectedErr     error
	}{
		{"existent user1", "test_user1", 10000, nil},
		{"existent user2", "test_user5", 50000, nil},
		{"nonexistent user", "unknown", 10000, errors.New("user not found")},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			balance, err := usecase.GetBalance(c.UserID)
			if err == nil {
				if balance != c.ExpectedBalance {
					t.Errorf("expect balance [%d], got [%d]", c.ExpectedBalance, balance)
				}
			} else if err.Error() != c.ExpectedErr.Error() {
				t.Errorf("expect error [%s], got [%s]", err, c.ExpectedErr)
			}
		})
	}

}
