package cmd

import (
	"ezlabctl/internal"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var kubeconfCmd = &cobra.Command{
    Use:   "kubeconf",
    Short: "Get the kubeconfig file for the deployment",
    Run: func(cmd *cobra.Command, args []string) {
        log.Println("Get the kubeconfig file...")
        appConfig := internal.GetAppConfiguration()
        ezlabFilesDir := "/tmp/ezlab-" + appConfig.Domain
        orchestratorKubeConfig := ezlabFilesDir + "/mgmt-kubeconfig"
        workloadKubeConfig := ezlabFilesDir + "/workload-kubeconfig"

        log.Println("Initializing the orchestrator fabric...")
        exitCode, err := internal.RunCommand("ezfabricctl workload get kubeconfig --clustername " + strings.Split(appConfig.Domain, ".")[0] + " --save-kubeconfig " + workloadKubeConfig + " --kubeconfig " + orchestratorKubeConfig)
        if err != nil {
            log.Fatalf("Error: %v\n", err)
        } else {
            log.Printf("Orchestrator init finished with exit code %d\n", exitCode)
        }
	},
}