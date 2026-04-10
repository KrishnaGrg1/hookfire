package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	db "github.com/KrishnaGrg1/hookfire/internal/db/sqlc"
	"github.com/KrishnaGrg1/hookfire/internal/queue"
	"github.com/KrishnaGrg1/hookfire/internal/store"
)

type contextKey string

const AppContextKey contextKey = "app"

func APIKeyAuth(s *store.Store, q *queue.Queue) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expect: Authorization: Bearer hf_yourkey
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			apiKey := strings.TrimPrefix(authHeader, "Bearer ")
			cacheKey := fmt.Sprintf("app:apikey:%s", apiKey)
			cachedData, err := q.GetCache(r.Context(), cacheKey)
			if err == nil {
				var app db.Application
				if err := json.Unmarshal([]byte(cachedData), &app); err == nil {
					ctx := context.WithValue(r.Context(), AppContextKey, app)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			// Look up app by API key

			app, err := s.Queries.GetApplicationByapikey(r.Context(), apiKey)
			if err != nil {
				http.Error(w, "invalid api key", http.StatusUnauthorized)
				return
			}

			go func() {
				q.SetCache(context.Background(), cacheKey, app, 10*time.Minute)
			}()

			// Attach app to context so handlers can use it
			ctx := context.WithValue(r.Context(), AppContextKey, app)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
