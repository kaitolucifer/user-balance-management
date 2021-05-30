package presentation

import (
	"net/http"

	"github.com/kaitolucifer/user-balance-management/domain"
)

type UserBalanceHandler struct {
	usecase domain.UserBalanceUsecase
}

func NewUserBalanceHander(usecase domain.UserBalanceUsecase) UserBalanceHandler {
	return UserBalanceHandler{
		usecase: usecase,
	}
}

func (h *UserBalanceHandler) Home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}
