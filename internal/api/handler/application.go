package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	db "github.com/KrishnaGrg1/hookfire/internal/db/sqlc"
	"github.com/KrishnaGrg1/hookfire/internal/helper"
	"github.com/KrishnaGrg1/hookfire/internal/queue"
	"github.com/KrishnaGrg1/hookfire/internal/store"
)

type ApplicationHanlder struct {
	store *store.Store
	queue *queue.Queue
}

func NewApplicationHanlder(s *store.Store, q *queue.Queue) *ApplicationHanlder {
	return &ApplicationHanlder{store: s, queue: q}
}

func generateApiKey() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "hf_" + hex.EncodeToString(bytes), nil
}

func (h *ApplicationHanlder) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Parse request body
	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		helper.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_001", "Request body must be valid JSON")
		return
	}
	if input.Name == "" {
		helper.Error(w, http.StatusBadRequest, "Validation failed", "VALIDATION_002", "name is required")
		return
	}
	// 2. Generate API key
	apiKey, err := generateApiKey()
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Failed to create application", "INTERNAL_001", "Could not generate API key")
		return
	}
	// 3. Save to database using sqlc generated code
	app, err := h.store.Queries.CreateApplication(r.Context(), db.CreateApplicationParams{
		Name:   input.Name,
		ApiKey: apiKey,
	})
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Failed to create application", "INTERNAL_002", "Database insert failed")
		return
	}

	cacheKey := fmt.Sprintf("app:apikey:%s", app.ApiKey)
	go func() {
		h.queue.SetCache(context.Background(), cacheKey, app, 10*time.Minute)
	}()

	helper.Success(w, http.StatusCreated, "Application created successfully", app)
}
