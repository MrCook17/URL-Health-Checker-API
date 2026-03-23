package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"log/slog"

	"healthcheck-api/internal/checker"
	"healthcheck-api/internal/model"
	"healthcheck-api/internal/stats"
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
	mux.HandleFunc("/stats", h.handleStats)
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
	case http.MethodGet:
		h.listChecks(w, r)
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

func (h *Handler) listChecks(w http.ResponseWriter, r *http.Request) {
	jobs := h.store.List()
	writeJSON(w, http.StatusOK, jobs)
}

func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	jobs := h.store.List()
	response := stats.ComputeSystemStats(jobs)
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) createCheck(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := RequestIDFromContext(r.Context())

	defer r.Body.Close()

	var req model.CheckRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		slog.Default().WarnContext(r.Context(), "check_request_invalid_json",
			"request_id", requestID,
			"error", err.Error(),
		)
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		slog.Default().WarnContext(r.Context(), "check_request_extra_json",
			"request_id", requestID,
		)
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must contain a single JSON object")
		return
	}

	if err := validation.NormalizeCheckRequest(&req); err != nil {
		slog.Default().WarnContext(r.Context(), "check_request_validation_failed",
			"request_id", requestID,
			"error", err.Error(),
		)
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
		Summary:   stats.ComputeJobSummary(nil),
	}

	h.store.Create(job)

	slog.Default().InfoContext(r.Context(), "job_created",
		"request_id", requestID,
		"job_id", job.ID,
		"url_count", len(job.URLs),
		"timeout_ms", job.TimeoutMS,
	)

	job.Status = model.StatusRunning
	h.store.Update(job)

	results := h.runChecksConcurrently(r.Context(), job.URLs, job.TimeoutMS)

	job.Results = results
	job.Summary = stats.ComputeJobSummary(results)

	if err := r.Context().Err(); err != nil {
		job.Status = model.StatusFailed
	} else {
		job.Status = model.StatusCompleted
	}

	h.store.Update(job)

	slog.Default().InfoContext(r.Context(), "job_finished",
		"request_id", requestID,
		"job_id", job.ID,
		"status", job.Status,
		"successes", job.Summary.Successes,
		"failures", job.Summary.Failures,
		"timeout_count", job.Summary.TimeoutCount,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	if err := r.Context().Err(); err != nil {
		slog.Default().WarnContext(r.Context(), "response_not_written_client_canceled",
			"request_id", requestID,
			"job_id", job.ID,
			"error", err.Error(),
		)
		return
	}

	w.Header().Set("Location", "/checks/"+job.ID)
	writeJSON(w, http.StatusCreated, job)
}

func (h *Handler) runChecksConcurrently(parentCtx context.Context, urls []string, timeoutMS int) []model.CheckResult {
	results := make([]model.CheckResult, len(urls))
	resultsCh := make(chan indexedResult, len(urls))

	var wg sync.WaitGroup
	wg.Add(len(urls))

	timeout := time.Duration(timeoutMS) * time.Millisecond
	requestID := RequestIDFromContext(parentCtx)

	for i, rawURL := range urls {
		i := i
		rawURL := rawURL

		go func() {
			defer wg.Done()

			slog.Default().InfoContext(parentCtx, "check_started",
				"request_id", requestID,
				"url", rawURL,
				"timeout_ms", timeoutMS,
			)

			checkCtx, cancel := context.WithTimeout(parentCtx, timeout)
			defer cancel()

			result := h.checker.Check(checkCtx, rawURL)

			if result.Success {
				slog.Default().InfoContext(parentCtx, "check_completed",
					"request_id", requestID,
					"url", rawURL,
					"status_code", result.StatusCode,
					"response_time_ms", result.ResponseTimeMS,
				)
			} else {
				slog.Default().WarnContext(parentCtx, "check_failed",
					"request_id", requestID,
					"url", rawURL,
					"status_code", result.StatusCode,
					"response_time_ms", result.ResponseTimeMS,
					"error", result.Error,
				)
			}

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