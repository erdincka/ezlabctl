package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var kubeconfCmd = &cobra.Command{
    Use:   "kubeconf",
    Short: "Get the kubeconfig file for the deployment",
    Run: func(cmd *cobra.Command, args []string) {

        log.Println("Get the kubeconfig file...")
	},
}