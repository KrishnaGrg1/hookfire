package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	mw "github.com/KrishnaGrg1/hookfire/internal/api/middleware"
	db "github.com/KrishnaGrg1/hookfire/internal/db/sqlc"
	"github.com/KrishnaGrg1/hookfire/internal/helper"
	"github.com/KrishnaGrg1/hookfire/internal/queue"
	"github.com/KrishnaGrg1/hookfire/internal/store"
	"github.com/go-chi/chi/v5"
)

type EndpointHandler struct {
	store *store.Store
	queue *queue.Queue
}

func NewEndpointHanlder(s *store.Store, q *queue.Queue) *EndpointHandler {
	return &EndpointHandler{
		store: s,
		queue: q,
	}
}

func generateSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (h *EndpointHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Get the app from context (set by auth middleware)
	app := r.Context().Value(mw.AppContextKey).(db.Application)
	cacheKey := fmt.Sprintf("endpoints:app:%d", app.ID)
	var input struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		helper.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_001", "Request body must be valid JSON")
		return
	}
	if input.URL == "" {
		helper.Error(w, http.StatusBadRequest, "Validation failed", "VALIDATION_002", "url is required")
		return
	}

	secret, err := generateSecret()
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Failed to create endpoint", "INTERNAL_001", "Could not generate endpoint secret")
		return
	}

	endpoint, err := h.store.Queries.CreateEndpoint(r.Context(), db.CreateEndpointParams{
		AppID:  app.ID,
		Url:    input.URL,
		Secret: secret,
	})
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Failed to create endpoint", "INTERNAL_002", "Database insert failed")
		return
	}
	go func() {
		h.queue.SetCache(context.Background(), cacheKey, endpoint, 10*time.Minute)
	}()
	helper.Success(w, http.StatusCreated, "Endpoint created successfully", endpoint)
}

func (h *EndpointHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	app := r.Context().Value(mw.AppContextKey).(db.Application)

	cacheKey := fmt.Sprintf("endpoints:app:%d", app.ID)

	// 2. Try the "Fast Path"
	cachedData, err := h.queue.GetCache(ctx, cacheKey)
	if err == nil {
		var endpoints []db.Endpoint
		if err := json.Unmarshal([]byte(cachedData), &endpoints); err == nil {
			helper.Success(w, http.StatusOK, "Retrieved (Cache)", endpoints)
			return
		}
	}
	endpoints, err := h.store.Queries.ListEndpointsByApp(r.Context(), app.ID)
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Failed to fetch endpoints", "INTERNAL_003", "Database query failed")
		return
	}
	go func() {
		h.queue.SetCache(context.Background(), cacheKey, endpoints, 10*time.Minute)
	}()
	helper.Success(w, http.StatusOK, "Endpoints retrieved successfully", endpoints)
}

func (h *EndpointHandler) Delete(w http.ResponseWriter, r *http.Request) {
	app := r.Context().Value(mw.AppContextKey).(db.Application)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		helper.Error(w, http.StatusBadRequest, "Invalid endpoint id", "VALIDATION_003", "id must be an integer")
		return
	}

	endpoint, err := h.store.Queries.GetEndpointByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			helper.Error(w, http.StatusNotFound, "Endpoint not found", "ENDPOINT_404", fmt.Sprintf("No endpoint found with id %d", id))
			return
		}
		helper.Error(w, http.StatusInternalServerError, "Failed to fetch endpoint", "INTERNAL_004", "Database lookup failed")
		return
	}

	if endpoint.AppID != app.ID || !endpoint.IsActive {
		helper.Error(w, http.StatusNotFound, "Endpoint not found", "ENDPOINT_404", fmt.Sprintf("No active endpoint found with id %d", id))
		return
	}

	err = h.store.Queries.DeleteEndpoint(r.Context(), db.DeleteEndpointParams{
		ID:    id,
		AppID: app.ID,
	})

	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Failed to delete endpoint", "INTERNAL_005", "Database update failed")
		return
	}

	helper.Success(w, http.StatusOK, fmt.Sprintf("Endpoint %d deleted successfully", id), nil)
}
