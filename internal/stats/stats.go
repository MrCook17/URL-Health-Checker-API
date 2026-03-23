package stats

import (
	"strings"

	"healthcheck-api/internal/model"
)

func ComputeJobSummary(results []model.CheckResult) model.Summary {
	summary := model.Summary{
		Total:         len(results),
		StatusClasses: newStatusClassMap(),
	}

	if len(results) == 0 {
		return summary
	}

	var totalLatency int64

	for i, result := range results {
		totalLatency += result.ResponseTimeMS

		if result.Success {
			summary.Successes++
		} else {
			summary.Failures++
		}

		if isTimeout(result) {
			summary.TimeoutCount++
		}

		summary.StatusClasses[statusClass(result.StatusCode)]++

		if i == 0 || result.ResponseTimeMS < summary.FastestResponseMS {
			summary.FastestURL = result.URL
			summary.FastestResponseMS = result.ResponseTimeMS
		}

		if i == 0 || result.ResponseTimeMS > summary.SlowestResponseMS {
			summary.SlowestURL = result.URL
			summary.SlowestResponseMS = result.ResponseTimeMS
		}
	}

	summary.AverageLatencyMS = float64(totalLatency) / float64(len(results))
	return summary
}

func ComputeSystemStats(jobs []model.CheckJob) model.StatsResponse {
	out := model.StatsResponse{
		TotalJobs:     len(jobs),
		StatusClasses: newStatusClassMap(),
	}

	var totalLatency int64
	var hasTiming bool

	for _, job := range jobs {
		switch job.Status {
		case model.StatusPending:
			out.PendingJobs++
		case model.StatusRunning:
			out.RunningJobs++
		case model.StatusCompleted:
			out.CompletedJobs++
		case model.StatusFailed:
			out.FailedJobs++
		}

		for _, result := range job.Results {
			out.TotalURLsChecked++
			totalLatency += result.ResponseTimeMS

			if result.Success {
				out.SuccessfulChecks++
			} else {
				out.FailedChecks++
			}

			if isTimeout(result) {
				out.TimeoutCount++
			}

			out.StatusClasses[statusClass(result.StatusCode)]++

			if !hasTiming || result.ResponseTimeMS < out.FastestResponseMS {
				out.FastestURL = result.URL
				out.FastestResponseMS = result.ResponseTimeMS
			}

			if !hasTiming || result.ResponseTimeMS > out.SlowestResponseMS {
				out.SlowestURL = result.URL
				out.SlowestResponseMS = result.ResponseTimeMS
			}

			hasTiming = true
		}
	}

	if out.TotalURLsChecked > 0 {
		out.AverageLatencyMS = float64(totalLatency) / float64(out.TotalURLsChecked)
		out.SuccessRatePercent = (float64(out.SuccessfulChecks) / float64(out.TotalURLsChecked)) * 100
	}

	return out
}

func newStatusClassMap() map[string]int {
	return map[string]int{
		"2xx":   0,
		"3xx":   0,
		"4xx":   0,
		"5xx":   0,
		"other": 0,
	}
}

func statusClass(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500 && code < 600:
		return "5xx"
	default:
		return "other"
	}
}

func isTimeout(result model.CheckResult) bool {
	return strings.Contains(strings.ToLower(result.Error), "timed out")
}