package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"healthcheck-api/internal/checker"
	"healthcheck-api/internal/store"
)

func BenchmarkCheckURLsSequential(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	urls := []string{srv.URL, srv.URL, srv.URL, srv.URL, srv.URL}
	chk := checker.NewHTTPChecker(&http.Client{})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, rawURL := range urls {
			_ = chk.Check(context.Background(), rawURL)
		}
	}
}

func BenchmarkCheckURLsConcurrent(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	urls := []string{srv.URL, srv.URL, srv.URL, srv.URL, srv.URL}

	st := store.NewMemoryStore()
	chk := checker.NewHTTPChecker(&http.Client{})
	h := NewHandler(st, chk)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = h.runChecksConcurrently(context.Background(), urls, 3000)
	}
}