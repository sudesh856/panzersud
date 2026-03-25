package reporter
import "fmt"

func Print(requests int64, rps float64, p99 int64, errors int64) {
	fmt.Printf("\rRequests: %d | RPS: %.0f | p99: %dms | Errors: %d",
		requests, rps, p99, errors)
}