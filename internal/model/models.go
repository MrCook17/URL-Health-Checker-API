package model

import "time"

type CheckRequest struct {
	URLs      []string `json:"urls"`
	TimeoutMS int      `json:"timeout_ms,omitempty"`
}

type CheckResult struct {
	URL            string    `json:"url"`
	Success        bool      `json:"success"`
	StatusCode     int       `json:"status_code,omitempty"`
	ResponseTimeMS int64     `json:"response_time_ms,omitempty"`
	CheckedAt      time.Time `json:"checked_at,omitempty"`
	Error          string    `json:"error,omitempty"`
}

type Summary struct {
	Total     int `json:"total"`
	Successes int `json:"successes"`
	Failures  int `json:"failures"`
}

type CheckJob struct {
	ID        string        `json:"id"`
	CreatedAt time.Time     `json:"created_at"`
	Status    string        `json:"status"`
	URLs      []string      `json:"urls"`
	TimeoutMS int           `json:"timeout_ms"`
	Results   []CheckResult `json:"results"`
	Summary   Summary       `json:"summary"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}