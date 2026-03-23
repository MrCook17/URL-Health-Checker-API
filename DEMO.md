# Demo Script

## Goal

Demonstrate:
- the API is running
- concurrent health checks work
- timeout handling works
- saved jobs can be retrieved
- aggregate stats are updated
- logs show useful structured information

## Start the server

```powershell
go run ./cmd/api
````

## Demo 1: Health endpoint

```powershell
curl.exe http://localhost:8080/health
```

Expected:

* HTTP 200
* JSON response with `status: ok`

## Demo 2: Successful concurrent check

```powershell
curl.exe -X POST http://localhost:8080/checks ^
  -H "Content-Type: application/json" ^
  -d "{\"urls\":[\"https://example.com\",\"https://golang.org\"],\"timeout_ms\":3000}"
```

What to say:

* one job can check multiple URLs
* checks run concurrently
* the response includes status code, response time, timestamps, and a summary

## Demo 3: Timeout handling

```powershell
curl.exe -X POST http://localhost:8080/checks ^
  -H "Content-Type: application/json" ^
  -d "{\"urls\":[\"https://httpstat.us/200?sleep=5000\",\"https://example.com\"],\"timeout_ms\":1000}"
```

What to say:

* each outbound request uses a timeout
* one slow URL times out without crashing the whole job
* the fast URL still completes successfully

## Demo 4: Retrieve a saved job

```powershell
curl.exe http://localhost:8080/checks/chk_002
```

What to say:

* jobs are stored in memory
* saved results can be retrieved later by ID

## Demo 5: Aggregate stats

```powershell
curl.exe http://localhost:8080/stats
```

What to say:

* the API aggregates totals, latency, success rate, timeouts, and status classes
* this helps show the outcome of concurrent processing across jobs

## Demo 6: Show logs

Point to the terminal running the server.

What to show:

* request log with request ID, path, status, and duration
* job created log
* check started/completed/failed logs
* job finished log with successes, failures, timeout count, and duration

## Demo close

Summarise:

* built in Go using `net/http`, contexts, goroutines, channels, mutexes, and tests
* concurrency improved performance significantly in benchmarks
* race detector reported no data races
* current limitations are in-memory storage and no worker-pool limit yet
