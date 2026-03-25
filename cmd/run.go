package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/sudesh856/LoadForge/internal/metrics"
	"github.com/sudesh856/LoadForge/internal/pool"
	"github.com/sudesh856/LoadForge/internal/worker"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"

)

var url  string
var vus int
var duration string
var rps int

var runCmd = &cobra.Command{
	Use: "run",
	Short: "Run a load test",
	Run:func(cmd *cobra.Command, args []string) {
		dur, err := time.ParseDuration(duration)
		if err != nil {
			fmt.Println("Invalid duration:", err)
			return
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		ctx, cancel = context.WithTimeout(context.Background(), dur)
		defer cancel()

		//rate limiter if --rps is 0, unlimited
		var limiter *rate.Limiter
		if rps > 0 {
			limiter = rate.NewLimiter(rate.Limit(rps), rps)
		}

		p := pool.New(100)
		p.Start(ctx, vus)

		agg := metrics.New()
		agg.Start(p.Results())

		//submitting jobs until context expires

		go func() {
			for {
				select{
				case <-ctx.Done():
					return
				default:
					if limiter != nil {
						limiter.Wait(ctx)
					}
					p.Submit(worker.Job{URL: url})
				}
			}
		}()


		//live terminal output
		start := time.Now()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()	

		for {
			select {
			case <-ctx.Done():

				elapsed := time.Since(start)
				fmt.Printf("\n\nDone!\n")
				fmt.Printf("Total Requests : %d\n", agg.TotalRequests())
				fmt.Printf("Duration       : %s\n", elapsed.Round(time.Second))
				fmt.Printf("Avg RPS        : %.2f\n", agg.RPS(elapsed))
				fmt.Printf("p50            : %dms\n", agg.P50())
				fmt.Printf("p99            : %dms\n", agg.P99())
				fmt.Printf("p999           : %dms\n", agg.P999())
				fmt.Printf("Errors         : %d\n", agg.ErrorCount())
				fmt.Printf("Error Rate     : %.2f%%\n", agg.ErrorRate())
				return

			case <-ticker.C:
				elapsed := time.Since(start)
				fmt.Printf("\rRequests: %d | RPS: %.0f | p99: %dms | Errors: %d",
				agg.TotalRequests(),
				agg.RPS(elapsed),
				agg.P99(),
				agg.ErrorCount(),
			
			)
			}
		}
	},
}

func init() {
	runCmd.Flags().StringVar(&url,  	 "url", 		"", 				"Target URL")
	runCmd.Flags().IntVar(&vus,          "vus", 10, 						"Virtual users")
	runCmd.Flags().StringVar(&duration,   "duration", 	"30s", 				"Test Duration")
	runCmd.Flags().IntVar(&rps,         "rps",      0,     "Max requests per second (0 = unlimited)")


	rootCmd.AddCommand(runCmd)
}

