package checker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"healthcheck-api/internal/model"
)

type HTTPChecker struct {
	client *http.Client
}

func NewHTTPChecker(client *http.Client) *HTTPChecker {
	if client == nil {
		client = &http.Client{}
	}

	return &HTTPChecker{
		client: client,
	}
}

func (c *HTTPChecker) Check(ctx context.Context, rawURL string) model.CheckResult {
	start := time.Now()

	result := model.CheckResult{
		URL:       rawURL,
		CheckedAt: start.UTC(),
		Success:   false,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("failed to build request: %v", err)
		return result
	}

	resp, err := c.client.Do(req)
	result.ResponseTimeMS = time.Since(start).Milliseconds()

	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.Success = true
		return result
	}

	result.Error = fmt.Sprintf(
		"received HTTP %d %s",
		resp.StatusCode,
		http.StatusText(resp.StatusCode),
	)

	return result
}