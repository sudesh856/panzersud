package metrics

import (
	"fmt"
	"testing"
	"time"

	"github.com/sudesh856/suddpanzer/internal/worker"
)

func TestAggregator(t *testing.T) {
	results := make(chan worker.Result, 10)
	a := New()
	a.Start(results)

	//send known inputs
	results <- worker.Result{Latency: 10 *time.Millisecond, StatusCode: 200, Err: nil}
	results <-worker.Result{Latency: 20*time.Millisecond, StatusCode: 200, Err: nil}
	results <- worker.Result{Latency: 50 * time.Millisecond, StatusCode: 200, Err: nil}
	results <- worker.Result{Latency: 100 * time.Millisecond, StatusCode: 500, Err: nil}
	results <- worker.Result{Latency: 200 * time.Millisecond, StatusCode: 200, Err: fmt.Errorf("timeout")}

	//giving time to process
	time.Sleep(50 * time.Millisecond)

	if a.TotalRequests() != 5 {
		t.Errorf("Expected 5 requests, got %d", a.ErrorCount())
	}

	t.Logf("P50: %dms", a.P50())
	t.Logf("P99:  %dms", a.P99())
	t.Logf("P999: %dms", a.P999())
	t.Logf("RPS:  %.2f", a.RPS(5*time.Second))
	t.Logf("Error rate: %.2f%%", a.ErrorRate())
}