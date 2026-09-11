package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type serverMetrics struct {
	httpRequestsTotal atomic.Uint64
}

func newMetricsHandler(
	next http.Handler,
	ready *atomic.Bool,
	result ScanResult,
	metrics *serverMetrics,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		readyValue := 0
		if ready.Load() {
			readyValue = 1
		}

		requestsTotal := metrics.httpRequestsTotal.Load()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		fmt.Fprintf(w, `# HELP godrop_http_requests_total Total number of HTTP requests handled.
# TYPE godrop_http_requests_total counter
godrop_http_requests_total %d
# HELP godrop_ready Whether the service is ready.
# TYPE godrop_ready gauge
godrop_ready %d
# HELP godrop_index_files Number of indexed files.
# TYPE godrop_index_files gauge
godrop_index_files %d
# HELP godrop_index_bytes Total bytes in the index.
# TYPE godrop_index_bytes gauge
godrop_index_bytes %d
`, requestsTotal, readyValue, len(result.Files), result.TotalSize)
	})

	mux.Handle("/", next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.httpRequestsTotal.Add(1)
		mux.ServeHTTP(w, r)
	})
}
