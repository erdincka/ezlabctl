package cmd

import (
	"log"

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
        log.Fatalln(err)
    }
}

func init() {
    // Here you will define your flags and configuration settings.
    statusCmd.Flags().BoolP("watch", "w", false, "Watch the cluster status")

    // Here we will add subcommands like prepare, deploy, etc.
    rootCmd.AddCommand(prepareCmd)
    rootCmd.AddCommand(setupCmd)
    rootCmd.AddCommand(deployCmd)
    rootCmd.AddCommand(statusCmd)
    rootCmd.AddCommand(kubeconfCmd)
    rootCmd.AddCommand(uiCmd)
}