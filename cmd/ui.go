package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
    Use:   "ui",
    Short: "Show/Open UI link for the deployment",
    Run: func(cmd *cobra.Command, args []string) {

        log.Println("Get the ui endpoint...")
	},
}