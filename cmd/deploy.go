package cmd

import (
	"ezlabctl/internal"
	"log"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
    Use:   "deploy",
    Short: "Update yaml files and start the installer",
    Run: func(cmd *cobra.Command, args []string) {

        clusterName, ezlabFiles, deployConf := internal.GetDeployConfig()
        deploySteps := internal.GetDeploySteps()

        if clusterName == "" {
            log.Fatal("Cluster name is required! Run prepare and storage first!!!")
        }

        log.Printf("Starting deployment for %s\n", clusterName)

        // Check if the templating flag is provided
        if cmd.Flags().Changed("template") {
            log.Println("Recreate templates...")
            internal.ProcessTemplates(ezlabFiles, deployConf)
        }

        if cmd.Flags().Changed("prechecks") {
            log.Println("Running prechecks...")
            precheckCmd := "ezfabricctl prechecks --input " + ezlabFiles.TemplateDirectory + "/" + deploySteps["prechecks"] + " --parallel=true --cleanup=true"
            log.Println(precheckCmd)
            exitCode, err := internal.RunCommand(precheckCmd)
            if err != nil {
                log.Fatalf("Error: %v\n", err)
            } else {
                log.Printf("Prechecks finished with exit code %d\n", exitCode)
                if exitCode != 0 {
                    log.Fatalln("Prechecks failed!")
                }
            }
        }

        if cmd.Flags().Changed("init") {
            log.Println("Initializing the orchestrator fabric...")
            orchInitCmd := "ezfabricctl orchestrator init --input " + ezlabFiles.TemplateDirectory + "/" + deploySteps["fabricinit"] + "--releasepkg /usr/local/share/applications/ezfab-release.tgz --save-kubeconfig " + ezlabFiles.OrchestratorKubeConfig
            log.Println(orchInitCmd)
            exitCode, err := internal.RunCommand(orchInitCmd)
            if err != nil {
                log.Fatalf("Error: %v\n", err)
            } else {
                log.Printf("Orchestrator init finished with exit code %d\n", exitCode)
                if exitCode != 0 {
                    log.Fatalln("Orchestrator init failed!")
                }
            }
        }
        // Define kubectl command against orchestrator cluster
        kubeOrch := "kubectl --kubeconfig=" + ezlabFiles.OrchestratorKubeConfig

        log.Printf("Create namespace: %s...", clusterName)
        createNamespaceCmd := kubeOrch + " create ns " + clusterName
        log.Printf("$ %s", createNamespaceCmd)
        exitCode, err := internal.RunCommand(createNamespaceCmd)
        if err != nil {
            log.Fatalf("Error: %v", err)
        } else {
            if exitCode !=  0 {
                log.Fatal("Namespace creation failed!")
            }
        }

        log.Print("Apply secrets...")
        applyWorkloadPrepareCmd := kubeOrch + " apply -f " + ezlabFiles.TemplateDirectory + "/" + deploySteps["workloadprepare"]
        log.Printf("$ %s", applyWorkloadPrepareCmd)
        exitCode, err = internal.RunCommand(applyWorkloadPrepareCmd)
        if err != nil {
            log.Fatalf("Error: %v\n", err)
            } else {
                if exitCode !=  0 {
                    log.Fatal("Secrets creation failed!")
                }
            }

        log.Print("Apply workload deploy CR...")
        workloadDeployCmd := kubeOrch + " apply -f " + ezlabFiles.TemplateDirectory + "/" + deploySteps["workloaddeploy"]
        log.Printf("$ %s",workloadDeployCmd)
        exitCode, err = internal.RunCommand(workloadDeployCmd)
        if err != nil {
            log.Fatalf("Error: %v\n", err)
        } else {
            if exitCode !=  0 {
                log.Fatal("Workload deploy CR failed!")
            }
        }

        log.Print("Create fabric cluster...")
        workloadCreateCmd := kubeOrch + " apply -f " + ezlabFiles.TemplateDirectory + "/" + deploySteps["fabriccluster"]
        log.Printf("$ %s", workloadCreateCmd)
        exitCode, err = internal.RunCommand(workloadCreateCmd)
        if err != nil {
            log.Fatalf("Error: %v\n", err)
        } else {
            if exitCode !=  0 {
                log.Fatal("Fabric creation failed!")
            }
        }

        log.Println("EzUA deployment started.")
    },
}


func init() {
	rootCmd.AddCommand(deployCmd)

    deployCmd.Flags().BoolP("template", "t", false, "Recreate templates")
    deployCmd.Flags().BoolP("prechecks", "p", false, "Run prechecks")
    deployCmd.Flags().BoolP("init", "i", false, "Initialize orchestrator")

}