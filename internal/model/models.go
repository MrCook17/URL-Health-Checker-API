package model

import "time"

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"

	DefaultTimeoutMS = 3000
	MinTimeoutMS     = 500
	MaxTimeoutMS     = 10000
	MaxURLsPerJob    = 20
)

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
	Total             int            `json:"total"`
	Successes         int            `json:"successes"`
	Failures          int            `json:"failures"`
	AverageLatencyMS  float64        `json:"average_latency_ms"`
	FastestURL        string         `json:"fastest_url,omitempty"`
	FastestResponseMS int64          `json:"fastest_response_ms,omitempty"`
	SlowestURL        string         `json:"slowest_url,omitempty"`
	SlowestResponseMS int64          `json:"slowest_response_ms,omitempty"`
	TimeoutCount      int            `json:"timeout_count"`
	StatusClasses     map[string]int `json:"status_classes"`
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

type StatsResponse struct {
	TotalJobs          int            `json:"total_jobs"`
	PendingJobs        int            `json:"pending_jobs"`
	RunningJobs        int            `json:"running_jobs"`
	CompletedJobs      int            `json:"completed_jobs"`
	FailedJobs         int            `json:"failed_jobs"`
	TotalURLsChecked   int            `json:"total_urls_checked"`
	SuccessfulChecks   int            `json:"successful_checks"`
	FailedChecks       int            `json:"failed_checks"`
	AverageLatencyMS   float64        `json:"average_latency_ms"`
	SuccessRatePercent float64        `json:"success_rate_percent"`
	TimeoutCount       int            `json:"timeout_count"`
	StatusClasses      map[string]int `json:"status_classes"`
	FastestURL         string         `json:"fastest_url,omitempty"`
	FastestResponseMS  int64          `json:"fastest_response_ms,omitempty"`
	SlowestURL         string         `json:"slowest_url,omitempty"`
	SlowestResponseMS  int64          `json:"slowest_response_ms,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}