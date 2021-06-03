package presentation

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/jackc/pgconn"
	"github.com/kaitolucifer/user-balance-management/domain"
)

var handler *UserBalanceHandler

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
	handler = NewUserBalanceHander(usecase, &app)
	code := m.Run()
	os.Exit(code)
}

func TestHealthCheck(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handler.HealthCheck))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Errorf("expect no error but got [%s]", err)
	}

	resBody, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Errorf("expect no error but got [%s]", err)
	}

	expectedResBody := `{"status": "success", "message": "healthy"}`
	strResBody := string(resBody)
	if strResBody != expectedResBody {
		t.Errorf("expect response body [%s]\nbut got [%s]", expectedResBody, strResBody)
	}
}

func TestNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handler.NotFound))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/test")
	if err != nil {
		t.Errorf("expect no error but got [%s]", err)
	}

	resBody, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Errorf("expect no error but got [%s]", err)
	}

	expectedResBody := `{"status": "error", "message": "The requested URL /test was not found on this server."}`
	strResBody := string(resBody)
	if strResBody != expectedResBody {
		t.Errorf("expect response body [%s]\nbut got [%s]", expectedResBody, strResBody)
	}
}

func TestGetUserBalance(t *testing.T) {
	cases := []struct {
		Name            string
		UserID          string
		ExpectedBalance int
		ExpectedStatus  string
		ExpectedMsg     string
		ExpectedCode    int
	}{
		{"existent user1", "test_user1", 10000, "success", "", http.StatusOK},
		{"existent user2", "test_user2", 20000, "success", "", http.StatusOK},
		{"existent user3", "test_user3", 30000, "success", "", http.StatusOK},
		{"nonexistent user1", "unknown", 0, "fail", "user_id not found", http.StatusNotFound},
		{"nonexistent user2", "someone", 0, "fail", "user_id not found", http.StatusNotFound},
		{"empty user id", "", 0, "fail", "user_id is empty", http.StatusBadRequest},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/balance", nil)
			w := httptest.NewRecorder()
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("userID", c.UserID)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
			h := http.HandlerFunc(handler.GetUserBalance)
			h.ServeHTTP(w, r)

			if w.Code != c.ExpectedCode {
				t.Errorf("expect http status code [%d] but got [%d]", c.ExpectedCode, w.Code)
			}

			body, err := io.ReadAll(w.Body)
			if err != nil {
				t.Errorf("expect no error but got [%s]", err)
			}
			var resp getUserBalanceResponse
			err = json.Unmarshal(body, &resp)
			if err != nil {
				t.Errorf("expect no error but got [%s]", err)
			}

			if resp.Status != c.ExpectedStatus {
				t.Errorf("expect status [%s] but got [%s]", c.ExpectedStatus, resp.Status)
			}
			if resp.Message != c.ExpectedMsg {
				if resp.Message == "" {
					t.Errorf("expect message [%s] but got no one", c.ExpectedMsg)
				} else if c.ExpectedMsg == "" {
					t.Errorf("expect no message but got [%s]", resp.Message)
				} else {
					t.Errorf("expect message [%s] but got [%s]", c.ExpectedMsg, resp.Message)
				}
			}
			if strings.HasPrefix(c.Name, "existent") {
				if *resp.Balance != c.ExpectedBalance {
					t.Errorf("expect balance [%d] but got [%d]", c.ExpectedBalance, resp.Balance)
				}
			}
		})
	}

}

