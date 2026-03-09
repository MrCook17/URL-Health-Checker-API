package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"healthcheck-api/internal/model"
	"healthcheck-api/internal/store"
	"healthcheck-api/internal/validation"
)

// Handler contains API dependencies and shared request state.
type Handler struct {
	store  *store.MemoryStore
	nextID atomic.Uint64
}

// NewHandler creates a new HTTP API handler.
func NewHandler(store *store.MemoryStore) *Handler {
	return &Handler{
		store: store,
	}
}

// Register attaches API routes to the provided mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/checks", h.handleChecks)
	mux.HandleFunc("/checks/", h.handleCheckByID)
}

// handleHealth provides a simple liveness check.
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// handleChecks dispatches requests for the /checks endpoint.
func (h *Handler) handleChecks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createCheck(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	}
}

// handleCheckByID returns one stored check job by ID.
func (h *Handler) handleCheckByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/checks/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "not_found", "check job not found")
		return
	}

	job, ok := h.store.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "check job not found")
		return
	}

	writeJSON(w, http.StatusOK, job)
}

// createCheck parses, validates, stores, and returns a new check job.
func (h *Handler) createCheck(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req model.CheckRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}

	// Reject trailing JSON or extra data after the first object.
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must contain a single JSON object")
		return
	}

	// Validate input and apply defaults such as timeout.
	if err := validation.NormalizeCheckRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	id := fmt.Sprintf("chk_%03d", h.nextID.Add(1))

	job := model.CheckJob{
		ID:        id,
		CreatedAt: time.Now().UTC(),
		Status:    model.StatusPending,
		URLs:      req.URLs,
		TimeoutMS: req.TimeoutMS,
		Results:   []model.CheckResult{},
		Summary: model.Summary{
			Total:     len(req.URLs),
			Successes: 0,
			Failures:  0,
		},
	}

	h.store.Create(job)

	// Tell the client where the newly created job can be fetched.
	w.Header().Set("Location", "/checks/"+job.ID)
	writeJSON(w, http.StatusCreated, job)
}

// writeJSON sends a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}

// writeError sends a consistent JSON error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, model.ErrorResponse{
		Error:   code,
		Message: message,
	})
}