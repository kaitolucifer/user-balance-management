package presentation

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// handleError エラーからハンドラに必要な情報を吐き出すヘルパー
func handleError(err error) (string, string, int) {
	var status string
	var msg string
	var httpCode int
	if err != nil {
		if err.Error() == "database error" {
			msg = "database error"
			status = "error"
			httpCode = http.StatusInternalServerError
		} else if err.Error() == "transaction_id must be unique" {
			msg = "transaction_id must be unique"
			status = "fail"
			httpCode = http.StatusUnprocessableEntity
		} else if err.Error() == "user not found" {
			status = "fail"
			msg = "user not found"
			httpCode = http.StatusNotFound
		} else if err.Error() == "balance insufficient" {
			status = "fail"
			msg = "user balance is insufficient"
			httpCode = http.StatusUnprocessableEntity
		} else if err.Error() == "update failed" {
			// データ競合が発生
			status = "fail"
			msg = "update failed, please retry"
			httpCode = http.StatusConflict
		} else if err.Error() == "current thread is not associated with a transaction" {
			status = "fail"
			msg = "current thread is not associated with a transaction"
			httpCode = http.StatusInternalServerError
		} else {
			status = "error"
			msg = "internal server error"
			httpCode = http.StatusInternalServerError
		}
	}

	return status, msg, httpCode
}

// getValidator 入力データのバリデーターを取得用のヘルパー
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
