// Package prometheus exposes a /metrics endpoint in Prometheus text format
// during a suddpanzer run. Scrape it with Grafana/Prometheus while the test runs.
//
// Start with --metrics-addr :9090
// Then scrape: http://localhost:9090/metrics
//
// Metrics exposed:
//   suddpanzer_rps_current
//   suddpanzer_latency_p99_ms
//   suddpanzer_latency_p95_ms
//   suddpanzer_latency_p50_ms
//   suddpanzer_error_total
//   suddpanzer_requests_total
//   suddpanzer_vus_active
package prometheus

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// Snapshot holds live metrics updated by the run loop.
// All fields use atomic ops — no locks, no contention with workers.
type Snapshot struct {
	RPSCurrent    atomic.Int64 // stored as rps*100 for fixed-point precision
	LatencyP99Ms  atomic.Int64
	LatencyP95Ms  atomic.Int64
	LatencyP50Ms  atomic.Int64
	ErrorTotal    atomic.Int64
	RequestsTotal atomic.Int64
	VUsActive     atomic.Int64
}

// Server is the Prometheus metrics HTTP server.
type Server struct {
	snap *Snapshot
	srv  *http.Server
}

// New creates a metrics server. Call Start() to begin listening.
func New(addr string) *Server {
	snap := &Snapshot{}
	s := &Server{snap: snap}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	s.srv = &http.Server{Addr: addr, Handler: mux}
	return s
}

// Start begins serving in a background goroutine.
func (s *Server) Start() {
	go func() { _ = s.srv.ListenAndServe() }()
}

// Stop shuts the server down.
func (s *Server) Stop() { _ = s.srv.Close() }

// Snap returns the Snapshot so the run loop can update values.
func (s *Server) Snap() *Snapshot { return s.snap }

// SetRPS stores rps as a fixed-point int64 (multiply by 100).
func (snap *Snapshot) SetRPS(rps float64) {
	snap.RPSCurrent.Store(int64(rps * 100))
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	rps := float64(s.snap.RPSCurrent.Load()) / 100.0

	fmt.Fprintf(w, "# HELP suddpanzer_rps_current Current requests per second\n")
	fmt.Fprintf(w, "# TYPE suddpanzer_rps_current gauge\n")
	fmt.Fprintf(w, "suddpanzer_rps_current %.2f\n\n", rps)

	fmt.Fprintf(w, "# HELP suddpanzer_latency_p99_ms P99 latency in milliseconds\n")
	fmt.Fprintf(w, "# TYPE suddpanzer_latency_p99_ms gauge\n")
	fmt.Fprintf(w, "suddpanzer_latency_p99_ms %d\n\n", s.snap.LatencyP99Ms.Load())

	fmt.Fprintf(w, "# HELP suddpanzer_latency_p95_ms P95 latency in milliseconds\n")
	fmt.Fprintf(w, "# TYPE suddpanzer_latency_p95_ms gauge\n")
	fmt.Fprintf(w, "suddpanzer_latency_p95_ms %d\n\n", s.snap.LatencyP95Ms.Load())

	fmt.Fprintf(w, "# HELP suddpanzer_latency_p50_ms P50 latency in milliseconds\n")
	fmt.Fprintf(w, "# TYPE suddpanzer_latency_p50_ms gauge\n")
	fmt.Fprintf(w, "suddpanzer_latency_p50_ms %d\n\n", s.snap.LatencyP50Ms.Load())

	fmt.Fprintf(w, "# HELP suddpanzer_error_total Total errors\n")
	fmt.Fprintf(w, "# TYPE suddpanzer_error_total counter\n")
	fmt.Fprintf(w, "suddpanzer_error_total %d\n\n", s.snap.ErrorTotal.Load())

	fmt.Fprintf(w, "# HELP suddpanzer_requests_total Total requests fired\n")
	fmt.Fprintf(w, "# TYPE suddpanzer_requests_total counter\n")
	fmt.Fprintf(w, "suddpanzer_requests_total %d\n\n", s.snap.RequestsTotal.Load())

	fmt.Fprintf(w, "# HELP suddpanzer_vus_active Active virtual users\n")
	fmt.Fprintf(w, "# TYPE suddpanzer_vus_active gauge\n")
	fmt.Fprintf(w, "suddpanzer_vus_active %d\n", s.snap.VUsActive.Load())
}