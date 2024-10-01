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

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
    Use:   "deploy",
    Short: "Update yaml files and start the installer",
    Run: func(cmd *cobra.Command, args []string) {
        log.Println("Starting deployment")

        appConfig := internal.GetAppConfiguration()
        clusterName := strings.Split(appConfig.Domain, ".")[0]
        ezlabFilesDir := "/tmp/ez-" + clusterName
        orchestratorKubeConfig := ezlabFilesDir + "/mgmt-kubeconfig"
        workloadKubeConfig := ezlabFilesDir + "/mgmt-kubeconfig"

        // Setup auth data for admin user
        authData := map[string]interface{}{
            "admin_user": map[string]string{
                "fullname": "Ezmeral Admin",
                "email":    fmt.Sprintf("ezadmin@%s", appConfig.Domain),
                "username": "ezua",
                "password": appConfig.Password,
            },
        }
        // Convert authData to JSON
        authDataJSON, err := json.Marshal(authData)
        if err != nil {
            fmt.Println("Error converting authData to JSON:", err)
            return
        }
        // log.Println("Auth data: " + string(authDataJSON))

        dfConfig := internal.GetMaprConfig()
        log.Printf("DEBUG: %v", dfConfig.TenantTicket)

        uaConfig := internal.UADeployConfig{
            Username: appConfig.Username,
            Password: appConfig.Password,
            // Password: base64.StdEncoding.EncodeToString([]byte(appConfig.Password)),
            Domain: appConfig.Domain,
            RegistryUrl: appConfig.RegistryUrl,
            RegistryInsecure: appConfig.RegistryInsecure,
            RegistryUsername: appConfig.RegistryUsername,
            RegistryPassword: appConfig.RegistryPassword,
            RegistryCa: "",
            // Orchestrator: internal.GetOutboundIP(),
            Orchestrator: appConfig.Orchestrator.IP,
            Master: appConfig.Controller.IP,
            Workers: strings.Split(internal.GetWorkerIPs(), ","),
            ClusterName: clusterName,
            AuthData: base64.StdEncoding.EncodeToString(authDataJSON),
            NoProxy: "10.96.0.0/12,10.224.0.0/16,10.43.0.0/16,192.168.0.0/16,.external.hpe.local,localhost,.cluster.local,.svc,.default.svc,127.0.0.1,169.254.169.254," + internal.GetWorkerIPs() + "," + appConfig.Controller.IP + "," + appConfig.Orchestrator.IP + ",." + appConfig.Domain,
            DF: dfConfig,
        }

        err = os.MkdirAll(ezlabFilesDir,0755)
        if err!= nil {
            log.Fatal("Error creating yaml directory: ",err)
        }

        // yamlfiles, err := internal.GetTemplateFiles()
        // if err != nil {
        //     fmt.Println("Error:", err)
        //     return
        // }

        deploySteps := map[string]string{
            "prechecks":        "00-prechecks.yaml",
            "fabricinit":       "01-fabricctl-init.yaml",
            "workloadprepare":  "02-workload-prepare.yaml",
            "workloaddeploy":   "03-workload-deploy.yaml",
            "fabriccluster":    "04-ezfabric-cluster.yaml",
        }

        for _, file := range deploySteps {
            // go func(f string) {
                internal.ProcessTemplate("templates/" + file, ezlabFilesDir + "/" + filepath.Base(file), uaConfig)
                log.Println("Processing: " + file)
            // }(file)
        }

        log.Println("YAML files ready")

        log.Fatalln("Stop here")

        log.Println("Running prechecks...")
        precheckCmd := "ezfabricctl prechecks --input " + ezlabFilesDir + "/" + deploySteps["prechecks"] + " --parallel=true --cleanup=true"
        log.Println(precheckCmd)
        exitCode, err := internal.RunCommand(precheckCmd)
        if err != nil {
            log.Fatalf("Error: %v\n", err)
        } else {
            log.Printf("Prechecks finished with exit code %d\n", exitCode)
            if exitCode != 0 {
                log.Fatalln("Prechecks failed")
             }
        }

        log.Println("Initializing the orchestrator fabric...")
        orchInitCmd := "ezfabricctl orchestrator init --input " + ezlabFilesDir + "/" + deploySteps["fabricinit"] + "--releasepkg /usr/local/share/applications/ezfab-release.tgz --save-kubeconfig " + orchestratorKubeConfig
        log.Println(orchInitCmd)
        exitCode, err = internal.RunCommand(orchInitCmd)
        if err != nil {
            log.Fatalf("Error: %v\n", err)
        } else {
            log.Printf("Orchestrator init finished with exit code %d\n", exitCode)
            if exitCode != 0 {
                log.Fatalln("Orchestrator init failed")
            }
        }

        // log.Println("Adding workload hosts to fabric...")
        // poolHostCmd := "ezfabricctl poolhost init --input " + ezlabFilesDir + "/hosts-init.yaml --kubeconfig " + orchestratorKubeConfig
        // log.Println(poolHostCmd)
        // exitCode, err = internal.RunCommand(poolHostCmd)
        // if err != nil {
        //     log.Fatalf("Error: %v\n", err)
        // } else {
        //     log.Printf("Hosts init with exit code %d\n", exitCode)
        //     if exitCode != 0 {
        //         log.Fatal("Workload hosts init failed")
        //     }
        // }

        // log.Println("Creating workload cluster...")
        // workloadInitCmd := "ezfabricctl workload init --input " + ezlabFilesDir + "/workload-init.yaml --kubeconfig " + orchestratorKubeConfig
        // log.Println(workloadInitCmd)
        // exitCode, err = internal.RunCommand(workloadInitCmd)
        // if err != nil {
        //     log.Fatalf("Error: %v\n", err)
        // } else {
        //     log.Printf("Workload init with exit code %d\n", exitCode)
        //     if exitCode != 0 {
        //         log.Fatal("Workload cluster init failed")
        //     }
        // }

        // Define kubectl command against orchestrator cluster
        kubeOrch := "kubectl --kubeconfig=" + orchestratorKubeConfig

        log.Println("Prepare for workload cluster")
        workloadPrepareCmd := kubeOrch + " create ns " + clusterName + ";" + kubeOrch + " apply -f " + ezlabFilesDir + "/" + deploySteps["workloadprepare"]
        log.Println(workloadPrepareCmd)
        exitCode, err = internal.RunCommand(workloadPrepareCmd)
        if err != nil {
            log.Fatalf("Error: %v\n", err)
        } else {
            log.Println("Secrets created for the workload cluster.")
            if exitCode !=  0 {
                log.Fatal("Secrets failed to create")
             }
        }

        log.Println("Create workload deploy CR")
        workloadDeployCmd := kubeOrch + " apply -f " + ezlabFilesDir + "/"  + deploySteps["workloaddeploy"]
        log.Println(workloadDeployCmd)
        exitCode, err = internal.RunCommand(workloadDeployCmd)
        if err != nil {
            log.Fatalf("Error: %v\n", err)
        } else {
            log.Println("Workload CR applied.")
            if exitCode !=  0 {
                log.Fatal("Workload deploy CR failed")
             }
        }

        log.Println("Create workload deploy CR")
        workloadCreateCmd := kubeOrch + " apply -f " + ezlabFilesDir + "/"  + deploySteps["fabriccluster"]
        log.Println(workloadCreateCmd)
        exitCode, err = internal.RunCommand(workloadCreateCmd)
        if err != nil {
            log.Fatalf("Error: %v\n", err)
        } else {
            log.Println("Created workload cluster.")
            if exitCode !=  0 {
                log.Fatal("Workload deploy failed")
             }
        }

        // workloadStatusCmd := "kubectl --kubeconfig=" + orchestratorKubeConfig + " get ezfabriccluster/" + clusterName + " -n " + clusterName + " -o json | jq -r '.status.workloadOpStatus.status'"
        WorkloadKubeconfigCmd := kubeOrch + " get secret " + clusterName + "-kubeconfig -n " + clusterName + " -o json | jq -r '.data.value' | base64 -d > " + workloadKubeConfig
        log.Println(workloadCreateCmd)
        exitCode, err = internal.RunCommand(WorkloadKubeconfigCmd)
        if err != nil {
            log.Fatalf("Error: %v\n", err)
        } else {
            log.Println("Saved workload kubeconfig.")
            if exitCode !=  0 {
                log.Fatal("Workload kubeconfig save failed")
             }
        }

        log.Println("EzUA deployed successfully!")
    },
}
