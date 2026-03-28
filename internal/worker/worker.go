package worker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sudesh856/suddpanzer/internal/scripting"
)

type Job struct {
	Name           string
	URL            string
	Method         string
	Body           string
	Headers        map[string]string
	ExpectedStatus int
	Timeout        time.Duration
	BasicAuth      string

	// ScriptPool is set when this endpoint uses a JS script.
	// Pool is shared across all workers; each worker calls Clone() once at startup
	// to get its own Engine (goja is NOT goroutine-safe).
	ScriptPool *scripting.ScriptPool
}

type Result struct {
	Latency      time.Duration
	StatusCode   int
	Err          error
	Bytes        int64
	Body         []byte
	EndpointName string
}

// one shared transport for all workers
var sharedTransport = &http.Transport{
	ForceAttemptHTTP2:     true,
	DisableKeepAlives:     false,
	MaxIdleConns:          1000,
	MaxIdleConnsPerHost:   1000,
	MaxConnsPerHost:       0,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func RunWorker(ctx context.Context, jobs <-chan Job, results chan<- Result) {
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: sharedTransport,
	}

	// Each worker keeps ONE engine per ScriptPool it encounters.
	// Key = pointer to ScriptPool, value = compiled Engine for this goroutine.
	engines := make(map[*scripting.ScriptPool]*scripting.Engine)

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			if job.Timeout > 0 {
				client.Timeout = job.Timeout
			}

			// ── JS scripting override ──────────────────────────────────────────
			if job.ScriptPool != nil {
				eng, exists := engines[job.ScriptPool]
				if !exists {
					// first time this worker sees this ScriptPool → clone an engine
					var err error
					eng, err = job.ScriptPool.Clone()
					if err != nil {
						results <- Result{
							Err:          fmt.Errorf("script engine init failed: %w", err),
							EndpointName: job.Name,
						}
						continue
					}
					engines[job.ScriptPool] = eng
				}

				override, err := eng.Call()
				if err != nil {
					results <- Result{
						Err:          fmt.Errorf("script call failed: %w", err),
						EndpointName: job.Name,
					}
					continue
				}

				// apply override — script wins over static YAML values
				if override.Method != "" {
					job.Method = override.Method
				}
				if override.Body != "" {
					job.Body = override.Body
				}
				for k, v := range override.Headers {
					if job.Headers == nil {
						job.Headers = make(map[string]string)
					}
					job.Headers[k] = v
				}
			}
			// ──────────────────────────────────────────────────────────────────

			start := time.Now()

			method := job.Method
			if method == "" {
				method = "GET"
			}

			var bodyReader io.Reader
			if job.Body != "" {
				bodyReader = strings.NewReader(job.Body)
			}

			req, err := http.NewRequestWithContext(ctx, method, job.URL, bodyReader)
			if err != nil {
				results <- Result{Latency: time.Since(start), Err: err, EndpointName: job.Name}
				continue
			}

			for k, v := range job.Headers {
				req.Header.Set(k, v)
			}

			if job.BasicAuth != "" {
				parts := strings.SplitN(job.BasicAuth, ":", 2)
				if len(parts) == 2 {
					req.SetBasicAuth(parts[0], parts[1])
				}
			}

			resp, err := client.Do(req)
			latency := time.Since(start)

			if err != nil {
				results <- Result{Latency: latency, Err: err, EndpointName: job.Name}
				continue
			}

			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			var resultErr error
			if job.ExpectedStatus != 0 && resp.StatusCode != job.ExpectedStatus {
				resultErr = fmt.Errorf("expected status %d got %d", job.ExpectedStatus, resp.StatusCode)
			}

			results <- Result{
				Latency:      latency,
				StatusCode:   resp.StatusCode,
				Bytes:        int64(len(bodyBytes)),
				Body:         bodyBytes,
				EndpointName: job.Name,
				Err:          resultErr,
			}
		}
	}
}