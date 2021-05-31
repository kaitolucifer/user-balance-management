package presentation

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func Routes(handler UserBalanceHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	r.Get("/", handler.Home)
	r.Get("/balance/{userID}", handler.UserBalance)
	r.Patch("/balance/add/{userID}", handler.ChangeUserBalance)
	r.Patch("/balance/reduce/{userID}", handler.ChangeUserBalance)
	r.Patch("/balance/add-all", handler.AddAllUserBalance)

	return r
}