func TestAddUserBalance(t *testing.T) {
	cases := []struct {
		Name           string
		UserID         string
		Amount         int
		TransactionID  string
		ExpectedStatus string
		ExpectedMsg    string
		ExpectedCode   int
	}{
		{"existent user1", "test_user1", 1000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "success", "user balance has been added successfully", http.StatusOK},
		{"existent user2", "test_user2", 10000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "success", "user balance has been added successfully", http.StatusOK},
		{"existent user3", "test_user3", 100000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "success", "user balance has been added successfully", http.StatusOK},
		{"nonexistent user1", "unknown", 1000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "fail", "user_id not found", http.StatusNotFound},
		{"nonexistent user2", "someone", 1000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "fail", "user_id not found", http.StatusNotFound},
		{"duplicated transaction_id", "test_user5", 1000, "b8eb7ccc-6bc3-4be3-b7f8-e2701bf19a6b", "fail", "transaction_id must be unique", http.StatusUnprocessableEntity},
		{"invalid amount1", "test_user3", -100, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "fail", "amount must be positive", http.StatusUnprocessableEntity},
		{"invalid amount2", "test_user5", 0, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "fail", "amount must be positive", http.StatusUnprocessableEntity},
		{"empty user id", "", 0, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "fail", "user_id is empty", http.StatusBadRequest},
		{"empty transaction_id", "test_user1", 0, "", "fail", "transaction_id can't be null", http.StatusBadRequest},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			r := httptest.NewRequest("PATCH", "/balance/add", nil)
			w := httptest.NewRecorder()
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("userID", c.UserID)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
			reqModel := ChangeUserBalanceRequest{
				Amount:        &c.Amount,
				TransactionID: c.TransactionID,
			}
			reqBody, _ := json.Marshal(&reqModel)
			r.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
			h := http.HandlerFunc(handler.ChangeUserBalance)
			h.ServeHTTP(w, r)

			if w.Code != c.ExpectedCode {
				t.Errorf("expect http status code [%d] but got [%d]", c.ExpectedCode, w.Code)
			}

			body, err := io.ReadAll(w.Body)
			if err != nil {
				t.Errorf("expect no error but got [%s]", err)
			}
			var resp changeUserBalanceResponse
			err = json.Unmarshal(body, &resp)
			if err != nil {
				t.Errorf("expect no error but got [%s]", err)
			}

			if resp.Status != c.ExpectedStatus {
				t.Errorf("expect status [%s] but got [%s]", c.ExpectedStatus, resp.Status)
			}
			if resp.Message != c.ExpectedMsg {
				if resp.Message == "" {
					t.Errorf("expect message [%s] but got no one", c.ExpectedMsg)
				} else if c.ExpectedMsg == "" {
					t.Errorf("expect no message but got [%s]", resp.Message)
				} else {
					t.Errorf("expect message [%s] but got [%s]", c.ExpectedMsg, resp.Message)
				}
			}
		})
	}
}

