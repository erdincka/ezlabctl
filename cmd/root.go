package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "os"
)

var rootCmd = &cobra.Command{
    Use:   "ezlab",
    Short: "Ezlab is a deployment tool for multi-node Kubernetes clusters",
    Long:  `Ezlab simplifies system preparation and deployment across multiple servers.`,
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
}