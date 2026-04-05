package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	db "github.com/KrishnaGrg1/hookfire/internal/db/sqlc"
	"github.com/KrishnaGrg1/hookfire/internal/store"
)

type ApplicationHanlder struct {
	store *store.Store
}

func NewApplicationHanlder(s *store.Store) *ApplicationHanlder {
	return &ApplicationHanlder{store: s}
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
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if input.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	// 2. Generate API key
	apiKey, err := generateApiKey()
	if err != nil {
		http.Error(w, "failed to generate api key", http.StatusInternalServerError)
		return
	}
	// 3. Save to database using sqlc generated code
	app, err := h.store.Queries.CreateApplication(r.Context(), db.CreateApplicationParams{
		Name:   input.Name,
		ApiKey: apiKey,
	})
	if err != nil {
		http.Error(w, "failed to create application", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
}
