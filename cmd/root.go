package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "ezlab",
    Short: "HPE UA deployment tool",
    Long:  `Ezlab is a deployment tool for setting up Ezmeral Unified Analytics on multiple hosts.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func init() {
    // Here we will add subcommands like prepare, deploy, etc.
    rootCmd.AddCommand(prepareCmd)
    rootCmd.AddCommand(setupDfCmd)
    rootCmd.AddCommand(installCmd)
}