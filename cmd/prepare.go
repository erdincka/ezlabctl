package cmd

import (
	"ezlab/internal"
	"fmt"
	"log"
	"strings"
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
                "echo installing packages...",
                "sudo dnf install -y nfs-utils policycoreutils-python-utils conntrack-tools jq tar sshpass",
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

var setupDfCmd = &cobra.Command{
    Use:   "setup-df",
    Short: "Configure external Data Fabric for UA deployment",
    Run: func(cmd *cobra.Command, args []string) {
        log.Println("Configure Data Fabric for UA...")

    // Get user input for DF
    appConfig, err := internal.GetDFInput()
    if err != nil {
        log.Fatalf("failed to get configuration for Data Fabric: %v", err)
    }
    // log.Printf("Got configuration with %s", df)

    filesToTransfer := []string{
        "maprtenantticket",
        "cldb_nodes.json",
        "rest_nodes.json",
        "s3_nodes.json",
        "s3_keys.json",
    }

    commands := []string{
        // "( id ezua && id -nG ezua | grep -s $(id -un ezua) ) || sudo useradd -m -s /bin/bash -U ezua",
        // "id ezua || sudo useradd -m -U ezua",
        "id ezua 2>&1 >/dev/null || sudo useradd -m -U ezua 2>&1 > /dev/null",
        "echo ezua: " + appConfig.DFPass + " | sudo chpasswd",
        "echo " + appConfig.DFPass + " | maprlogin password -user " + appConfig.DFUser,
        "echo Setting up the volumes for UA...",
        "maprcli acl edit -type cluster -user ezua:login,cv",
        "maprcli volume create -name ezua-base-volume -path /ezua -type rw -json -rootdiruser ezua -rootdirgroup ezua -createparent 1 || true",
        "[ -f /tmp/maprtenantticket ] || maprlogin generateticket -type tenant -user ezua -out /tmp/maprtenantticket",
        "[ -f /tmp/cldb_nodes.json ] || /opt/mapr/bin/maprcli node list -columns hn -filter svc==cldb -json > /tmp/cldb_nodes.json",
        "[ -f /tmp/rest_nodes.json ] || /opt/mapr/bin/maprcli node list -columns hn -filter svc==apiserver -json > /tmp/rest_nodes.json",
        "[ -f /tmp/s3_nodes.json ] || /opt/mapr/bin/maprcli node list -columns hn -filter svc==s3server -json > /tmp/s3_nodes.json",
        "[ -f /tmp/s3_keys.json ] || maprcli s3keys generate -domainname primary -accountname default -username ezua -json > /tmp/s3_keys.json",
        "sudo chown mapr:mapr /tmp/" + strings.Join(filesToTransfer, " /tmp/"),
    }

    var wg sync.WaitGroup // Create a WaitGroup
    wg.Add(1)
    go internal.SSHCommands(appConfig.DFHost, appConfig.Username, appConfig.Password, commands, &wg)
    wg.Wait()

    // Perform SCP transfer
	go internal.SCPGetFiles(appConfig.DFHost, appConfig.DFUser, appConfig.DFPass, "/tmp", "/tmp", filesToTransfer)
    log.Println("Files transferred.")

    log.Println("DF Setup is done!")
    },
}

func init() {
    rootCmd.AddCommand(prepareCmd)
    rootCmd.AddCommand(setupDfCmd)
}