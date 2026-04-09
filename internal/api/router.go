package api

import (
	"encoding/json"
	"net/http"

	"github.com/KrishnaGrg1/hookfire/internal/api/handler"
	mw "github.com/KrishnaGrg1/hookfire/internal/api/middleware"
	"github.com/KrishnaGrg1/hookfire/internal/queue"
	"github.com/KrishnaGrg1/hookfire/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(s *store.Store, q *queue.Queue) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	appHandler := handler.NewApplicationHanlder(s)
	endpointHandler := handler.NewEndpointHanlder(s, q)
	eventHandler := handler.NewEventHandler(s, q)
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/applications", appHandler.Create)
		r.Group(func(r chi.Router) {
			r.Use(mw.APIKeyAuth(s))

			r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
				app := r.Context().Value(mw.AppContextKey)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(app)
			})

			r.Post("/endpoints", endpointHandler.Create)
			r.Get("/endpoints", endpointHandler.List)
			r.Delete("/endpoints/{id}", endpointHandler.Delete)

			r.Post("/events", eventHandler.Create)
			r.Get("/events", eventHandler.List)
			r.Get("/events/{id}/attempts", eventHandler.ListAttempts)
		})

	})

	return r
}