func TestReduceUserBalance(t *testing.T) {
	cases := []struct {
		Name           string
		UserID         string
		Amount         int
		TransactionID  string
		ExpectedStatus string
		ExpectedMsg    string
		ExpectedCode   int
	}{
		{"existent user1", "test_user1", 1000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "success", "user balance has been added successfully", http.StatusOK},
		{"existent user2", "test_user2", 10000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "success", "user balance has been added successfully", http.StatusOK},
		{"insuffcient balance1", "test_user4", 50000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "fail", "user balance is insufficient", http.StatusUnprocessableEntity},
		{"insuffcient balance2", "test_user5", 60000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "fail", "user balance is insufficient", http.StatusUnprocessableEntity},
		{"nonexistent user1", "unknown", 1000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "fail", "user_id not found", http.StatusNotFound},
		{"nonexistent user2", "someone", 1000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "fail", "user_id not found", http.StatusNotFound},
		{"duplicated transaction_id", "test_user5", 1000, "b8eb7ccc-6bc3-4be3-b7f8-e2701bf19a6b", "fail", "transaction_id must be unique", http.StatusUnprocessableEntity},
		{"invalid amount1", "test_user3", -100, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "fail", "amount must be positive", http.StatusUnprocessableEntity},
		{"invalid amount2", "test_user5", 0, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "fail", "amount must be positive", http.StatusUnprocessableEntity},
		{"empty user id", "", 0, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "fail", "user_id is empty", http.StatusBadRequest},
		{"empty transaction_id", "test_user1", 0, "", "fail", "transaction_id can't be null", http.StatusBadRequest},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			r := httptest.NewRequest("PATCH", "/balance/reduce", nil)
			w := httptest.NewRecorder()
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("userID", c.UserID)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
			reqModel := ChangeUserBalanceRequest{
				Amount:        &c.Amount,
				TransactionID: c.TransactionID,
			}
			reqBody, _ := json.Marshal(&reqModel)
			r.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
			h := http.HandlerFunc(handler.ChangeUserBalance)
			h.ServeHTTP(w, r)

			if w.Code != c.ExpectedCode {
				t.Errorf("expect http status code [%d] but got [%d]", c.ExpectedCode, w.Code)
			}

			body, err := io.ReadAll(w.Body)
			if err != nil {
				t.Errorf("expect no error but got [%s]", err)
			}
			var resp changeUserBalanceResponse
			err = json.Unmarshal(body, &resp)
			if err != nil {
				t.Errorf("expect no error but got [%s]", err)
			}

			if resp.Status != c.ExpectedStatus {
				t.Errorf("expect status [%s] but got [%s]", c.ExpectedStatus, resp.Status)
			}
			if resp.Message != c.ExpectedMsg {
				if resp.Message == "" {
					t.Errorf("expect message [%s] but got no one", c.ExpectedMsg)
				} else if c.ExpectedMsg == "" {
					t.Errorf("expect no message but got [%s]", resp.Message)
				} else {
					t.Errorf("expect message [%s] but got [%s]", c.ExpectedMsg, resp.Message)
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
		ExpectedStatus string
		ExpectedMsg    string
		ExpectedCode   int
	}{
		{"normal case1", 1000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "success", "user balance has been added successfully", http.StatusOK},
		{"normal case2", 10000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "success", "user balance has been added successfully", http.StatusOK},
		{"normal case3", 100000, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "success", "user balance has been added successfully", http.StatusOK},
		{"duplicated transaction_id", 1000, "b8eb7ccc-6bc3-4be3-b7f8-e2701bf19a6b", "fail", "transaction_id must be unique", http.StatusUnprocessableEntity},
		{"invalid amount1", -100, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "fail", "amount must be positive", http.StatusUnprocessableEntity},
		{"invalid amount2", 0, "917cd5c0-0bfc-4283-bc88-b5de8ad13635", "fail", "amount must be positive", http.StatusUnprocessableEntity},
		{"empty transaction_id", 0, "", "fail", "transaction_id can't be null", http.StatusBadRequest},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			r := httptest.NewRequest("PATCH", "/balance/add-all", nil)
			w := httptest.NewRecorder()
			reqModel := ChangeUserBalanceRequest{
				Amount:        &c.Amount,
				TransactionID: c.TransactionID,
			}
			reqBody, _ := json.Marshal(&reqModel)
			r.Body = ioutil.NopCloser(bytes.NewReader(reqBody))
			h := http.HandlerFunc(handler.AddAllUserBalance)
			h.ServeHTTP(w, r)

			if w.Code != c.ExpectedCode {
				t.Errorf("expect http status code [%d] but got [%d]", c.ExpectedCode, w.Code)
			}

			body, err := io.ReadAll(w.Body)
			if err != nil {
				t.Errorf("expect no error but got [%s]", err)
			}
			var resp changeUserBalanceResponse
			err = json.Unmarshal(body, &resp)
			if err != nil {
				t.Errorf("expect no error but got [%s]", err)
			}

			if resp.Status != c.ExpectedStatus {
				t.Errorf("expect status [%s] but got [%s]", c.ExpectedStatus, resp.Status)
			}
			if resp.Message != c.ExpectedMsg {
				if resp.Message == "" {
					t.Errorf("expect message [%s] but got no one", c.ExpectedMsg)
				} else if c.ExpectedMsg == "" {
					t.Errorf("expect no message but got [%s]", resp.Message)
				} else {
					t.Errorf("expect message [%s] but got [%s]", c.ExpectedMsg, resp.Message)
				}
			}
		})
	}
}
