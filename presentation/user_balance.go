package presentation

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/kaitolucifer/user-balance-management/domain"
)

type UserBalanceHandler struct {
	usecase domain.UserBalanceUsecase
	app     *App
}

func NewUserBalanceHander(usecase domain.UserBalanceUsecase, app *App) UserBalanceHandler {
	return UserBalanceHandler{
		usecase: usecase,
		app:     app,
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
		switch err {
		case sql.ErrNoRows:
			resp.Status = "fail"
			resp.Message = "user_id not found"
			w.WriteHeader(http.StatusNotFound)
		default:
			h.app.ErrorLog.Println(err)
			resp.Status = "error"
			resp.Message = "internal server error"
			w.WriteHeader(http.StatusInternalServerError)
		}

		out, _ := json.Marshal(resp)
		w.Write(out)
		return
	}

	resp.Status = "success"
	resp.Balance = strconv.Itoa(balance)
	out, _ := json.Marshal(resp)
	w.Write(out)
}

// changeUserBalanceResponseは残高を加減算するエンドポイントのレスポンスフォーマット
type changeUserBalanceResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// changeUserBalanceResponseは残高を加減算するエンドポイントのリクエストフォーマット
type ChangeUserBalanceRequest struct {
	Amount int `json:"amount"`
	TransactionID string `json:"transaction_id"`
}

func (h *UserBalanceHandler) ChangeUserBalance(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	w.Header().Set("Content-Type", "application/json")

	splited := strings.Split(r.RequestURI, "/")
	change_type := splited[2] // 加算か減算かを示す部分

	var resp changeUserBalanceResponse
	var reqBody ChangeUserBalanceRequest

	body := json.NewDecoder(r.Body)
	err := body.Decode(&reqBody)

	if err != nil {
		resp.Status = "fail"
		resp.Message = "request body is not valid"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
		return
	}

	if reqBody.Amount <= 0 {
		resp.Status = "fail"
		resp.Message = "amount is non-positive number"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
		return
	}

	if reqBody.TransactionID == "" {
		resp.Status = "fail"
		resp.Message = "transaction_id is empty"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
		return
	}

	if change_type == "add" {
		err = h.usecase.AddBalance(userID, reqBody.Amount, reqBody.TransactionID)
	} else {
		err = h.usecase.ReduceBalance(userID, reqBody.Amount, reqBody.TransactionID)
	}

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			resp.Status = "fail"
			resp.Message = "user_id not found"
			w.WriteHeader(http.StatusNotFound)
		default:
			h.app.ErrorLog.Println(err)
			resp.Status = "error"
			resp.Message = "internal server error"
			w.WriteHeader(http.StatusInternalServerError)
		}

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
