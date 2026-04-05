package api

import (
	"encoding/json"
	"net/http"

	"github.com/KrishnaGrg1/hookfire/internal/api/handler"
	mw "github.com/KrishnaGrg1/hookfire/internal/api/middleware"
	"github.com/KrishnaGrg1/hookfire/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(s *store.Store) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	appHandler := handler.NewApplicationHanlder(s)

	r.Post("/api/v1/applications", appHandler.Create)

	r.Group(func(r chi.Router) {
		r.Use(mw.APIKeyAuth(s))

		r.Get("/api/v1/me", func(w http.ResponseWriter, r *http.Request) {
			app := r.Context().Value(mw.AppContextKey)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(app)
		})
	})
	return r
}
