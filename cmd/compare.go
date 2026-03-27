package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/sudesh856/suddpanzer/internal/store"
)

type runSummary struct {
	AvgRPS        float64 `json:"avg_rps"`
	P50           int64   `json:"p50_ms"`
	P75           int64   `json:"p75_ms"`
	P90           int64   `json:"p90_ms"`
	P95           int64   `json:"p95_ms"`
	P99           int64   `json:"p99_ms"`
	P999          int64   `json:"p999_ms"`
	Max           int64   `json:"max_ms"`
	Errors        int64   `json:"errors"`
	ErrorRate     float64 `json:"error_rate_pct"`
	TotalRequests int64   `json:"total_requests"`
	DurationSecs  float64 `json:"duration_secs"`
}

var compareCmd = &cobra.Command{
	Use:   "compare <run-id-1> <run-id-2>",
	Short: "Compare two past run results",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		id1, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("invalid run id:", args[0])
			return
		}
		id2, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			fmt.Println("invalid run id:", args[1])
			return
		}

		st, err := store.New("./blast.db")
		if err != nil {
			fmt.Println("store error:", err)
			return
		}
		defer st.Close()

		r1, err := st.GetRun(id1)
		if err != nil {
			fmt.Printf("run %d not found\n", id1)
			return
		}
		r2, err := st.GetRun(id2)
		if err != nil {
			fmt.Printf("run %d not found\n", id2)
			return
		}

		var s1, s2 runSummary
		if err := json.Unmarshal([]byte(r1.Summary), &s1); err != nil {
			fmt.Println("failed to parse run 1 summary:", err)
			return
		}
		if err := json.Unmarshal([]byte(r2.Summary), &s2); err != nil {
			fmt.Println("failed to parse run 2 summary:", err)
			return
		}

		fmt.Printf("\n========== BLAST RUN COMPARISON ==========\n")
		fmt.Printf("%-22s %-18s %-18s %s\n", "METRIC", fmt.Sprintf("Run #%d (%s)", id1, r1.Name), fmt.Sprintf("Run #%d (%s)", id2, r2.Name), "DELTA")
		fmt.Printf("------------------------------------------------------------------\n")

		printRow("Total Requests", s1.TotalRequests, s2.TotalRequests, "")
		printRowF("Avg RPS", s1.AvgRPS, s2.AvgRPS, "req/s")
		printRowF("Duration", s1.DurationSecs, s2.DurationSecs, "s")
		fmt.Printf("------------------------------------------------------------------\n")
		printRowMs("p50", s1.P50, s2.P50)
		printRowMs("p75", s1.P75, s2.P75)
		printRowMs("p90", s1.P90, s2.P90)
		printRowMs("p95", s1.P95, s2.P95)
		printRowMs("p99", s1.P99, s2.P99)
		printRowMs("p999", s1.P999, s2.P999)
		printRowMs("Max", s1.Max, s2.Max)
		fmt.Printf("------------------------------------------------------------------\n")
		printRow("Errors", s1.Errors, s2.Errors, "")
		printRowF("Error Rate", s1.ErrorRate, s2.ErrorRate, "%")
		fmt.Printf("==========================================\n\n")
	},
}

func printRowMs(label string, a, b int64) {
	delta := b - a
	sign := "+"
	if delta < 0 {
		sign = ""
	}
	indicator := "✅"
	if delta > 0 {
		indicator = "🔺"
	}
	fmt.Printf("%-22s %-18s %-18s %s%dms %s\n",
		label,
		fmt.Sprintf("%dms", a),
		fmt.Sprintf("%dms", b),
		sign, delta, indicator,
	)
}

func printRow(label string, a, b int64, unit string) {
	delta := b - a
	sign := "+"
	if delta < 0 {
		sign = ""
	}
	fmt.Printf("%-22s %-18s %-18s %s%d%s\n",
		label,
		fmt.Sprintf("%d%s", a, unit),
		fmt.Sprintf("%d%s", b, unit),
		sign, delta, unit,
	)
}

func printRowF(label string, a, b float64, unit string) {
	delta := b - a
	sign := "+"
	if delta < 0 {
		sign = ""
	}
	fmt.Printf("%-22s %-18s %-18s %s%.2f%s\n",
		label,
		fmt.Sprintf("%.2f%s", a, unit),
		fmt.Sprintf("%.2f%s", b, unit),
		sign, delta, unit,
	)
}

func init() {
	rootCmd.AddCommand(compareCmd)
}