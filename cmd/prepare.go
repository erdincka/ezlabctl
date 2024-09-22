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
        }(append(nodes.Workers, nodes.Controller))

        // Run commands on all hosts
        var wg sync.WaitGroup // Create a WaitGroup

        for _, host := range allHosts {
            wg.Add(1) // Increment the WaitGroup counter

            commands := []string{
                "sudo hostnamectl set-hostname " + host,
                "sudo sed -i '/^::1/d' /etc/hosts",
                "sudo sed -i 's/myhostname//g' /etc/nsswitch.conf",
                "sudo localectl set-locale LANG=en_US.UTF-8",
                "echo installing packages... RPM DEPENDS CLUASE SHOULD REPLACE THIS",
                "sudo dnf install -y nfs-utils policycoreutils-python-utils conntrack-tools iptables-services iptables-utils jq tar sshpass",
                "sudo systemctl enable --now iptables",
                "sudo dnf --setopt=tsflags=noscripts install -y -q iscsi-initiator-utils",
                "echo \"InitiatorName=$(/sbin/iscsi-iname)\" | sudo tee /etc/iscsi/initiatorname.iscsi",
                "sudo systemctl enable --now iscsid; sudo systemctl restart iscsid",
                "sudo /sbin/ethtool -K eth0 tx-checksum-ip-generic off",
                "echo preinstall finished for " + host,
            }
            // Run the commands using goroutines
            go internal.SSHCommands(host, nodes.Username, nodes.Password, commands, &wg) // Launch the SSH command in a new goroutine
        }
        wg.Wait() // Wait for all goroutines to finish
        log.Println("Hosts ready for installation!")
    },
}
