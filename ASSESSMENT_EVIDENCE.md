# Assessment Evidence

## Test results

```text
go test ./...
ok      healthcheck-api/internal/checker
ok      healthcheck-api/internal/httpapi
ok      healthcheck-api/internal/stats
ok      healthcheck-api/internal/validation
````

## Race detector results

```text
go test -race -count=2 ./...
ok      healthcheck-api/internal/checker
ok      healthcheck-api/internal/httpapi
ok      healthcheck-api/internal/stats
ok      healthcheck-api/internal/validation
```

## Benchmark results

```text
BenchmarkCheckURLsSequential-8   10  103353040 ns/op  28469 B/op  295 allocs/op
BenchmarkCheckURLsConcurrent-8   55   24267751 ns/op  71803 B/op  547 allocs/op
```

## Benchmark interpretation

* Concurrent execution was about 4.3x faster than sequential execution
* The concurrent version used more memory and allocations
* This shows a clear performance benefit from goroutines and concurrent result collection

## Screenshot checklist

Take screenshots of:

1. terminal showing `go test ./...`
2. terminal showing `go test -race -count=2 ./...`
3. terminal showing `go test -bench . -benchmem ./internal/httpapi`
4. terminal showing server logs during a timeout demo
5. successful `POST /checks` JSON response
6. timed-out `POST /checks` JSON response
7. `GET /stats` JSON response

## Poster-ready findings sentence

Testing passed across the API, no data races were found, and benchmarking showed the concurrent checker was approximately 4.3x faster than the sequential version.
