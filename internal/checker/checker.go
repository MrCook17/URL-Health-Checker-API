package checker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"healthcheck-api/internal/model"
)

// HTTPChecker performs single-URL HTTP health checks.
type HTTPChecker struct {
	client *http.Client
}

// NewHTTPChecker creates a checker with a reusable HTTP client.
func NewHTTPChecker(client *http.Client) *HTTPChecker {
	if client == nil {
		client = &http.Client{}
	}

	return &HTTPChecker{
		client: client,
	}
}

// Check performs one outbound HTTP GET request and returns a CheckResult.
// The provided context controls timeout and cancellation for the whole request.
func (c *HTTPChecker) Check(ctx context.Context, rawURL string) (result model.CheckResult) {
	start := time.Now()

	result = model.CheckResult{
		URL:       rawURL,
		CheckedAt: start.UTC(),
		Success:   false,
	}

	// Always record elapsed time, no matter how the check exits.
	defer func() {
		result.ResponseTimeMS = time.Since(start).Milliseconds()
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("failed to build request: %v", err)
		return
	}

	resp, err := c.client.Do(req)
	if err != nil {
		// Classify timeout and cancellation separately for clearer API output.
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			result.Error = "request timed out"
		case errors.Is(err, context.Canceled):
			result.Error = "request canceled"
		default:
			result.Error = err.Error()
		}
		return
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.Success = true
		return
	}

	result.Error = fmt.Sprintf(
		"received HTTP %d %s",
		resp.StatusCode,
		http.StatusText(resp.StatusCode),
	)

	return
}