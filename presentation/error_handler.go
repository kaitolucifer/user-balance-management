package presentation

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/jackc/pgconn"
)

func handleError(err error) (string, string, int) {
	var status string
	var msg string
	var httpCode int

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			msg = "transaction_id must be unique"
			status = "fail"
			httpCode = http.StatusBadRequest
		default:
			msg = "database error"
			status = "error"
			httpCode = http.StatusInternalServerError
		}
	} else {
		if err == sql.ErrNoRows {
			status = "fail"
			msg = "user_id not found"
			httpCode = http.StatusNotFound
		} else if err.Error() == "balance insufficient" {
			status = "fail"
			msg = "user balance is insufficient"
			httpCode = http.StatusBadRequest
		} else {
			status = "error"
			msg = "internal server error"
			httpCode = http.StatusInternalServerError
		}
	}

	return status, msg, httpCode
}
