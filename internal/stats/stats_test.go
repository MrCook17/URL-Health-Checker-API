package stats

import (
	"testing"
	"time"

	"healthcheck-api/internal/model"
)

func TestComputeJobSummary(t *testing.T) {
	results := []model.CheckResult{
		{
			URL:            "https://fast.example",
			Success:        true,
			StatusCode:     200,
			ResponseTimeMS: 50,
			CheckedAt:      time.Now().UTC(),
		},
		{
			URL:            "https://slow.example",
			Success:        false,
			StatusCode:     500,
			ResponseTimeMS: 300,
			CheckedAt:      time.Now().UTC(),
			Error:          "received HTTP 500 Internal Server Error",
		},
		{
			URL:            "https://timeout.example",
			Success:        false,
			ResponseTimeMS: 1000,
			CheckedAt:      time.Now().UTC(),
			Error:          "request timed out",
		},
	}

	summary := ComputeJobSummary(results)

	if summary.Total != 3 {
		t.Fatalf("expected total 3, got %d", summary.Total)
	}
	if summary.Successes != 1 {
		t.Fatalf("expected successes 1, got %d", summary.Successes)
	}
	if summary.Failures != 2 {
		t.Fatalf("expected failures 2, got %d", summary.Failures)
	}
	if summary.TimeoutCount != 1 {
		t.Fatalf("expected timeout count 1, got %d", summary.TimeoutCount)
	}
	if summary.FastestURL != "https://fast.example" {
		t.Fatalf("expected fastest URL https://fast.example, got %q", summary.FastestURL)
	}
	if summary.SlowestURL != "https://timeout.example" {
		t.Fatalf("expected slowest URL https://timeout.example, got %q", summary.SlowestURL)
	}
	if summary.StatusClasses["2xx"] != 1 {
		t.Fatalf("expected 2xx count 1, got %d", summary.StatusClasses["2xx"])
	}
	if summary.StatusClasses["5xx"] != 1 {
		t.Fatalf("expected 5xx count 1, got %d", summary.StatusClasses["5xx"])
	}
	if summary.StatusClasses["other"] != 1 {
		t.Fatalf("expected other count 1, got %d", summary.StatusClasses["other"])
	}
}

func TestComputeSystemStats(t *testing.T) {
	jobs := []model.CheckJob{
		{
			ID:     "chk_001",
			Status: model.StatusCompleted,
			Results: []model.CheckResult{
				{URL: "https://a.example", Success: true, StatusCode: 200, ResponseTimeMS: 100},
				{URL: "https://b.example", Success: false, StatusCode: 404, ResponseTimeMS: 200, Error: "received HTTP 404 Not Found"},
			},
		},
		{
			ID:     "chk_002",
			Status: model.StatusFailed,
			Results: []model.CheckResult{
				{URL: "https://c.example", Success: false, ResponseTimeMS: 500, Error: "request timed out"},
			},
		},
	}

	got := ComputeSystemStats(jobs)

	if got.TotalJobs != 2 {
		t.Fatalf("expected total jobs 2, got %d", got.TotalJobs)
	}
	if got.CompletedJobs != 1 {
		t.Fatalf("expected completed jobs 1, got %d", got.CompletedJobs)
	}
	if got.FailedJobs != 1 {
		t.Fatalf("expected failed jobs 1, got %d", got.FailedJobs)
	}
	if got.TotalURLsChecked != 3 {
		t.Fatalf("expected total URLs checked 3, got %d", got.TotalURLsChecked)
	}
	if got.SuccessfulChecks != 1 {
		t.Fatalf("expected successful checks 1, got %d", got.SuccessfulChecks)
	}
	if got.FailedChecks != 2 {
		t.Fatalf("expected failed checks 2, got %d", got.FailedChecks)
	}
	if got.TimeoutCount != 1 {
		t.Fatalf("expected timeout count 1, got %d", got.TimeoutCount)
	}
	if got.StatusClasses["2xx"] != 1 {
		t.Fatalf("expected 2xx count 1, got %d", got.StatusClasses["2xx"])
	}
	if got.StatusClasses["4xx"] != 1 {
		t.Fatalf("expected 4xx count 1, got %d", got.StatusClasses["4xx"])
	}
	if got.StatusClasses["other"] != 1 {
		t.Fatalf("expected other count 1, got %d", got.StatusClasses["other"])
	}
}