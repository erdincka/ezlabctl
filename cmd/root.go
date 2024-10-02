package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "ezlabctl",
    Short: "HPE UA deployment tool",
    Long:  `Ezlab is a deployment tool for setting up Ezmeral Unified Analytics on multiple hosts.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
    if err := rootCmd.Execute(); err != nil {
        log.Fatalln(err)
    }
}

func init() {
    // Flags and configuration settings
    // Subcommands
}