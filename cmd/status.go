package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "Check the status for the deployment",
    Run: func(cmd *cobra.Command, args []string) {

        log.Println("Check cluster status...")
	},
}