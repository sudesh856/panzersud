# suddpanzer

A fast, open-source HTTP load testing tool written in Go.
Point it at any endpoint, set your virtual users and duration — it hammers it and tells you exactly how your server held up.

> A self-hosted, developer-friendly alternative to k6 — single binary, no cloud account needed.

---

## Install
```bash
git clone https://github.com/sudesh856/suddpanzer.git
cd suddpanzer
go build -o sudd .
```

## Quickstart
```bash
# Basic load test
sudd run --url https://api.example.com --vus 100 --duration 30s

# With RPS cap
sudd run --url https://api.example.com --vus 100 --duration 30s --rps 50

# JSON output (great for CI)
sudd run --url https://api.example.com --vus 100 --duration 30s --output json
```

## Example Output
```
Requests: 1658 | RPS: 28 | p99: 994ms | Errors: 0

===== SUDD LOAD TEST SUMMARY =====
URL            : https://api.example.com
VUs            : 10
Duration       : 30s
-----------------------------------
Total Requests : 1658
Avg RPS        : 27.55
-----------------------------------
p50            : 295ms
p75            : 363ms
p90            : 523ms
p95            : 646ms
p99            : 994ms
p999           : 1585ms
Max            : 1586ms
-----------------------------------
Errors         : 0
Error Rate     : 0.00%
===================================
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--url` | required | Target URL to load test |
| `--vus` | 10 | Number of virtual users (goroutines) |
| `--duration` | 30s | How long to run the test |
| `--rps` | 0 (unlimited) | Max requests per second |
| `--output` | text | Output format: `text` or `json` |

## Features

- Goroutine-based worker pool — up to 100,000+ concurrent virtual users
- HDR histogram latency tracking — p50, p75, p90, p95, p99, p999, max
- Token bucket rate limiter — cap RPS with `--rps`
- Live terminal output — updates every second
- JSON output mode — pipe into scripts or CI
- Graceful shutdown — Ctrl+C prints full summary
- Single binary — no runtime, no dependencies

## Roadmap

- [ ] YAML scenario config
- [ ] Web dashboard with live charts
- [ ] CI mode with pass/fail thresholds
- [ ] Distributed mode across multiple machines
- [ ] gRPC and WebSocket support

## License

MIT