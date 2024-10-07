package cmd

import (
	"ezlabctl/internal"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
    Use:   "ui <workload-kubeconfig> <clustername>",
    Short: "UI access for the deployment",
    Args: cobra.MinimumNArgs(2),
    Run: func(cmd *cobra.Command, args []string) {
        kubeconf = args[0]
        clusterName = args[1]

        command :=  fmt.Sprintf("kubectl --kubeconfig=%s get pod -n istio-system -l app=istio-ingressgateway -o jsonpath='{.items[*].spec.nodeName}'", kubeconf)
        log.Printf("Get UI endpoints with: %s\n", command)

        out := internal.GetCommandOutput(command)

        endpoints := strings.Split(out, " ")
        // log.Println(endpoints)

        domain := strings.Join(strings.Split(endpoints[0], ".")[1:], ".")
        log.Println(domain)

        log.Printf("Update DNS to point %s.%s to %s", clusterName, domain, out)

        log.Printf("Open GUI after DNS configuration: https://home.%s.%s", clusterName, domain)
	},
}

func init() {
    unifiedAnalyticsCmd.AddCommand(uiCmd)
}