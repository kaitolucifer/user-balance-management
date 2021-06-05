package presentation

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Routes chiを使用してHTTP用のmultiplexerを作成する
func Routes(handler *RestfulUserBalanceHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	r.Get("/", handler.HealthCheck)
	r.NotFound(handler.NotFound)
	r.Get("/balance/{userID}", handler.GetUserBalance)
	r.Patch("/balance/add/{userID}", handler.ChangeUserBalance)
	r.Patch("/balance/reduce/{userID}", handler.ChangeUserBalance)
	r.Patch("/balance/add-all", handler.AddAllUserBalance)

	return r
}
