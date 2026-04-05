package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/KrishnaGrg1/hookfire/internal/store"
)

type contextKey string

const AppContextKey contextKey = "app"

func APIKeyAuth(s *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expect: Authorization: Bearer hf_yourkey
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			apiKey := strings.TrimPrefix(authHeader, "Bearer ")

			// Look up app by API key
			app, err := s.Queries.GetApplicationByapikey(r.Context(), apiKey)
			if err != nil {
				http.Error(w, "invalid api key", http.StatusUnauthorized)
				return
			}

			// Attach app to context so handlers can use it
			ctx := context.WithValue(r.Context(), AppContextKey, app)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
