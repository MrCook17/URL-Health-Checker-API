# Go Website Health Check API

A concurrent REST API in Go that checks website availability and performance, executes HTTP checks in parallel, and returns structured JSON results.

## Features

- `GET /health` health endpoint
- `POST /checks` create and run a website check job
- `GET /checks/{id}` retrieve a saved job
- `GET /checks` list saved jobs
- `GET /stats` return aggregate API statistics
- Concurrent URL checking using goroutines
- Per-request timeout handling using `context.WithTimeout`
- Structured JSON error responses
- In-memory job storage
- Structured request and job logging
- Unit tests, handler tests, race detector coverage, and benchmarks

## Project structure

```text
healthcheck-api/
  cmd/api/main.go
  internal/checker/checker.go
  internal/httpapi/handler.go
  internal/httpapi/middleware.go
  internal/model/models.go
  internal/stats/stats.go
  internal/store/memory.go
  internal/validation/validation.go
````

## Run the API

```powershell
go run ./cmd/api
```

Server address:

```text
http://localhost:8080
```

## Example requests

Health check:

```powershell
curl.exe http://localhost:8080/health
```

Create a normal job:

```powershell
curl.exe -X POST http://localhost:8080/checks ^
  -H "Content-Type: application/json" ^
  -d "{\"urls\":[\"https://example.com\",\"https://golang.org\"],\"timeout_ms\":3000}"
```

Create a timeout demo job:

```powershell
curl.exe -X POST http://localhost:8080/checks ^
  -H "Content-Type: application/json" ^
  -d "{\"urls\":[\"https://httpstat.us/200?sleep=5000\",\"https://example.com\"],\"timeout_ms\":1000}"
```

Get one saved job:

```powershell
curl.exe http://localhost:8080/checks/chk_001
```

List all jobs:

```powershell
curl.exe http://localhost:8080/checks
```

Get aggregate stats:

```powershell
curl.exe http://localhost:8080/stats
```

## Design decisions

* Healthy website = HTTP status `200` to `399`
* Checks run immediately when `POST /checks` is called
* Results are stored in memory
* Maximum 20 URLs per job
* Default timeout is 3000 ms
* One goroutine is started per URL in the current implementation

## Testing

Run all tests:

```powershell
go test ./...
```

Run the race detector:

```powershell
go test -race -count=2 ./...
```

Run benchmarks:

```powershell
go test -bench . -benchmem ./internal/httpapi
```

## Benchmark evidence

Current benchmark results:

* Sequential: `103353040 ns/op`
* Concurrent: `24267751 ns/op`

This is about a 4.3x speedup for the concurrent version.

## Limitations

* Storage is in-memory only, so jobs are lost when the server stops
* Concurrency is currently one goroutine per URL rather than a worker pool
* No authentication or rate limiting is implemented
* No persistent configuration file or environment-based configuration yet
