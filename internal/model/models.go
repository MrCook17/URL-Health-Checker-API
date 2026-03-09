package model

import "time"

// CheckRequest is the request body for POST /checks.
type CheckRequest struct {
	URLs      []string `json:"urls"`
	TimeoutMS int      `json:"timeout_ms,omitempty"`
}

// CheckResult stores the outcome of checking a single URL.
type CheckResult struct {
	URL            string    `json:"url"`
	Success        bool      `json:"success"`
	StatusCode     int       `json:"status_code,omitempty"`
	ResponseTimeMS int64     `json:"response_time_ms,omitempty"`
	CheckedAt      time.Time `json:"checked_at,omitempty"`
	Error          string    `json:"error,omitempty"`
}

// Summary provides aggregate job statistics.
type Summary struct {
	Total     int `json:"total"`
	Successes int `json:"successes"`
	Failures  int `json:"failures"`
}

// CheckJob represents one submitted health-check job.
type CheckJob struct {
	ID        string        `json:"id"`
	CreatedAt time.Time     `json:"created_at"`
	Status    string        `json:"status"`
	URLs      []string      `json:"urls"`
	TimeoutMS int           `json:"timeout_ms"`
	Results   []CheckResult `json:"results"`
	Summary   Summary       `json:"summary"`
}

// ErrorResponse is a standard JSON error payload.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}