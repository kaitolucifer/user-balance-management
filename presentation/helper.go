package presentation

import (
	"database/sql"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
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
			httpCode = http.StatusUnprocessableEntity
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
			httpCode = http.StatusUnprocessableEntity
		} else if err.Error() == "update failed"{
			// データ競合が発生
			status = "fail"
			msg = "update failed, please retry"
			httpCode = http.StatusConflict
		} else {
			status = "error"
			msg = "internal server error"
			httpCode = http.StatusInternalServerError
		}
	}

	return status, msg, httpCode
}

func getValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return v
}

