package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
    Use:   "install",
    Short: "Configure and set up Data Fabric for UA deployment",
    Run: func(cmd *cobra.Command, args []string) {
        log.Println("Setting up Data Fabric...")
	},
}
