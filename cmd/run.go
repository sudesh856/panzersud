package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var url  string
var vus int
var duration string


var runCmd = &cobra.Command{
	Use: "run",
	Short: "Run a load test",
	Run:func(cmd *cobra.Command, args []string) {
		fmt.Printf("Hitting %s with %d VUs for %s\n", url, vus, duration)

	},
}

func init() {
	runCmd.Flags().StringVar(&url,  	 "url", 		"", 				"Target URL")
	runCmd.Flags().IntVar(&vus,          "vus", 10, 						"Virtual users")
	runCmd.Flags().StringVar(&duration,   "duration", 	"30s", 				"Test Duration")

	rootCmd.AddCommand(runCmd)
}

