package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
    Use:   "ui",
    Short: "Show/Open UI link for the deployment",
    Run: func(cmd *cobra.Command, args []string) {
        log.Println("Get the ui endpoint...")

        // _, ezlabFiles, uaConfig := internal.GetDeployConfig()

        // uiHosts := internal.GetCommandOutput(fmt.Sprintf("kubectl --kubeconfig=%s get pod -n istio-system -l app=istio-ingressgateway -o jsonpath='{.items[*].spec.nodeName}'", templateFiles.WorkloadKubeConfig))
        // log.Printf("Update DNS to point %s to %s", uaConfig.Domain, uiHosts)
        // exitCode, err := internal.RunCommand(fmt.Sprintf("kubectl --kubeconfig=%s get pod -n istio-system -l app=istio-ingressgateway -o jsonpath='{.items[*].spec.nodeName}'", ezlabFiles.WorkloadKubeConfig))
        // if err != nil {
        //     log.Printf("Error: %v\n", err)
        // } else {
        //     if exitCode!= 0 {
        //         log.Printf("Error: %v\n", err)
        //     }
        //     log.Println("Done")
        // }
        // kubectl -n istio-system get pod -l app=istio-ingressgateway -o jsonpath='{.items[*].status.hostIP}'
        // OR
        // nodePort=$(kubectl get service -n ezfabric-ui ezfabric-ui -o jsonpath='{.spec.ports[0].nodePort}')
        // echo http://$(hostname -i):${nodePort}

	},
}

func init() {
    unifiedAnalyticsCmd.AddCommand(uiCmd)

}