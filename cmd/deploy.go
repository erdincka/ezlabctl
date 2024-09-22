package cmd

import (
	"encoding/base64"
	"encoding/json"
	"ezlabctl/internal"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
    Use:   "deploy",
    Short: "Update yaml files and start the installer",
    Run: func(cmd *cobra.Command, args []string) {
        log.Println("Creating yaml files")

        appConfig := internal.GetAppConfiguration()
        ezlabFilesDir := "/tmp/ezlab-" + appConfig.Domain
        orchestratorKubeConfig := ezlabFilesDir + "/mgmt-kubeconfig"

        // Setup auth data for admin user
        authData := map[string]interface{}{
            "admin_user": map[string]string{
                "fullname": "Ezmeral Admin",
                "email":    fmt.Sprintf("ezadmin@%s", appConfig.Domain),
                "username": "ezua",
                "password": appConfig.DFPass,
            },
        }
        // Convert authData to JSON
        authDataJSON, err := json.Marshal(authData)
        if err != nil {
            fmt.Println("Error converting authData to JSON:", err)
            return
        }
        // log.Println("Auth data: " + string(authDataJSON))

        uaConfig := internal.UADeployConfig{
            Username: appConfig.Username,
            Password: base64.StdEncoding.EncodeToString([]byte(appConfig.Password)),
            Domain: appConfig.Domain,
            RegistryUsername: "",
            RegistryPassword: "",
            RegistryUrl: "",
            RegistryInsecure: "",
            // Orchestrator: internal.GetOutboundIP(),
            Orchestrator: appConfig.Orchestrator.IP,
            Master: appConfig.Controller.IP,
            Workers: strings.Split(internal.GetWorkerIPs(), ","),
            ClusterName: strings.Split(appConfig.Domain, ".")[0],
            AuthData: base64.StdEncoding.EncodeToString(authDataJSON),
            NoProxy: "10.96.0.0/12,10.224.0.0/16,10.43.0.0/16,192.168.0.0/16,.external.hpe.local,localhost,.cluster.local,.svc,.default.svc,127.0.0.1,169.254.169.254," + internal.GetWorkerIPs() + "," + appConfig.Controller.IP + "," + appConfig.Orchestrator.IP + ",." + appConfig.Domain,
        }

        err = os.MkdirAll(ezlabFilesDir,0755)
        if err!= nil {
            log.Fatal("Error creating yaml directory: ",err)
        }

        yamlfiles, err := internal.GetTemplateFiles()
        if err != nil {
            fmt.Println("Error:", err)
            return
        }

        var wg sync.WaitGroup
        for _, file := range yamlfiles {
            wg.Add(1)
            go func(f string) {
                defer wg.Done()
                internal.ProcessTemplate(f, ezlabFilesDir + "/" + filepath.Base(f), uaConfig)
                log.Println("Processing: " + f)
            }(file)
        }
        wg.Wait()

        log.Println("Successfully created yaml files")

        // Run Prechecks
        log.Println("Running prechecks...")
        exitCode, err := internal.RunCommand("ezfabricctl prechecks --input " + ezlabFilesDir + "/prechecks.yaml --parallel=true --cleanup=true")
        if err != nil {
            log.Fatalf("Error: %v\n", err)
        } else {
            log.Printf("Prechecks finished with exit code %d\n", exitCode)
        }
        // TODO: Process non-zero exit code

        // Initialize fabric orchestrator
        log.Println("Initializing the orchestrator fabric...")
        exitCode, err = internal.RunCommand("ezfabricctl orchestrator init --input " + ezlabFilesDir + "/coord-init.yaml --releasepkg /usr/local/share/applications/ezfab-release.tgz --save-kubeconfig " + orchestratorKubeConfig)
        if err != nil {
            log.Fatalf("Error: %v\n", err)
        } else {
            log.Printf("Orchestrator init finished with exit code %d\n", exitCode)
        }
        // TODO: Process non-zero exit code

        // Add hosts for Workload cluster
        log.Println("Adding workload hosts to fabric...")
        exitCode, err = internal.RunCommand("ezfabricctl poolhost init --input " + ezlabFilesDir + "/hosts-init.yaml --kubeconfig " + orchestratorKubeConfig)
        if err != nil {
            log.Printf("Error: %v\n", err)
        } else {
            log.Printf("Hosts init with exit code %d\n", exitCode)
        }
        // TODO: Process non-zero exit code

        // Create workload cluster
        log.Println("Creating workload cluster...")
        exitCode, err = internal.RunCommand("ezfabricctl workload init --input " + ezlabFilesDir + "/workload-init.yaml --kubeconfig " + orchestratorKubeConfig)
        if err != nil {
            log.Printf("Error: %v\n", err)
        } else {
            log.Printf("Workload init with exit code %d\n", exitCode)
        }
        // TODO: Process non-zero exit code


    },
}
