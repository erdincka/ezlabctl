package cmd

import (
	"ezlabctl/internal"
	"fmt"
	"log"
	"sync"

	"github.com/spf13/cobra"
)

// prepareCmd represents the prepare command
var prepareCmd = &cobra.Command{
    Use:   "prepare",
    Short: "Prepare the system for deployment",
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
        var wg sync.WaitGroup // Create a WaitGroup

        for _, host := range allHosts {
            wg.Add(1) // Increment the WaitGroup counter

            commands := []string{
                "sudo hostnamectl set-hostname " + host,
                "sudo localectl set-locale LANG=en_US.UTF-8",
                "sudo timedatectl set-timezone " + nodes.Timezone,
                "echo '" + nodes.Username + " ALL=(ALL) NOPASSWD:ALL' | sudo tee /etc/sudoers.d/" + nodes.Username,
                "sudo sed -i '/^::1/d' /etc/hosts",
                "sudo sed -i '/" + host + "/d' /etc/hosts",
                "sudo sed -i 's/myhostname//g' /etc/nsswitch.conf",
                "sudo subscription-manager repos --enable=rhel-8-for-x86_64-highavailability-rpms",
                "echo installing required packages... RPM DEPENDS CLUASE SHOULD REPLACE THIS",
                "sudo systemctl enable --now rpc-statd", // FIX: mount option nolock is not propogated here
                // "sudo dnf install -yq firewalld sshpass",
                // "sudo sed -i 's/^FirewallBackend.*$/FirewallBackend=iptables/' /etc/firewalld/firewalld.conf",
                // "sudo modprobe ip_tables",
                "echo Updating packages, might take a while...",
                "sudo dnf upgrade -yq",
                "echo preinstall finished for " + host,
            }
            // Run the commands using goroutines
            go internal.SSHCommands(host, nodes.Username, nodes.Password, commands, &wg) // Launch the SSH command in a new goroutine
        }
        wg.Wait() // Wait for all goroutines to finish
        log.Println("Hosts ready for installation!")
    },
}
