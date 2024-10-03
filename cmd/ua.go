package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var unifiedAnalyticsCmd = &cobra.Command{
    Use:   "ua",
	Short: "Install Unified Analytics, run prechecks if -p flag is set, install orchestrator if -o flag is set etc",
	Args: cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {

		host := args[0]
		log.Println("Installing Unified Analytics on host: ", host)
	},
}

func init() {
	rootCmd.AddCommand(unifiedAnalyticsCmd)
}