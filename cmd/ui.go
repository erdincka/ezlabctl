package cmd

import (
	"ezlabctl/internal"
	"log"

	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
    Use:   "ui",
    Short: "Show/Open UI link for the deployment",
    Run: func(cmd *cobra.Command, args []string) {
        log.Println("Get the ui endpoint...")

        // appConfig := internal.GetAppConfiguration()
        orchKubeconfig := "/tmp/ezlab-ua15.kayalab.uk/mgmt-kubeconfig"
        exitCode, err := internal.RunCommand("kubectl get service -n ezfabric-ui ezfabric-ui -o jsonpath='{.spec.ports[0].nodePort}' --kubeconfig=" + orchKubeconfig)
        if err != nil {
            log.Printf("Error: %v\n", err)
        } else {
            if exitCode!= 0 {
                log.Printf("Error: %v\n", err)
            }
            log.Println("Done")
        }
        // kubectl -n istio-system get pod -l app=istio-ingressgateway -o jsonpath='{.items[*].status.hostIP}'
        // OR
        // nodePort=$(kubectl get service -n ezfabric-ui ezfabric-ui -o jsonpath='{.spec.ports[0].nodePort}')
        // echo http://$(hostname -i):${nodePort}

	},
}