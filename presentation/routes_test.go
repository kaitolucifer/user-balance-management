package presentation

import (
	"testing"

	"github.com/go-chi/chi"
)

func TestRoutes(t *testing.T) {
	handler := new(UserBalanceHandler)
	mux := Routes(handler)
	switch v := mux.(type) {
	case *chi.Mux:
	default:
		t.Errorf("return type is not [*chi.Mux], got [%s]", v)
	}
}
