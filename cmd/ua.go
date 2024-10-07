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
			_ = cmd.MarkFlagRequired("sshuser")
			_ = cmd.MarkFlagRequired("sshpass")
		}

		attach, _ := cmd.Flags().GetBool("attach")
		if attach {
			_ = cmd.MarkFlagRequired("dfhost")
			// _ = cmd.MarkFlagRequired("dfuser") # using default
			_ = cmd.MarkFlagRequired("dfpass")
		}

		template, _ := cmd.Flags().GetBool("template")
		if template {
			_ = cmd.MarkFlagRequired("domain")
			_ = cmd.MarkFlagRequired("master")
			_ = cmd.MarkFlagRequired("worker")
			input, _ := cmd.Flags().GetIPSlice("worker")
			if len(input) < 3 {
				log.Fatal("Need at least three workers")
			}
		}

		validate, _ := cmd.Flags().GetBool("validate")
		if validate {
			_ = cmd.MarkFlagRequired("master")
			_ = cmd.MarkFlagRequired("worker")
			input, _ := cmd.Flags().GetIPSlice("worker")
			if len(input) < 3 {
				log.Fatal("Need at least three workers")
			}
		}

	},

    Run: func(cmd *cobra.Command, args []string) {
		var err error

		isDryRun := ! cmd.Flags().Changed("confirm")

		host := internal.GetOutboundIP()

		// // Check root privileges
		// internal.IfRoot("Setting up host as Unified Analytics orchestrator")

		// Validate host resolution
		orchNode, err := internal.ResolveNode(host)
		if err != nil {
			log.Fatal("failed to validate this host: %w", err)
		}
		orch := *orchNode

		// Commands required only for UA
		extraCommands := []string {
			// TODO: update for RHEL 9
			"( dnf repolist | grep -q HighAvailability ) || sudo subscription-manager repos --enable=rhel-8-for-x86_64-highavailability-rpms",
			"sudo systemctl add-wants default.target rpc-statd.service",
		}

		// Check and execute if configuration requested
		if cmd.Flags().Changed("configure") {
			log.Println("Preinstall setup...")

			var wg sync.WaitGroup
			// Prep orch host locally
			log.Println("Configuring host:", orch.FQDN)
			wg.Add(1)
			go internal.Preinstall(orch.FQDN, extraCommands, &wg, isDryRun)

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
				err = internal.TestCredentials(node.IP, &sshuser, &sshpass); if err != nil { log.Fatal(err); }
				// Configure master
				wg.Add(1)
				go internal.PreinstallOverSsh(node.FQDN, sshuser, sshpass, &wg, isDryRun)
			} else {
				log.Printf("No master option, skipping master configuration.")
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
					err = internal.TestCredentials(node.IP, &sshuser, &sshpass); if err!= nil {
						log.Fatal(err)
					}
					// Configure worker
					wg.Add(1)
					go internal.PreinstallOverSsh(node.FQDN, sshuser, sshpass, &wg, isDryRun)
				}

			} else {
				log.Printf("No worker option, skipping worker configuration.")
			}

			wg.Wait()
		} else {
			log.Println("Skipping pre-install configuration.")
		}

		// Check and execute if EDF configuration requested
		if cmd.Flags().Changed("attach") {
			log.Println("Configuring EDF to use by UA...")
			PrepareEDF(cmd, isDryRun)
		} else {
			log.Println("Skipping EDF attach.")
			// PrepareEDF(cmd, true)
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
		} else {
			log.Println("Skipping templating.")
		}

		deploySteps := internal.GetDeploySteps()
		// Define commands
		precheckCmd := "/usr/local/bin/ezfabricctl prechecks --input " + templateFiles.TemplateDirectory + "/" + deploySteps["prechecks"] + " --parallel=true --cleanup=true"
		orchInitCmd := "/usr/local/bin/ezfabricctl orchestrator init --input " + templateFiles.TemplateDirectory + "/" + deploySteps["fabricinit"] + " --releasepkg /usr/local/share/applications/ezfab-release.tgz --save-kubeconfig " + templateFiles.OrchestratorKubeConfig
		kubeOrch := "kubectl --kubeconfig=" + templateFiles.OrchestratorKubeConfig + " " // add trailing space for remainder of command
		createNamespaceCmd:= fmt.Sprintf("%sget namespace %s || %screate namespace %s", kubeOrch, clusterName, kubeOrch, clusterName)
		applyWorkloadPrepareCmd := kubeOrch + "apply -f " + templateFiles.TemplateDirectory + "/" + deploySteps["workloadprepare"]
		workloadDeployCmd := kubeOrch + "apply -f " + templateFiles.TemplateDirectory + "/" + deploySteps["workloaddeploy"]
		workloadCreateCmd := kubeOrch + "apply -f " + templateFiles.TemplateDirectory + "/" + deploySteps["fabriccluster"]

		if cmd.Flags().Changed("validate") {
            log.Println("Running prechecks...")
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
        } else {
			log.Println("Skipping precheck validation.")
			log.Print("Skipped precheck command: ", precheckCmd)
		}

		if cmd.Flags().Changed("orchinit") && ! isDryRun {
			log.Println("Initializing orchestrator on host:", orch.FQDN)
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
		} else {
			log.Println("Skipping orchestrator initialization.")
			log.Print("Skipped orchestrator init command: ", orchInitCmd)
		}

		// confirm if kubeconf file for orchestrator is available
		_, err = os.Stat(templateFiles.OrchestratorKubeConfig)
		if os.IsNotExist(err) {
			log.Fatalf("Orchestrator kubeconfig %s is needed to continue!\n", templateFiles.OrchestratorKubeConfig)
		} else if err != nil {
			log.Fatalf("An error occurred while checking the orchestrator kubeconfig file: %v\n", err)
		}

		// DEBUG:
		// fmt.Printf("%+v\n", uaConfig)

		if cmd.Flags().Changed("confirm") {

			log.Print("Create namespace...")
			log.Printf("Running: %s", createNamespaceCmd)
			exitCode, err := internal.RunCommand(createNamespaceCmd)
			if err != nil {
				log.Fatalf("Error: %v\n", err)
			} else {
				if exitCode !=  0 {
					log.Fatal("Namespace creation failed!")
				}
			}

			log.Print("Apply secrets...")
			log.Printf("Running: %s", applyWorkloadPrepareCmd)
			exitCode, err = internal.RunCommand(applyWorkloadPrepareCmd)
			if err != nil {
				log.Fatalf("Error: %v\n", err)
			} else {
				if exitCode !=  0 {
					log.Fatal("Secrets creation failed!")
				}
			}

			log.Print("Apply workload deploy CR...")
			log.Printf("Running: %s",workloadDeployCmd)
			exitCode, err = internal.RunCommand(workloadDeployCmd)
			if err != nil {
				log.Fatalf("Error: %v\n", err)
			} else {
				if exitCode !=  0 {
					log.Fatal("Workload deploy CR failed!")
				}
			}

			log.Print("Create fabric cluster...")
			log.Printf("Running: %s", workloadCreateCmd)
			exitCode, err = internal.RunCommand(workloadCreateCmd)
			if err != nil {
				log.Fatalf("Error: %v\n", err)
			} else {
				if exitCode !=  0 {
					log.Fatal("Fabric creation failed!")
				}
			}

			log.Println("EzUA deployment started.")
		} else {
			log.Print("Skipping deploy CRs...")
			log.Printf("Skipped: %s\n", createNamespaceCmd)
			log.Printf("Skipped: %s\n", applyWorkloadPrepareCmd)
			log.Printf("Skipped: %s\n", workloadDeployCmd)
			log.Printf("Skipped: %s\n", workloadCreateCmd)
		}
		log.Println("Done.")
	},
}
