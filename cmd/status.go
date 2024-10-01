package cmd

import (
	"ezlabctl/internal"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
    Use:   "status [-w]",
    Short: "Check the status for the deployment, -w for waiting",
    Run: func(cmd *cobra.Command, args []string) {
        appConfig := internal.GetAppConfiguration()
        clusterName := strings.Split(appConfig.Domain, ".")[0]
        ezlabFilesDir := "/tmp/ez-" + clusterName
        orchestratorKubeConfig := ezlabFilesDir + "/mgmt-kubeconfig"

        command := fmt.Sprintf("kubectl --kubeconfig=%s describe ezfabriccluster/%s -n %s", orchestratorKubeConfig, clusterName, clusterName)

        // Check if the watch flag is provided
        if cmd.Flags().Changed("watch") {
            command = fmt.Sprintf("kubectl --kubeconfig=%s logs -n %s -f --pod-running-timeout=600s --since=1m w-op-workload-deploy-%s", orchestratorKubeConfig, clusterName, clusterName)
        }

        log.Printf("Check cluster status with: %s\n", command)
        exitCode, err := internal.RunCommand(command)
        if err != nil {
            log.Fatalf("Error: %v\n", err)
        } else {
            log.Printf("Finished: %d\n", exitCode)
        }
	},
}
