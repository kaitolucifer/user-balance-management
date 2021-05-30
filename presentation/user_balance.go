package presentation

import (
	"net/http"

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
