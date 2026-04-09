package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	mw "github.com/KrishnaGrg1/hookfire/internal/api/middleware"
	db "github.com/KrishnaGrg1/hookfire/internal/db/sqlc"
	"github.com/KrishnaGrg1/hookfire/internal/helper"
	"github.com/KrishnaGrg1/hookfire/internal/queue"
	"github.com/KrishnaGrg1/hookfire/internal/store"
	"github.com/go-chi/chi/v5"
)

type EventHandler struct {
	store *store.Store
	queue *queue.Queue
}

func NewEventHandler(s *store.Store, q *queue.Queue) *EventHandler {
	return &EventHandler{store: s, queue: q}
}

func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	app := r.Context().Value(mw.AppContextKey).(db.Application)

	// 1. Parse request body
	var input struct {
		EventType string          `json:"event_type"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		helper.Error(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_001", "Request body must be valid JSON")
		return
	}
	if input.EventType == "" {
		helper.Error(w, http.StatusBadRequest, "Validation failed", "VALIDATION_002", "event_type is required")
		return
	}

	event, err := h.store.Queries.CreateEvent(r.Context(), db.CreateEventParams{
		AppID:     app.ID,
		EventType: input.EventType,
		Payload:   input.Payload,
	})
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Failed to create event", "INTERNAL_001", "Database insert failed")
		return
	}

	// 3. Get all active endpoints for this app
	endpoints, err := h.store.Queries.ListEndpointsByApp(r.Context(), app.ID)
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Failed to fetch endpoints", "INTERNAL_002", "Database query failed")
		return
	}

	// 4. Push a delivery job for each endpoint into Redis
	// This is instant — workers handle actual delivery in background
	for _, ep := range endpoints {
		job := queue.Job{
			EventID:    event.ID,
			EndpointID: ep.ID,
			AttemptNum: 1,
		}
		h.queue.Enqueue(r.Context(), job)
	}

	// 5. Respond immediately — don't wait for delivery
	helper.Success(w, http.StatusAccepted, "Event received", event)
}

func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	app := r.Context().Value(mw.AppContextKey).(db.Application)

	events, err := h.store.Queries.ListEventsByApp(r.Context(), app.ID)
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Failed to fetch events", "INTERNAL_003", "Database query failed")
		return
	}

	helper.Success(w, http.StatusOK, "Events retrieved successfully", events)
}

func (h *EventHandler) ListAttempts(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		helper.Error(w, http.StatusBadRequest, "Invalid event id", "VALIDATION_003", "id must be an integer")
		return
	}
	attempts, err := h.store.Queries.ListAttemptsByEvent(r.Context(), id)
	if err != nil {
		helper.Error(w, http.StatusInternalServerError, "Failed to fetch attempts", "INTERNAL_004", "Database query failed")
		return
	}

	helper.Success(w, http.StatusOK, "Attempts retrieved successfully", attempts)
}
