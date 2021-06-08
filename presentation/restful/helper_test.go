package presentation

import (
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestHandleError(t *testing.T) {
	cases := []struct {
		Name             string
		Err              error
		ExpectedMsg      string
		ExpectedStatus   string
		ExpectedHTTPCode int
	}{
		{"duplicated transaction_id", errors.New("transaction_id must be unique"), "transaction_id must be unique", "fail", http.StatusUnprocessableEntity},
		{"other postgresql error", errors.New("database error"), "database error", "error", http.StatusInternalServerError},
		{"user not found", errors.New("user not found"), "user not found", "fail", http.StatusNotFound},
		{"balance insufficient error", errors.New("balance insufficient"), "user balance is insufficient", "fail", http.StatusUnprocessableEntity},
		{"update failed error", errors.New("update failed"), "update failed, please retry", "fail", http.StatusConflict},
		{"other server error", errors.New("server error"), "internal server error", "error", http.StatusInternalServerError},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			status, msg, httpCode := handleError(c.Err)
			if status != c.ExpectedStatus {
				t.Errorf("expect status [%s], got [%s]", c.ExpectedStatus, status)
			}
			if msg != c.ExpectedMsg {
				t.Errorf("expect message [%s], got [%s]", c.ExpectedMsg, msg)
			}
			if httpCode != c.ExpectedHTTPCode {
				t.Errorf("expect http code [%d], got [%d]", c.ExpectedHTTPCode, httpCode)
			}
		})
	}
}

func TestGetValidator(t *testing.T) {
	validate := getValidator()
	var expectedType *validator.Validate
	if reflect.TypeOf(validate) != reflect.TypeOf(expectedType) {
		t.Errorf("return type is not *validator.Validate")
	}
}
