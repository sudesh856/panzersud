package thresholds

import "fmt"


type Thresholds struct {
	P99Ms        float64 `yaml:"p99_ms"`
	P95Ms        float64 `yaml:"p95_ms"`
	ErrorRatePct float64 `yaml:"error_rate_pct"`
	MinRPS       float64 `yaml:"min_rps"`
}

// Failure represents a single threshold that was breached.
type Failure struct {
	Metric   string
	Actual   float64
	Limit    float64
	Exceeded bool // true = actual > limit, false = actual < limit (for min_rps)
}

// String returns a human-readable failure message printed to the terminal.
func (f Failure) String() string {
	if f.Exceeded {
		return fmt.Sprintf("THRESHOLD FAILED: %s=%.2f > %.2f", f.Metric, f.Actual, f.Limit)
	}
	return fmt.Sprintf("THRESHOLD FAILED: %s=%.2f < %.2f (minimum)", f.Metric, f.Actual, f.Limit)
}

// Results holds the actual measured values to compare against thresholds.
type Results struct {
	P99Ms        float64
	P95Ms        float64
	ErrorRatePct float64
	AvgRPS       float64
}

// Evaluate compares actual Results against declared Thresholds.
// Returns a slice of Failures (empty = all passed).
func Evaluate(t Thresholds, r Results) []Failure {
	var failures []Failure

	if t.P99Ms > 0 && r.P99Ms > t.P99Ms {
		failures = append(failures, Failure{
			Metric:   "p99",
			Actual:   r.P99Ms,
			Limit:    t.P99Ms,
			Exceeded: true,
		})
	}

	if t.P95Ms > 0 && r.P95Ms > t.P95Ms {
		failures = append(failures, Failure{
			Metric:   "p95",
			Actual:   r.P95Ms,
			Limit:    t.P95Ms,
			Exceeded: true,
		})
	}

	if t.ErrorRatePct > 0 && r.ErrorRatePct > t.ErrorRatePct {
		failures = append(failures, Failure{
			Metric:   "error_rate_pct",
			Actual:   r.ErrorRatePct,
			Limit:    t.ErrorRatePct,
			Exceeded: true,
		})
	}

	if t.MinRPS > 0 && r.AvgRPS < t.MinRPS {
		failures = append(failures, Failure{
			Metric:   "min_rps",
			Actual:   r.AvgRPS,
			Limit:    t.MinRPS,
			Exceeded: false,
		})
	}

	return failures
}