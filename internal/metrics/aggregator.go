package metrics
import (
	"sync"
	"time"
	hdrhistogram "github.com/HdrHistogram/hdrhistogram-go"
	"github.com/sudesh856/suddpanzer/internal/worker"
)

type Aggregator struct {
	mu 			sync.Mutex
	totalRequests int64
	errorCount int64
	bytesTotal int64
	histogram *hdrhistogram.Histogram
}

func New() *Aggregator {
	return &Aggregator{
		//tracking latency from 1ms to 1 minute, 3 significant figures

		histogram: hdrhistogram.New(1, 60000, 3),
	}
}

//starting reads from results channel in a goroutine
func (a *Aggregator) Start(results <-chan worker.Result) {
	go func() {
		for result := range results {
			a.mu.Lock()
			a.totalRequests++
			if result.Err != nil {
				a.errorCount++
			}

			if result.Bytes > 0 {
				a.bytesTotal += result.Bytes
			}

			//recording latency in milliseconds
			ms := result.Latency.Milliseconds()
			if ms < 1{
				ms = 1 //minimum 1ms
			}
			a.histogram.RecordValue(ms)
			a.mu.Unlock()
		}
	}()
}

func (a *Aggregator) P50() int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.histogram.ValueAtQuantile(50)
}

func (a *Aggregator) P99() int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.histogram.ValueAtQuantile(99)
}

func (a *Aggregator) P999() int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.histogram.ValueAtQuantile(99.9)
}

func (a *Aggregator) RPS(elapsed time.Duration) float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	if elapsed.Seconds() == 0 {
		return 0
	}
	return float64(a.totalRequests) / elapsed.Seconds()
}


func (a *Aggregator) P75() int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.histogram.ValueAtQuantile(75)
}

func (a *Aggregator) P90() int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.histogram.ValueAtQuantile(90)
}

func (a *Aggregator) P95() int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.histogram.ValueAtQuantile(95)
}

func (a *Aggregator) Max() int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.histogram.Max()
}

func (a *Aggregator) ErrorRate() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.totalRequests == 0 {
		return 0
	}
	return float64(a.errorCount) / float64(a.totalRequests) * 100
}


func (a *Aggregator) TotalRequests() int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.totalRequests
}

func (a *Aggregator) ErrorCount() int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.errorCount
}