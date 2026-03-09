package validation

import (
	"fmt"
	"net/url"
	"strings"

	"healthcheck-api/internal/model"
)

// NormalizeCheckRequest validates the incoming request and applies defaults.
func NormalizeCheckRequest(req *model.CheckRequest) error {
	if len(req.URLs) == 0 {
		return fmt.Errorf("urls must contain at least one valid URL")
	}

	if len(req.URLs) > model.MaxURLsPerJob {
		return fmt.Errorf("urls must contain no more than %d entries", model.MaxURLsPerJob)
	}

	// Trim and validate each submitted URL.
	cleaned := make([]string, 0, len(req.URLs))
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

		cleaned = append(cleaned, rawURL)
	}

	req.URLs = cleaned

	// Apply the default timeout when omitted.
	if req.TimeoutMS == 0 {
		req.TimeoutMS = model.DefaultTimeoutMS
	}

	if req.TimeoutMS < model.MinTimeoutMS || req.TimeoutMS > model.MaxTimeoutMS {
		return fmt.Errorf(
			"timeout_ms must be between %d and %d",
			model.MinTimeoutMS,
			model.MaxTimeoutMS,
		)
	}

	return nil
}