package presentation

import (
	"context"
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/kaitolucifer/user-balance-management/domain"
	"github.com/kaitolucifer/user-balance-management/presentation/grpc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var handler *GrpcUserBalanceHander

type mockUsecase struct {
	userBalance        []domain.UserBalanceModel
	transactionHistory []domain.TransactionHistoryModel
}

func NewMockUsecase() domain.UserBalanceUsecase {
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
			TransactionType: domain.TypeAddUserBalance,
			Amount:          5000,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	return &mockUsecase{
		userBalance:        userBalances,
		transactionHistory: transactionHistory,
	}
}

func (u *mockUsecase) AddBalance(userID string, amount int, transactionID string) error {
	userExist := false
	for _, ub := range u.userBalance {
		if ub.UserID == userID {
			userExist = true
		}
	}
	if !userExist {
		return sql.ErrNoRows
	}

	for _, th := range u.transactionHistory {
		if th.TransactionID == transactionID {
			return &pgconn.PgError{
				Code: "23505",
			}
		}
	}

	return nil
}

func (u *mockUsecase) ReduceBalance(userID string, amount int, transactionID string) error {
	userExist := false
	for _, ub := range u.userBalance {
		if ub.UserID == userID && ub.Balance-amount >= 0 {
			userExist = true
		} else if ub.UserID == userID && ub.Balance-amount < 0 {
			return errors.New("balance insufficient")
		}
	}
	if !userExist {
		return sql.ErrNoRows
	}

	for _, th := range u.transactionHistory {
		if th.TransactionID == transactionID {
			return &pgconn.PgError{
				Code: "23505",
			}
		}
	}

	return nil
}

func (u *mockUsecase) AddAllUserBalance(amount int, transactionID string) error {
	for _, th := range u.transactionHistory {
		if th.TransactionID == transactionID {
			return &pgconn.PgError{
				Code: "23505",
			}
		}
	}

	return nil
}

func (u *mockUsecase) GetBalance(userID string) (int, error) {
	for _, ub := range u.userBalance {
		if ub.UserID == userID {
			return ub.Balance, nil
		}
	}
	return 0, sql.ErrNoRows
}

func TestMain(m *testing.M) {
	usecase := NewMockUsecase()
	app := App{
		InfoLog:  log.New(ioutil.Discard, "INFO\t", log.Ldate|log.Ltime),
		ErrorLog: log.New(ioutil.Discard, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
	}
	handler = NewGrpcUserBalanceHander(usecase, &app)

	code := m.Run()
	os.Exit(code)
}
func TestGetBalanceByUserID(t *testing.T) {
	cases := []struct {
		Name            string
		UserID          string
		ExpectedBalance int32
		ExpectedMsg     string
		ExpectedCode    codes.Code
	}{
		{"existent user1", "test_user1", 10000, "", codes.OK},
		{"existent user2", "test_user2", 20000, "", codes.OK},
		{"existent user3", "test_user3", 30000, "", codes.OK},
		{"nonexistent user1", "unknown", 0, "user_id not found", codes.NotFound},
		{"nonexistent user2", "someone", 0, "user_id not found", codes.NotFound},
		{"empty user id", "", 0, "user_id is empty", codes.InvalidArgument},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ctx := context.Background()
			req := &proto.GetUserBalanceRequest{
				UserId: c.UserID,
			}
			resp, err := handler.GetBalanceByUserID(ctx, req)
			st, ok := status.FromError(err)
			if !ok {
				t.Fatal("failed to get status from error")
			}
			if st.Code() != c.ExpectedCode {
				t.Errorf("expect status code [%s] but got [%s]", c.ExpectedCode, st.Code())
			}
			if st.Message() != c.ExpectedMsg {
				if st.Message() == "" {
					t.Errorf("expect message [%s] but got no one", c.ExpectedMsg)
				} else if c.ExpectedMsg == "" {
					t.Errorf("expect no message but got [%s]", st.Message())
				} else {
					t.Errorf("expect message [%s] but got [%s]", c.ExpectedMsg, st.Message())
				}
			}
			if strings.HasPrefix(c.Name, "existent") {
				if resp.GetBalance() != c.ExpectedBalance {
					t.Errorf("expect balance [%d] but got [%d]", c.ExpectedBalance, resp.Balance)
				}
			}
		})
	}
}

