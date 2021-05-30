package presentation

import (
	"net/http"

	"github.com/jackc/pgconn"
)

func handlePgError(pgErr *pgconn.PgError) (string, string, int) {
	var status string
	var msg string
	var httpCode int

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

	return status, msg, httpCode
}
