package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"healthcheck-api/internal/model"
	"healthcheck-api/internal/store"
)

type Handler struct {
	store  *store.MemoryStore
	nextID atomic.Uint64
}

func NewHandler(store *store.MemoryStore) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/checks", h.handleChecks)
	mux.HandleFunc("/checks/", h.handleCheckByID)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) handleChecks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createCheck(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	}
}

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

func (h *Handler) createCheck(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req model.CheckRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}

	if err := validateCheckRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if req.TimeoutMS == 0 {
		req.TimeoutMS = 3000
	}

	id := fmt.Sprintf("chk_%03d", h.nextID.Add(1))

	job := model.CheckJob{
		ID:        id,
		CreatedAt: time.Now().UTC(),
		Status:    "accepted",
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
	writeJSON(w, http.StatusCreated, job)
}

func validateCheckRequest(req *model.CheckRequest) error {
	if len(req.URLs) == 0 {
		return fmt.Errorf("urls must contain at least one valid URL")
	}

	if len(req.URLs) > 20 {
		return fmt.Errorf("urls must contain no more than 20 entries")
	}

	for _, rawURL := range req.URLs {
		rawURL = strings.TrimSpace(rawURL)
		if rawURL == "" {
			return fmt.Errorf("urls must not contain empty values")
		}

		u, err := url.Parse(rawURL)
		if err != nil || u.Host == "" {
			return fmt.Errorf("invalid URL: %s", rawURL)
		}

		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("only http and https URLs are allowed")
		}
	}

	if req.TimeoutMS != 0 && (req.TimeoutMS < 500 || req.TimeoutMS > 10000) {
		return fmt.Errorf("timeout_ms must be between 500 and 10000")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, model.ErrorResponse{
		Error:   code,
		Message: message,
	})
}