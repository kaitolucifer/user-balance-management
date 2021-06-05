package presentation

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

// UserBalanceHandler usecaseとアプリケーション設定を格納
type RestfulUserBalanceHandler struct {
	usecase domain.UserBalanceUsecase
	App     *App
}

// NewUserBalanceHander 新しいRESTfulハンドラを作成
func NewRestfulUserBalanceHander(usecase domain.UserBalanceUsecase, app *App) *RestfulUserBalanceHandler {
	return &RestfulUserBalanceHandler{
		usecase: usecase,
		App:     app,
	}
}

// HealthCheck ヘルスチェック用ハンドラ
func (h *RestfulUserBalanceHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := []byte(`{"status": "success", "message": "healthy"}`)
	w.Write(resp)
}

// NotFound 404時のハンドラ
func (h *RestfulUserBalanceHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	respJSON := fmt.Sprintf(`{"status": "error", "message": "The requested URL %s was not found on this server."}`,
		r.RequestURI)
	resp := []byte(respJSON)
	w.WriteHeader(http.StatusNotFound)
	w.Write(resp)
}

// getUserBalanceResponse 残高を参照するエンドポイントのレスポンスフォーマット
type getUserBalanceResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Balance *int    `json:"balance,omitempty"`
}

// GetUserBalance ユーザーIDでの残高を取得するハンドラ
func (h *RestfulUserBalanceHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	w.Header().Set("Content-Type", "application/json")
	var resp getUserBalanceResponse

	if userID == "" {
		resp.Status = "fail"
		resp.Message = "user_id is empty"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
		return
	}

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
	resp.Balance = &balance
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
	Amount        *int    `json:"amount" validate:"required"`
	TransactionID string `json:"transaction_id" validate:"required"`
}

// ChangeUserBalance ユーザーIDでの残高加減算処理を扱うハンドラ
func (h *RestfulUserBalanceHandler) ChangeUserBalance(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	w.Header().Set("Content-Type", "application/json")
	var resp changeUserBalanceResponse
	var req ChangeUserBalanceRequest

	if userID == "" {
		resp.Status = "fail"
		resp.Message = "user_id is empty"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
		return
	}

	splited := strings.Split(r.RequestURI, "/")
	change_type := splited[2] // 加算か減算かを示す部分

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
		resp.Message = strings.Join(invalidFields, ", ") + " can't be null"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
		return
	}

	if *req.Amount <= 0 {
		resp.Status = "fail"
		resp.Message = "amount must be positive"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(out)
		return
	}

	if change_type == "add" {
		err = h.usecase.AddBalance(userID, *req.Amount, req.TransactionID)
	} else {
		err = h.usecase.ReduceBalance(userID, *req.Amount, req.TransactionID)
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
	resp.Message = "user balance has been added successfully"
	out, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

// AddAllUserBalance 残高の一斉加算処理を扱うハンドラ
func (h *RestfulUserBalanceHandler) AddAllUserBalance(w http.ResponseWriter, r *http.Request) {
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
		resp.Message = strings.Join(invalidFields, ", ") + " can't be null"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(out)
		return
	}

	if *req.Amount <= 0 {
		resp.Status = "fail"
		resp.Message = "amount must be positive"
		out, _ := json.Marshal(resp)
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(out)
		return
	}

	err = h.usecase.AddAllUserBalance(*req.Amount, req.TransactionID)
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
	resp.Message = "user balance has been added successfully"
	out, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}
