package infrastructure

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/kaitolucifer/user-balance-management/domain"
	_ "github.com/mattn/go-sqlite3"
)

var repo domain.UserBalanceRepository

// NewMockDatabase 新しいSQLiteの接続を作成
// SQLiteではマルチスレッドの書き込みが制限されるため、テストケース毎に新しい接続の作成が必要
func NewMockDatabase(file string) *DB {
	conn, _ := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=memory&cache=shared", file))

	conn.Exec(`CREATE TABLE user_balance (
		user_id TEXT PRIMARY KEY,
		balance INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	)`)

	conn.Exec(`CREATE TABLE transaction_history(
		transaction_id TEXT PRIMARY KEY,
		user_id TEXT,
		transaction_type INTEGER NOT NULL,
		amount INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	)`)

	conn.Exec(`INSERT INTO user_balance (user_id, balance, created_at, updated_at) VALUES
		('test_user1', 10000, '2021-05-29', '2021-05-29'),
		('test_user2', 20000, '2021-05-29', '2021-05-29'),
		('test_user3', 30000, '2021-05-29', '2021-05-29'),
		('test_user4', 40000, '2021-05-29', '2021-05-29'),
		('test_user5', 50000, '2021-05-29', '2021-05-29')`)

	conn.Exec(`INSERT INTO transaction_history (transaction_id, transaction_type, amount, created_at, updated_at) VALUES
		('b8eb7ccc-6bc3-4be3-b7f8-e2701bf19a6b', 2, 10000, '2021-05-29', '2021-05-29')`)

	return &DB{conn}
}

func TestGetUserBalanceByUserID(t *testing.T) {
	cases := []struct {
		Name            string
		UserID          string
		ExpectedBalance int
		ExpectedErrMsg  string
	}{
		{"existent user1", "test_user1", 10000, ""},
		{"existent user2", "test_user3", 30000, ""},
		{"sql injection", "'; DROP TABLE user_balance;'", 0, "sql: no rows in result set"},
		{"nonexistent user", "unknown", 0, "sql: no rows in result set"},
	}

	for i, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			file := "get-" + strconv.Itoa(i)
			db := NewMockDatabase(file)
			defer db.Close()
			repo = NewUserBalanceRepository(*db)
			userBalance, err := repo.GetUserBalanceByUserID(c.UserID)
			if err != nil {
				var pgErr *pgconn.PgError
				if c.ExpectedErrMsg == "" {
					t.Errorf("expect no error but got [%s]", err)
				} else if errors.As(err, &pgErr) {
					if pgErr.Message != c.ExpectedErrMsg {
						t.Errorf("expect error [%s], got [%s]", c.ExpectedErrMsg, err)
					}
				} else {
					if err.Error() != c.ExpectedErrMsg {
						t.Errorf("expect error [%s], got [%s]", c.ExpectedErrMsg, err)
					}
				}
			} else {
				if c.ExpectedErrMsg != "" {
					t.Errorf("expect error [%s] but got no one", c.ExpectedErrMsg)
				} else if userBalance.UserID != c.UserID {
					t.Errorf("expect user_id [%s], got [%s]", c.UserID, userBalance.UserID)
				} else if userBalance.Balance != c.ExpectedBalance {
					t.Errorf("expect balance [%d], got [%d]", c.ExpectedBalance, userBalance.Balance)
				}
			}
		})
	}
}

func TestAddUserBalanceByUserID(t *testing.T) {
	cases := []struct {
		Name            string
		UserID          string
		Amount          int
		TransactionID   string
		ExpectedBalance int
		ExpectedErrMsg  string
	}{
		{"existent user", "test_user1", 1000, "ab20818d-9889-4e6b-b32f-c2be401ec02d", 11000, ""},
		{"duplicated transaction_id", "test_user1", 1000, "b8eb7ccc-6bc3-4be3-b7f8-e2701bf19a6b", 10000, "UNIQUE constraint failed: transaction_history.transaction_id"},
		{"sql injection", "'; DROP TABLE user_balance;'", 0, "ab20818d-9889-4e6b-b32f-c2be401ec02d", 0, "sql: no rows in result set"},
		{"nonexistent user", "unknown", 0, "ab20818d-9889-4e6b-b32f-c2be401ec02d", 0, "sql: no rows in result set"},
	}

	for i, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			file := "add-" + strconv.Itoa(i)
			db := NewMockDatabase(file)
			defer db.Close()
			repo = NewUserBalanceRepository(*db)
			err := repo.AddUserBalanceByUserID(c.UserID, c.Amount, c.TransactionID)
			if err != nil {
				var pgErr *pgconn.PgError
				if c.ExpectedErrMsg == "" {
					t.Errorf("expect no error but got [%s]", err)
				} else if errors.As(err, &pgErr) {
					if pgErr.Message != c.ExpectedErrMsg {
						t.Errorf("expect error [%s], got [%s]", c.ExpectedErrMsg, err)
					}
				} else {
					if err.Error() != c.ExpectedErrMsg {
						t.Errorf("expect error [%s], got [%s]", c.ExpectedErrMsg, err)
					}
				}
			} else {
				if c.ExpectedErrMsg != "" {
					t.Errorf("expect error [%s] but got no one", c.ExpectedErrMsg)
				}
				// 更新後残高を検証
				if !strings.HasPrefix(c.Name, "nonexistent") {
					var balance int
					row := db.QueryRow("SELECT balance FROM user_balance WHERE user_id = $1", c.UserID)
					row.Scan(&balance)
					if balance != c.ExpectedBalance {
						t.Errorf("expect balance [%d] but got [%d]", c.ExpectedBalance, balance)
					}
				}
			}
		})
	}
}

