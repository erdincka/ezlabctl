package cmd

import (
	"ezlabctl/internal"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// prepareCmd represents the prepare command
var prepareCmd = &cobra.Command{
    Use:   "prepare (deprecated)",
    Short: "Deprecated - Use with caution! Prepare the UA systems for deployment",
    Run: func(cmd *cobra.Command, args []string) {

        log.Println("Prepare the nodes for installation...")

        // Get user input for UA hosts
        nodes, err := internal.GetUAInput()
        if err != nil {
            _ = fmt.Errorf("failed to get configuration: %w", err)
        }

        // Connect to the remote systems and prepare
        allHosts := func(hosts []internal.Node) []string {
            fqdns := []string{}
            for _, host := range hosts {
                fqdns = append(fqdns, host.FQDN)
            }
            return fqdns
        }(append(nodes.Workers, nodes.Controller, nodes.Orchestrator))

        // Run commands on all hosts
        for _, host := range allHosts {
            commands := internal.PrepareCommands(host)
            // UA requirements below
            commands = append(commands,
                "sudo subscription-manager repos --enable=rhel-8-for-x86_64-highavailability-rpms",
                "sudo systemctl enable --now rpc-statd", // FIX: mount option nolock is not propogated here
            )
            // Run the commands using goroutines
            go internal.SshCommands(host, nodes.Username, nodes.Password, commands) // Launch the SSH command in a new goroutine
        }
        log.Println("Hosts ready for installation!")
    },
}

func init() {
    // rootCmd.AddCommand(prepareCmd)
}