func TestChangeBalanceByUserID(t *testing.T) {
	cases := []struct {
		Name          string
		UserID        string
		Amount        int32
		TransactionID string
		ExpectedMsg   string
		ExpectedCode  codes.Code
	}{
		{"existent user1", "test_user1", 1000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "", codes.OK},
		{"existent user2", "test_user2", 10000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "", codes.OK},
		{"existent user3", "test_user3", 100000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "", codes.OK},
		{"existent user3", "test_user4", -10000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "", codes.OK},
		{"existent user3", "test_user5", -20000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "", codes.OK},
		{"nonexistent user1", "unknown", 1000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "user_id not found", codes.NotFound},
		{"nonexistent user2", "someone", 1000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "user_id not found", codes.NotFound},
		{"duplicated transaction_id", "test_user5", 1000, "b8eb7ccc-6bc3-4be3-b7f8-e2701bf19a6b", "transaction_id must be unique", codes.AlreadyExists},
		{"invalid amount2", "test_user5", 0, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "amount can't be 0", codes.InvalidArgument},
		{"empty user id", "", 0, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "user_id is empty", codes.InvalidArgument},
		{"empty transaction_id", "test_user1", 0, "", "transaction_id is empty", codes.InvalidArgument},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ctx := context.Background()
			req := &proto.ChangeUserBalanceRequest{
				UserId:        c.UserID,
				Amount:        c.Amount,
				TransactionId: c.TransactionID,
			}
			_, err := handler.ChangeBalanceByUserID(ctx, req)
			st, ok := status.FromError(err)
			if !ok {
				t.Fatal("failed to get status from error")
			}
			if st.Code() != c.ExpectedCode {
				t.Errorf("expect status code [%s] but got [%s]", c.ExpectedCode, st.Code())
			}
			if st.Message() != c.ExpectedMsg {
				if st.Message() == "" {
					t.Errorf("expect message [%s] but got no one", c.ExpectedMsg)
				} else if c.ExpectedMsg == "" {
					t.Errorf("expect no message but got [%s]", st.Message())
				} else {
					t.Errorf("expect message [%s] but got [%s]", c.ExpectedMsg, st.Message())
				}
			}
		})
	}
}

func TestAddAllUserBalance(t *testing.T) {
	cases := []struct {
		Name           string
		Amount         int32
		TransactionID  string
		ExpectedMsg    string
		ExpectedCode   codes.Code
	}{
		{"normal case1", 1000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "", codes.OK},
		{"normal case2", 10000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "", codes.OK},
		{"normal case3", 100000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "", codes.OK},
		{"duplicated transaction_id", 1000, "b8eb7ccc-6bc3-4be3-b7f8-e2701bf19a6b", "transaction_id must be unique", codes.AlreadyExists},
		{"invalid amount1", -100, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "amount must be positive", codes.InvalidArgument},
		{"invalid amount2", 0, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "amount must be positive", codes.InvalidArgument},
		{"empty transaction_id", 0, "", "transaction_id is empty", codes.InvalidArgument},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ctx := context.Background()
			req := &proto.AddAllUserBalanceRequest{
				Amount:        c.Amount,
				TransactionId: c.TransactionID,
			}
			_, err := handler.AddAllUserBalance(ctx, req)
			st, ok := status.FromError(err)
			if !ok {
				t.Fatal("failed to get status from error")
			}
			if st.Code() != c.ExpectedCode {
				t.Errorf("expect status code [%s] but got [%s]", c.ExpectedCode, st.Code())
			}
			if st.Message() != c.ExpectedMsg {
				if st.Message() == "" {
					t.Errorf("expect message [%s] but got no one", c.ExpectedMsg)
				} else if c.ExpectedMsg == "" {
					t.Errorf("expect no message but got [%s]", st.Message())
				} else {
					t.Errorf("expect message [%s] but got [%s]", c.ExpectedMsg, st.Message())
				}
			}
		})
	}
}