func TestReduceUserBalanceByUserID(t *testing.T) {
	cases := []struct {
		Name            string
		UserID          string
		Amount          int
		TransactionID   string
		ExpectedBalance int
		ExpectedErrMsg  string
	}{
		{"existent user", "test_user1", 1000, "ab20818d-9889-4e6b-b32f-c2be401ec02d", 9000, ""},
		{"duplicated transaction_id", "test_user1", 1000, "b8eb7ccc-6bc3-4be3-b7f8-e2701bf19a6b", 10000, "UNIQUE constraint failed: transaction_history.transaction_id"},
		{"nonexistent user", "unknown", 0, "ab20818d-9889-4e6b-b32f-c2be401ec02d", 0, "update failed"},
		{"sql injection", "'; DROP TABLE user_balance;'", 0, "ab20818d-9889-4e6b-b32f-c2be401ec02d", 0, "update failed"},
		{"insufficient balance", "test_user5", 60000, "ab20818d-9889-4e6b-b32f-c2be401ec02d", 50000, "update failed"},
	}

	for i, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			file := "reduce-" + strconv.Itoa(i)
			db := NewMockDatabase(file)
			defer db.Close()
			repo = NewUserBalanceRepository(*db)
			err := repo.ReduceUserBalanceByUserID(c.UserID, c.Amount, c.TransactionID)
			if err != nil {
				var pgErr *pgconn.PgError
				if c.ExpectedErrMsg == "" {
					t.Errorf("expect no error but got [%s]", err)
				} else if errors.As(err, &pgErr) {
					if pgErr.Message != c.ExpectedErrMsg {
						t.Errorf("expect error [%s], got [%s]", c.ExpectedErrMsg, err)
					}
				} else {
					if err.Error() != c.ExpectedErrMsg {
						t.Errorf("expect error [%s], got [%s]", c.ExpectedErrMsg, err)
					}
				}
			} else {
				if c.ExpectedErrMsg != "" {
					t.Errorf("expect error [%s] but got no one", c.ExpectedErrMsg)
				}
				// 更新後残高を検証
				if !strings.HasPrefix(c.Name, "nonexistent") {
					var balance int
					row := db.QueryRow("SELECT balance FROM user_balance WHERE user_id = $1", c.UserID)
					row.Scan(&balance)
					if balance != c.ExpectedBalance {
						t.Errorf("expect balance [%d] but got [%d]", c.ExpectedBalance, balance)
					}
				}
			}
		})
	}
}

func TestAddAllUserBalance(t *testing.T) {
	cases := []struct {
		Name             string
		Amount           int
		TransactionID    string
		ExpectedBalances map[string]int
		ExpectedErrMsg   string
	}{
		{"normal case", 1000, "ab20818d-9889-4e6b-b32f-c2be401ec02d",
			map[string]int{"test_user1": 11000, "test_user2": 21000, "test_user3": 31000, "test_user4": 41000, "test_user5": 51000},
			""},
		{"duplicated transaction_id", 1000, "b8eb7ccc-6bc3-4be3-b7f8-e2701bf19a6b",
			map[string]int{"test_user1": 10000, "test_user2": 20000, "test_user3": 30000, "test_user4": 40000, "test_user5": 50000},
			"UNIQUE constraint failed: transaction_history.transaction_id"},
	}

	for i, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			file := "add-all-" + strconv.Itoa(i)
			db := NewMockDatabase(file)
			defer db.Close()
			repo = NewUserBalanceRepository(*db)
			err := repo.AddAllUserBalance(c.Amount, c.TransactionID)
			if err != nil {
				var pgErr *pgconn.PgError
				if c.ExpectedErrMsg == "" {
					t.Errorf("expect no error but got [%s]", err)
				} else if errors.As(err, &pgErr) {
					if pgErr.Message != c.ExpectedErrMsg {
						t.Errorf("expect error [%s], got [%s]", c.ExpectedErrMsg, err)
					}
				} else {
					if err.Error() != c.ExpectedErrMsg {
						t.Errorf("expect error [%s], got [%s]", c.ExpectedErrMsg, err)
					}
				}
			} else {
				if c.ExpectedErrMsg != "" {
					t.Errorf("expect error [%s] but got no one", c.ExpectedErrMsg)
				}
				// 更新後残高を検証
				for userID, expectedBalance := range c.ExpectedBalances {
					var balance int
					row := db.QueryRow("SELECT balance FROM user_balance WHERE user_id = $1", userID)
					row.Scan(&balance)
					if balance != expectedBalance {
						t.Errorf("expect balance [%d] for [%s] but got [%d]", expectedBalance, userID, balance)
					}
				}
			}
		})
	}
}
