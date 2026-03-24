package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/sudesh856/LoadForge/internal/pool"
	"github.com/sudesh856/LoadForge/internal/worker"
)

var url  string
var vus int
var duration string


var runCmd = &cobra.Command{
	Use: "run",
	Short: "Run a load test",
	Run:func(cmd *cobra.Command, args []string) {
		dur, err := time.ParseDuration(duration)
		if err != nil {
			fmt.Println("Invalid duration:", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), dur)
		defer cancel()
		p := pool.New(100)
		p.Start(ctx, vus)

		//submitting jobs until context expires

		go func() {
			for {
				select{
				case <-ctx.Done():
					return
				default:
					p.Submit(worker.Job{URL: url})
				}
			}
		}()

		//collecting results until context expires

		total := 0
		errors := 0

		for {
			select {
			case <-ctx.Done():
				fmt.Printf("\nDone! Total requests: %d Errors: %d\n", total, errors)
				return
			case result := <-p.Results():
				total++
				if result.Err != nil {
					errors++
				}

				fmt.Printf("\rRequests: %d | Errors: %d", total, errors)
			}
		}
	},
}

func init() {
	runCmd.Flags().StringVar(&url,  	 "url", 		"", 				"Target URL")
	runCmd.Flags().IntVar(&vus,          "vus", 10, 						"Virtual users")
	runCmd.Flags().StringVar(&duration,   "duration", 	"30s", 				"Test Duration")

	rootCmd.AddCommand(runCmd)
}

