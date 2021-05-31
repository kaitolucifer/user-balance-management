package presentation

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/kaitolucifer/user-balance-management/domain"
)

// App アプリケーションが持つコンポーネントや設定を格納
type App struct {
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

type UserBalanceHandler struct {
	usecase domain.UserBalanceUsecase
	App     *App
}

func NewUserBalanceHander(usecase domain.UserBalanceUsecase, app *App) UserBalanceHandler {
	return UserBalanceHandler{
		usecase: usecase,
		App:     app,
	}
}

func (h *UserBalanceHandler) Home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

type userBalanceResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Balance string `json:"balance,omitempty"`
}

func (h *UserBalanceHandler) UserBalance(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	w.Header().Set("Content-Type", "application/json")

	var resp userBalanceResponse

	balance, err := h.usecase.GetBalance(userID)
	if err != nil {
		status, msg, httpCode := handleError(err)
		if status == "error" {
			h.App.ErrorLog.Println(err)
		}

		resp.Status = status
		resp.Message = msg
		w.WriteHeader(httpCode)
		out, _ := json.Marshal(resp)
		w.Write(out)
		return
	}

	resp.Status = "success"
	resp.Balance = strconv.Itoa(balance)
	out, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

// changeUserBalanceResponse 残高を加減算するエンドポイントのレスポンスフォーマット
type changeUserBalanceResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// changeUserBalanceResponse 残高を加減算するエンドポイントのリクエストフォーマット
type ChangeUserBalanceRequest struct {
	Amount        int    `json:"amount" validate:"required"`
	TransactionID string `json:"transaction_id" validate:"required"`
}

func (h *UserBalanceHandler) ChangeUserBalance(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	w.Header().Set("Content-Type", "application/json")

	splited := strings.Split(r.RequestURI, "/")
	change_type := splited[2] // 加算か減算かを示す部分

	var resp changeUserBalanceResponse
	var req ChangeUserBalanceRequest

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		resp.Status = "fail"
		resp.Message = "request body is invalid"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		resp.Status = "fail"
		resp.Message = "request body's JSON format is invalid (amount: int, transaction_id: string)"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
		return
	}

	v := getValidator()
	if err := v.Struct(req); err != nil {
		resp.Status = "fail"
		invalidFields := []string{}
		for _, validErr := range err.(validator.ValidationErrors) {
			invalidFields = append(invalidFields, validErr.Field())
		}
		resp.Message = strings.Join(invalidFields, ", ") + " is requreid"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
		return
	}

	if req.Amount <= 0 {
		resp.Status = "fail"
		resp.Message = "amount is non-positive number"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(out)
		return
	}

	if change_type == "add" {
		err = h.usecase.AddBalance(userID, req.Amount, req.TransactionID)
	} else {
		err = h.usecase.ReduceBalance(userID, req.Amount, req.TransactionID)
	}

	if err != nil {
		status, msg, httpCode := handleError(err)
		if status == "error" {
			h.App.ErrorLog.Println(err)
		}

		resp.Status = status
		resp.Message = msg
		w.WriteHeader(httpCode)
		out, _ := json.Marshal(resp)
		w.Write(out)
		return
	}

	resp.Status = "success"
	resp.Message = "user balance added successfully"
	out, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}


func (h *UserBalanceHandler) AddAllUserBalance(w http.ResponseWriter, r *http.Request) {
	var resp changeUserBalanceResponse
	var req ChangeUserBalanceRequest

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		resp.Status = "fail"
		resp.Message = "request body is invalid"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		resp.Status = "fail"
		resp.Message = "request body's JSON format is invalid (amount: int, transaction_id: string)"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
		return
	}

	v := getValidator()
	if err := v.Struct(req); err != nil {
		resp.Status = "fail"
		invalidFields := []string{}
		for _, validErr := range err.(validator.ValidationErrors) {
			invalidFields = append(invalidFields, validErr.Field())
		}
		resp.Message = strings.Join(invalidFields, ", ") + " is requreid"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
		return
	}
	err = h.usecase.AddAllUserBalance(req.Amount, req.TransactionID)
	if err != nil {
		status, msg, httpCode := handleError(err)
		if status == "error" {
			h.App.ErrorLog.Println(err)
		}

		resp.Status = status
		resp.Message = msg
		w.WriteHeader(httpCode)
		out, _ := json.Marshal(resp)
		w.Write(out)
		return
	}

	resp.Status = "success"
	resp.Message = "user balance added successfully"
	out, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}
