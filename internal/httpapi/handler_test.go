package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"healthcheck-api/internal/checker"
	"healthcheck-api/internal/model"
	"healthcheck-api/internal/store"
)

func newTestMux() *http.ServeMux {
	st := store.NewMemoryStore()
	chk := checker.NewHTTPChecker(&http.Client{})
	h := NewHandler(st, chk)

	mux := http.NewServeMux()
	h.Register(mux)
	return mux
}

func TestHandleHealth(t *testing.T) {
	mux := newTestMux()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestCreateCheckAndGetByID(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	mux := newTestMux()

	body := []byte(`{"urls":["` + target.URL + `"],"timeout_ms":1000}`)
	req := httptest.NewRequest(http.MethodPost, "/checks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	var created model.CheckJob
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	if created.ID == "" {
		t.Fatalf("expected job ID, got empty")
	}
	if created.Status != model.StatusCompleted {
		t.Fatalf("expected status completed, got %q", created.Status)
	}
	if len(created.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(created.Results))
	}

	getReq := httptest.NewRequest(http.MethodGet, "/checks/"+created.ID, nil)
	getRR := httptest.NewRecorder()

	mux.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getRR.Code)
	}

	var fetched model.CheckJob
	if err := json.NewDecoder(getRR.Body).Decode(&fetched); err != nil {
		t.Fatalf("failed to decode get response: %v", err)
	}

	if fetched.ID != created.ID {
		t.Fatalf("expected fetched ID %q, got %q", created.ID, fetched.ID)
	}
}

func TestCreateCheck_InvalidJSON(t *testing.T) {
	mux := newTestMux()

	req := httptest.NewRequest(http.MethodPost, "/checks", bytes.NewReader([]byte(`{"urls":`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	var errResp model.ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error == "" {
		t.Fatalf("expected structured error code, got empty")
	}
}

func TestHandleStats(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	mux := newTestMux()

	for i := 0; i < 2; i++ {
		body := []byte(`{"urls":["` + target.URL + `"],"timeout_ms":1000}`)
		req := httptest.NewRequest(http.MethodPost, "/checks", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rr.Code)
		}
	}

	statsReq := httptest.NewRequest(http.MethodGet, "/stats", nil)
	statsRR := httptest.NewRecorder()

	mux.ServeHTTP(statsRR, statsReq)

	if statsRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", statsRR.Code)
	}

	var statsResp model.StatsResponse
	if err := json.NewDecoder(statsRR.Body).Decode(&statsResp); err != nil {
		t.Fatalf("failed to decode stats response: %v", err)
	}

	if statsResp.TotalJobs != 2 {
		t.Fatalf("expected total jobs 2, got %d", statsResp.TotalJobs)
	}
	if statsResp.TotalURLsChecked != 2 {
		t.Fatalf("expected total URLs checked 2, got %d", statsResp.TotalURLsChecked)
	}
	if statsResp.SuccessfulChecks != 2 {
		t.Fatalf("expected successful checks 2, got %d", statsResp.SuccessfulChecks)
	}
}

func TestCreateCheck_TimeoutResult(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(700 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer slow.Close()

	mux := newTestMux()

	body := []byte(`{"urls":["` + slow.URL + `"],"timeout_ms":500}`)
	req := httptest.NewRequest(http.MethodPost, "/checks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	var job model.CheckJob
	if err := json.NewDecoder(rr.Body).Decode(&job); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(job.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(job.Results))
	}
	if job.Results[0].Error != "request timed out" {
		t.Fatalf("expected timeout error, got %q", job.Results[0].Error)
	}
}