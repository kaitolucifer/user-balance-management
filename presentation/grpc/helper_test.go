package presentation

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
)

func TestHandleError(t *testing.T) {
	cases := []struct {
		Name         string
		Err          error
		ExpectedMsg  string
		ExpectedCode codes.Code
	}{
		{"empty user_id", errors.New("user_id is empty"), "user_id is empty", codes.InvalidArgument},
		{"empty transaction_id", errors.New("transaction_id is empty"), "transaction_id is empty", codes.InvalidArgument},
		{"non-positive amount", errors.New("amount must be positive"), "amount must be positive", codes.InvalidArgument},
		{"0 amount", errors.New("amount can't be 0"), "amount can't be 0", codes.InvalidArgument},
		{"duplicated transaction_id", errors.New("transaction_id must be unique"), "transaction_id must be unique", codes.AlreadyExists},
		{"other postgresql error", errors.New("database error"), "database error", codes.Internal},
		{"user not found", errors.New("user not found"), "user not found", codes.NotFound},
		{"balance insufficient error", errors.New("balance insufficient"), "user balance is insufficient", codes.FailedPrecondition},
		{"update failed error", errors.New("update failed"), "update failed, please retry", codes.Unavailable},
		{"other server error", errors.New("server error"), "internal server error", codes.Internal},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			st := handleError(c.Err)
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
