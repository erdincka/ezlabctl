package cmd

import (
	"encoding/base64"
	"encoding/json"
	"ezlabctl/internal"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var sshuser, sshpass, domain, dfhost, dfuser, dfpass, registryCa, registryUrl, registryUsername, registryPassword string
var master net.IP
var workers []net.IP
var uaConfig internal.UAConfig
var registryInsecure bool

func init() {
	rootCmd.AddCommand(unifiedAnalyticsCmd)
	unifiedAnalyticsCmd.Flags().BoolP("configure", "c", false, "Run pre-install configuration steps")
	unifiedAnalyticsCmd.Flags().BoolP("attach", "a", false, "Prepare external storage to attach to UA")
	unifiedAnalyticsCmd.Flags().BoolP("template", "t", false, "Update yaml templates for deployment")
	unifiedAnalyticsCmd.Flags().BoolP("validate", "v", false, "Verify node readiness for deployment")
	unifiedAnalyticsCmd.Flags().BoolP("orchinit", "o", false, "Install orchestrator on this node")
	unifiedAnalyticsCmd.Flags().Bool("confirm", false, "Confirm deployment for workload cluster")
    unifiedAnalyticsCmd.PersistentFlags().StringVarP(&sshuser, "sshuser", "u", "ezmeral", "SSH User")
    unifiedAnalyticsCmd.PersistentFlags().StringVarP(&sshpass, "sshpass", "p", "", "SSH Password")
	unifiedAnalyticsCmd.Flags().StringVarP(&domain, "domain", "d", "ua.my.lab", "UA Domain")
	unifiedAnalyticsCmd.Flags().StringVar(&dfhost, "dfhost", "", "EDF Server")
	unifiedAnalyticsCmd.Flags().StringVar(&dfuser, "dfuser", "mapr", "EDF Admin user")
	unifiedAnalyticsCmd.Flags().StringVar(&dfpass, "dfpass", "", "EDF Admin password")
	unifiedAnalyticsCmd.Flags().IPVarP(&master, "master", "m", nil, "UA Master")
	unifiedAnalyticsCmd.Flags().IPSliceVarP(&workers, "worker", "w", nil, "UA Workers")
	unifiedAnalyticsCmd.Flags().StringVar(&registryUrl, "registryUrl", "", "UA Registry URL")
	unifiedAnalyticsCmd.Flags().StringVar(&registryUsername, "registryUsername", "", "UA Registry Username")
	unifiedAnalyticsCmd.Flags().StringVar(&registryPassword, "registryPassword", "", "UA Registry Password")
	unifiedAnalyticsCmd.Flags().StringVar(&registryCa, "registryCa", "", "UA Registry CA base64 encoded")
	unifiedAnalyticsCmd.Flags().BoolVar(&registryInsecure, "registryInsecure", true, "UA Registry Insecure")
}

var unifiedAnalyticsCmd = &cobra.Command{
    Use:   	"ua",
	Short: 	"Prepare, Configure and Install Unified Analytics",
	Long: 	"see README.md",
	PreRun: func (cmd *cobra.Command, args []string)  {
		// Ensure credentials for remote hosts
		// master, _ := cmd.Flags().GetIP("master")
		// workers, _ := cmd.Flags().GetIPSlice("worker")
		if master != nil || len(workers) > 0 {
			cmd.MarkFlagRequired("sshuser")
			cmd.MarkFlagRequired("sshpass")
		}

		attach, _ := cmd.Flags().GetBool("attach")
		if attach {
			cmd.MarkFlagRequired("dfhost")
			cmd.MarkFlagRequired("dfuser")
			cmd.MarkFlagRequired("dfpass")
		}

		template, _ := cmd.Flags().GetBool("template")
		if template {
			cmd.MarkFlagRequired("domain")
			cmd.MarkFlagRequired("master")
			cmd.MarkFlagRequired("worker")
			input, _ := cmd.Flags().GetIPSlice("worker")
			if len(input) < 3 {
				log.Fatal("Need at least three workers")
			}
		}

		validate, _ := cmd.Flags().GetStringSlice("validate")
		if template {
			cmd.MarkFlagRequired("master")
			cmd.MarkFlagRequired("worker")
			input, _ := cmd.Flags().GetIPSlice("worker")
			if len(input) < 3 {
				log.Fatal("Need at least three workers")
			}
		}

	},
    Run: func(cmd *cobra.Command, args []string) {
		var err error = nil

		host := internal.GetOutboundIP()

		// Check root privileges
		internal.IfRoot("Setting up host as Unified Analytics orchestrator")

		// Validate host resolution
		orchNode, err := internal.ResolveNode(host)
		if err != nil {
			log.Fatal("failed to validate this host: %w", err)
		}
		orch := *orchNode

		// Check and execute if configuration requested
		if cmd.Flags().Changed("configure") {
			log.Println("Preinstall setup...")

			// Commands required only for UA
			extraCommands := []string {
				// TODO: update for RHEL 9
				"( dnf repolist | grep -q HighAvailability ) || sudo subscription-manager repos --enable=rhel-8-for-x86_64-highavailability-rpms",
				"sudo systemctl add-wants default.target rpc-statd.service",
			}

			var wg sync.WaitGroup
			// Prep orch host locally
			log.Println("Configuring host:", orch.FQDN)
			wg.Add(1)
			go internal.Preinstall(orch.FQDN, extraCommands, &wg)

			// Configure master if set
			if cmd.Flags().Changed("master") {
				// Get master IP
				input, err := cmd.Flags().GetIP("master")
				if err != nil {
					log.Fatal("Failed to parse master: %w", err)
				}

				log.Printf("Checking master: %s", input)

				// Validate node
				node, err := internal.ResolveNode(input.To4().String())
				if err != nil {
					log.Fatal("Failed to resolve master: %w", err)
				}

				// Validate connectivity and sudo permissions
				internal.TestCredentials(node.IP, &sshuser, &sshpass)
				// Configure master
				internal.PreinstallOverSsh(node.FQDN, sshuser, sshpass)
			}

			// Configure workers if set
			if cmd.Flags().Changed("worker") {
				input, err :=  cmd.Flags().GetIPSlice("worker")
				if err != nil {
					log.Fatal("Failed to parse workers:", err)
				}
				for _, worker := range input {
					if err != nil {
						log.Fatal("Failed to scan workers", err)
					}
					log.Printf("Checking worker: %s", worker)

					// Validate
					node, err := internal.ResolveNode(worker.To4().String())
					if err != nil {
						log.Fatalf("Failed to resolve worker: %v %v", node, err)
					}

					// Validate connectivity and sudo permissions
					internal.TestCredentials(node.IP, &sshuser, &sshpass)
					// Configure master
					internal.PreinstallOverSsh(node.FQDN, sshuser, sshpass)
				}

			}

			wg.Wait()
		}

		// Check and execute if EDF configuration requested
		if cmd.Flags().Changed("attach") {
			log.Println("Configuring EDF to use by UA...")
			PrepareEDF(cmd)
		}

		clusterName := strings.Split(domain, ".")[0]
		templateFiles := internal.TemplateFiles{
			TemplateDirectory: "/tmp/ez-" + clusterName,
			OrchestratorKubeConfig: "/tmp/ez-" + clusterName + "/mgmt-kubeconfig",
			WorkloadKubeConfig: "/tmp/ez-" + clusterName + "/workload-kubeconfig",
		}

		// Setup auth data for admin user
		authData := map[string]interface{}{
			"admin_user": map[string]string{
				"fullname": "Ezmeral Admin",
				"email":    fmt.Sprintf("admin@%s", domain),
				"username": "admin",
				"password": sshpass,
			},
		}
		// Convert authData to JSON
		authDataJSON, err := json.Marshal(authData)
		if err != nil {
			log.Fatal("Error converting authData to JSON:", err)
		}

		workerIps := []string{}
		for _, node := range workers {
			workerIps = append(workerIps, node.String())
		}

		// TODO: check all required vars are set
		uaConfig = internal.UAConfig{
			Username: sshuser,
			Password: sshpass,
			// Password: base64.StdEncoding.EncodeToString([]byte(appConfig.Password)),
			Domain: domain,
			RegistryUrl: registryUrl,
			RegistryInsecure: registryInsecure,
			RegistryUsername: registryUsername,
			RegistryPassword: registryPassword,
			RegistryCa: registryCa,
			// Orchestrator: GetOutboundIP(),
			Orchestrator: orch.IP,
			Master: master.String(),
			Workers: workerIps,
			ClusterName: clusterName,
			AuthData: base64.StdEncoding.EncodeToString(authDataJSON),
			NoProxy: "10.96.0.0/12,10.224.0.0/16,10.43.0.0/16,192.168.0.0/16,.external.hpe.local,localhost,.cluster.local,.svc,.default.svc,127.0.0.1,169.254.169.254," + strings.Join(workerIps, ",") + "," + master.String() + "," + orch.IP + ",." + domain,
			DF: internal.GetMaprConfig(),
		}

		if cmd.Flags().Changed("template") {
			log.Println("Configure templates...")
			internal.ProcessTemplates(templateFiles.TemplateDirectory, uaConfig)
		}

		deploySteps := internal.GetDeploySteps()

		if cmd.Flags().Changed("validate") {
            log.Println("Running prechecks...")
            precheckCmd := "/usr/local/bin/ezfabricctl prechecks --input " + templateFiles.TemplateDirectory + "/" + deploySteps["prechecks"] + " --parallel=true --cleanup=true"
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

		if cmd.Flags().Changed("orchinit") {
			log.Println("Initializing orchestrator on host:", orch.FQDN)
			orchInitCmd := "/usr/local/bin/ezfabricctl orchestrator init --input " + templateFiles.TemplateDirectory + "/" + deploySteps["fabricinit"] + " --releasepkg /usr/local/share/applications/ezfab-release.tgz --save-kubeconfig " + templateFiles.OrchestratorKubeConfig
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
		kubeOrch := "kubectl --kubeconfig=" + templateFiles.OrchestratorKubeConfig

		// confirm if kubeconf file for orchestrator is available
		_, err = os.Stat(templateFiles.OrchestratorKubeConfig)
		if os.IsNotExist(err) {
			log.Fatalf("Orchestrator kubeconfig %s does not exist.\n", templateFiles.OrchestratorKubeConfig)
		} else if err != nil {
			log.Fatalf("An error occurred while checking the orchestrator kubeconfig file: %v\n", err)
		}

		// DEBUG:
		// fmt.Printf("%+v\n", uaConfig)

		if cmd.Flags().Changed("confirm") {
			log.Print("Apply secrets...")
			applyWorkloadPrepareCmd := kubeOrch + " apply -f " + templateFiles.TemplateDirectory + "/" + deploySteps["workloadprepare"]
			log.Printf("$ %s", applyWorkloadPrepareCmd)
			exitCode, err := internal.RunCommand(applyWorkloadPrepareCmd)
			if err != nil {
				log.Fatalf("Error: %v\n", err)
				} else {
					if exitCode !=  0 {
						log.Fatal("Secrets creation failed!")
					}
				}

			log.Print("Apply workload deploy CR...")
			workloadDeployCmd := kubeOrch + " apply -f " + templateFiles.TemplateDirectory + "/" + deploySteps["workloaddeploy"]
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
			workloadCreateCmd := kubeOrch + " apply -f " + templateFiles.TemplateDirectory + "/" + deploySteps["fabriccluster"]
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
		}
	},
}
