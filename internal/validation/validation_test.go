package validation

import (
	"testing"

	"healthcheck-api/internal/model"
)

func TestNormalizeCheckRequest_DefaultTimeoutAndTrim(t *testing.T) {
	req := model.CheckRequest{
		URLs: []string{
			" https://example.com ",
			"https://golang.org",
		},
	}

	err := NormalizeCheckRequest(&req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if req.TimeoutMS != model.DefaultTimeoutMS {
		t.Fatalf("expected default timeout %d, got %d", model.DefaultTimeoutMS, req.TimeoutMS)
	}

	if req.URLs[0] != "https://example.com" {
		t.Fatalf("expected trimmed URL, got %q", req.URLs[0])
	}
}

func TestNormalizeCheckRequest_InvalidCases(t *testing.T) {
	tests := []struct {
		name string
		req  model.CheckRequest
	}{
		{
			name: "empty urls",
			req:  model.CheckRequest{URLs: []string{}},
		},
		{
			name: "unsupported scheme",
			req: model.CheckRequest{
				URLs: []string{"ftp://example.com"},
			},
		},
		{
			name: "invalid url",
			req: model.CheckRequest{
				URLs: []string{"://bad-url"},
			},
		},
		{
			name: "timeout too low",
			req: model.CheckRequest{
				URLs:      []string{"https://example.com"},
				TimeoutMS: 100,
			},
		},
		{
			name: "timeout too high",
			req: model.CheckRequest{
				URLs:      []string{"https://example.com"},
				TimeoutMS: 20000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.req
			err := NormalizeCheckRequest(&req)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}