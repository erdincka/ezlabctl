package cmd

import (
	"ezlabctl/internal"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var unifiedAnalyticsCmd = &cobra.Command{
    Use:   	"ua",
	Short: 	"Install Unified Analytics, run prechecks if -p flag is set, install orchestrator if -o flag is set etc",
	Long: 	"",
	// Args: cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
		var err error = nil
		host := internal.GetOutboundIP()

		if os.Geteuid() != 0 {
			log.Fatalf("You must be root to run this command. Now using UID: %d", os.Geteuid())
		} else {
			log.Println("Setting up host as Unified Analytics orchestrator")
		}

		orchNode, err := internal.ResolveNode(host)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to validate host: %w", err))
		}
		orch := *orchNode
		log.Println("Initializing orchestrator on host:", orch.FQDN)

		if cmd.Flags().Changed("configure") {
			log.Println("Preinstall setup...")
			internal.Preinstall(orch.FQDN)
			// Run additional commands for UA
			commands := []string {
				"( dnf repolist | grep -q HighAvailability ) || sudo subscription-manager repos --enable=rhel-8-for-x86_64-highavailability-rpms",
				"sudo systemctl add-wants default.target rpc-statd.service",
			}
			for _, command := range commands {
				exitCode, err := internal.RunCommand(command)
				if err!= nil {
					log.Fatal("Failed with subscription manager")
				}
				if exitCode!= 0 {
					log.Fatal("Failed with subscription manager")
				}
			}
		}

		if cmd.Flags().Changed("storage") {
			log.Println("Configure storage...")

			var dfhost, dfuser, dfpass string
			// Get EDF host
			if cmd.Flags().Changed("dfhost") {
				dfhost, err = cmd.Flags().GetString("dfhost")
				if err != nil {
					log.Fatal("failed to get host value: %w", err)
				}
			} else {
				dfhost = internal.AskForInput("EDF Host: ", "")
			}

			// Get EDF user
			if cmd.Flags().Changed("dfuser") {
				dfuser, err = cmd.Flags().GetString("dfuser")
				if err != nil {
					log.Fatal("failed to get user value: %w", err)
				}
			} else {
				dfuser = internal.AskForInput("EDF SSH user: ", "")
			}

			// Get EDF password
			if cmd.Flags().Changed("dfpass") {
				dfpass, err = cmd.Flags().GetString("dfpass")
				if err != nil {
					log.Fatal("failed to get pass value: %w", err)
				}
			} else {
				dfpass = internal.AskForInput("EDF SSH password: ", "")
			}

			log.Print(dfhost, dfuser, dfpass)
		}

		if cmd.Flags().Changed("template") {
			log.Println("Configure templates...")

		}

		if cmd.Flags().Changed("master") {
			log.Println("Setup master node...")
			input, err := cmd.Flags().GetString("master")
			if err != nil {
				log.Fatal("failed to get master value: %w", err)
			}
			masterNode, err := internal.ResolveNode(input)
			if err != nil {
				log.Fatal(fmt.Errorf("failed to validate host: %w", err))
			}
			master := *masterNode
			log.Print(master)
			// PreinstallSSH
			// internal.Preinstall(master.FQDN)

			// var wg sync.WaitGroup // Create a WaitGroup
			// // Test connection to all nodes
			// for _, node := range append(workers, controller, orchestrator) {
			// 	wg.Add(1)
			// 	go testCredentials(node, &username, &password, &wg)
			// }
			// wg.Wait()


		}

		if cmd.Flags().Changed("workers") {
			log.Println("Setup worker nodes...")
		}

		log.Print("Finished.")
	},
}

func init() {
	rootCmd.AddCommand(unifiedAnalyticsCmd)
	unifiedAnalyticsCmd.Flags().BoolP("configure", "c", false, "Run pre-install configuration steps")
	unifiedAnalyticsCmd.Flags().BoolP("storage", "s", false, "Prepare external storage")
	unifiedAnalyticsCmd.Flags().BoolP("template", "t", false, "Update yaml templates")
    unifiedAnalyticsCmd.Flags().StringP("username", "u", "", "SSH User")
    unifiedAnalyticsCmd.Flags().StringP("password", "p", "", "SSH Password")
	unifiedAnalyticsCmd.Flags().StringP("domain", "d", "ez.company.lab", "UA Domain")
	unifiedAnalyticsCmd.Flags().StringP("master", "m", "", "UA Master")
	unifiedAnalyticsCmd.Flags().StringArrayP("worker", "w", []string{}, "UA Workers")
}