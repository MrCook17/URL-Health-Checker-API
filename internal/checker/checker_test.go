package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPChecker_Check_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	chk := NewHTTPChecker(&http.Client{})

	result := chk.Check(context.Background(), srv.URL)

	if !result.Success {
		t.Fatalf("expected success true, got false with error %q", result.Error)
	}
	if result.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", result.StatusCode)
	}
	if result.URL != srv.URL {
		t.Fatalf("expected URL %q, got %q", srv.URL, result.URL)
	}
}

func TestHTTPChecker_Check_HTTPFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	chk := NewHTTPChecker(&http.Client{})

	result := chk.Check(context.Background(), srv.URL)

	if result.Success {
		t.Fatalf("expected success false, got true")
	}
	if result.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", result.StatusCode)
	}
	if result.Error == "" {
		t.Fatalf("expected error message, got empty string")
	}
}

func TestHTTPChecker_Check_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(700 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	chk := NewHTTPChecker(&http.Client{})

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	result := chk.Check(ctx, srv.URL)

	if result.Success {
		t.Fatalf("expected success false, got true")
	}
	if result.Error != "request timed out" {
		t.Fatalf("expected timeout error, got %q", result.Error)
	}
}