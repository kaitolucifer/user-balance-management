package presentation

import (
	"encoding/json"
	"net/http"
	"strconv"

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
		h.app.InfoLog.Println(err)
		resp.Status = "fail"
		resp.Message = err.Error()
		out, _ := json.Marshal(resp)
		w.Write(out)
		return
	}

	resp.Status = "success"
	resp.Balance = strconv.Itoa(balance)
	out, _ := json.Marshal(resp)
	w.Write(out)
}

type AddUserBalanceResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type AddUserBalanceRequest struct {
	Amount int `json:"amount"`
}

func (h *UserBalanceHandler) AddUserBalance(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	w.Header().Set("Content-Type", "application/json")

	var resp AddUserBalanceResponse
	var reqBody AddUserBalanceRequest

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

	err = h.usecase.AddBalance(userID, reqBody.Amount)
	if err != nil {
		resp.Status = "fail"
		resp.Message = err.Error()
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusNotFound)
		w.Write(out)
		return
	}

	resp.Status = "success"
	resp.Message = "user balance added successfully"
	out, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}
