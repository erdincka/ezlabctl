package cmd

import (
	"encoding/base64"
	"encoding/json"
	"ezlabctl/internal"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var kubeconf, clusterName string

var kubeconfCmd = &cobra.Command{
    Use:   "kubeconf <orch-kubeconfig> <clustername>",
    Args: cobra.MinimumNArgs(2),
    Short: "Get the kubeconfig file for the deployment",
    Run: func(cmd *cobra.Command, args []string) {
        kubeconf = args[0]
        clusterName = args[1]

        workloadKubeconfig := fmt.Sprintf("/tmp/ez-%s/wkld-kubeconfig", clusterName)

        command := fmt.Sprintf("kubectl --kubeconfig=%s get secret %s-kubeconfig -n %s -o json", kubeconf, clusterName, clusterName)
        log.Printf("Get workload kubeconfig with: %s\n", command)
        out := internal.GetCommandOutput(command)

        var outJson map[string]interface{}
	    err := json.Unmarshal([]byte(out), &outJson)
        if err!= nil {
            log.Fatal(err)
        }

        data, ok := outJson["data"].(map[string]interface{})
        if !ok {
            log.Fatal("Could not get the data from the json")
        }

        value, ok := data["value"].(string)
        if!ok {
            log.Fatal("Could not get the value from the json")
        }
        kubeConf, err := base64.StdEncoding.DecodeString(value)
        if err != nil {
            fmt.Println("Error decoding base64 string:", err)
            return
        }

        // Backup existing file
        err = os.Rename(workloadKubeconfig, workloadKubeconfig + ".bak")
        if err != nil {
            log.Println("Could not backup existing file, ignoring...")
        }

        // Create a new file and write the decoded content to it
        err = os.WriteFile(workloadKubeconfig, kubeConf, 0644)
        if err != nil {
            fmt.Println("Error writing to file:", err)
            return
        }

        fmt.Printf("Kube configuration saved to %s\n", workloadKubeconfig)

	},
}

func init() {
    unifiedAnalyticsCmd.AddCommand(kubeconfCmd)
}