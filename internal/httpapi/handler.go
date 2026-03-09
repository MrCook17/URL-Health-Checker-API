package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"healthcheck-api/internal/checker"
	"healthcheck-api/internal/model"
	"healthcheck-api/internal/store"
	"healthcheck-api/internal/validation"
)

type Handler struct {
	store   *store.MemoryStore
	checker *checker.HTTPChecker
	nextID  atomic.Uint64
}

type indexedResult struct {
	index  int
	result model.CheckResult
}

func NewHandler(store *store.MemoryStore, checker *checker.HTTPChecker) *Handler {
	return &Handler{
		store:   store,
		checker: checker,
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

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must contain a single JSON object")
		return
	}

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

	job.Status = model.StatusRunning
	h.store.Update(job)

	results := h.runChecksConcurrently(r, job.URLs)

	job.Results = results
	job.Summary = summarizeResults(results)
	job.Status = model.StatusCompleted

	h.store.Update(job)

	w.Header().Set("Location", "/checks/"+job.ID)
	writeJSON(w, http.StatusCreated, job)
}

func (h *Handler) runChecksConcurrently(r *http.Request, urls []string) []model.CheckResult {
	results := make([]model.CheckResult, len(urls))
	resultsCh := make(chan indexedResult, len(urls))

	var wg sync.WaitGroup
	wg.Add(len(urls))

	for i, rawURL := range urls {
		i := i
		rawURL := rawURL

		go func() {
			defer wg.Done()

			result := h.checker.Check(r.Context(), rawURL)

			resultsCh <- indexedResult{
				index:  i,
				result: result,
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	for item := range resultsCh {
		results[item.index] = item.result
	}

	return results
}

func summarizeResults(results []model.CheckResult) model.Summary {
	summary := model.Summary{
		Total: len(results),
	}

	for _, result := range results {
		if result.Success {
			summary.Successes++
		} else {
			summary.Failures++
		}
	}

	return summary
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