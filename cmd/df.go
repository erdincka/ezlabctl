package cmd

import (
	"ezlabctl/internal"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var datafabricCmd = &cobra.Command{
    Use:   "df",
	Short: "Install Data Fabric, runs pre-install configuration if -c flag is set, deploys installer if -i flag is set",
	// Args: cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {

		var err error = nil
		host := internal.GetOutboundIP()

		if os.Geteuid() != 0 {
			log.Fatalf("You must be root to run this command. Now using UID: %d", os.Geteuid())
		} else {
			log.Println("Setting up host for Data Fabric installation")
		}

		dfNode, err := internal.ResolveNode(host)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to validate host: %w", err))
		}
		df := *dfNode

		if cmd.Flags().Changed("configure") {
			log.Println("Preinstall setup...")
			internal.Preinstall(df.FQDN)
		}

		if cmd.Flags().Changed("installer") {
			var username, password, repo, disk string
			// Need credentials and repo if installer is set
			username = internal.GetStringInput(cmd, "username", "Username", "root")
			password = internal.GetStringInput(cmd, "password", "Password", "")
			disk = internal.GetStringInput(cmd, "disk", "Data Disk", "/dev/sdb")
			repo = internal.GetStringInput(cmd, "repo", "MapR Repo", "https://package.ezmeral.hpe.com/releases")
			// TODO: check if repo is accessible and ask for credentials if auth needed

			log.Println("Deploying mapr-installer...")
			Installer(df.FQDN, username, password, repo, disk)
		}

		exitCode, err := internal.RunCommand("[ -f /opt/mapr/conf/mapr-clusters.conf ] || sudo /opt/mapr/installer/bin/mapr-installer-cli install -nvpf -t /tmp/mapr-stanza.yaml -u mapr:mapr@127.0.0.1:9443")
		if err != nil {
			log.Fatal("Error running installer-cli:", err)
		}
		if exitCode != 0 {
			log.Fatal("Error running installer-cli:", exitCode)
		}

		log.Printf("EDF deployed on %s\n", df.FQDN)
    },
}

func init() {
	rootCmd.AddCommand(datafabricCmd)
	datafabricCmd.Flags().BoolP("configure", "c", false, "Run pre-install configuration steps")
    datafabricCmd.Flags().BoolP("installer", "i", false, "Deploy mapr-installer")
    datafabricCmd.Flags().StringP("username", "u", "", "SSH User")
    datafabricCmd.Flags().StringP("password", "p", "", "SSH Password")
    datafabricCmd.Flags().StringP("repo", "r", "https://package.ezmeral.hpe.com/releases", "MapR Repo (use your HPE Passport credentials in ~/.wgetrc if using default)")
    datafabricCmd.Flags().StringP("disk", "d", "/dev/sdb", "MapR data disk")
    // datafabricCmd.Flags().String("host", "", "IP/FQDN to install")
}

func Installer(host, username, password, repo, disk string) {
	// using UAConfig as proxy to pass parameters to template engine
	internal.ProcessTemplate("./templates/df-stanza.yaml", "/tmp/mapr-stanza.yaml", internal.UAConfig{ Master: host, Username: username, Password: password, Domain: disk })
	// internal.SCPPutFile(host, username, password, "/tmp/mapr-stanza.yaml", "/tmp/mapr-stanza.yaml")

	log.Printf("Using repo: %s\n", repo)
	commands := internal.DfInstallerCommands(username, repo)

	// Run the commands
	for _, command := range commands {
		// log.Printf("%s: %s", host, command)
		// err := internal.SSHCommand(host, username, password, command) // Launch the SSH command
		exitCode, err := internal.RunCommand(command) // Run the command
		if err != nil {
			log.Fatal("Error running installer:", err)
		}
		if exitCode > 0 {
			log.Fatal("Installer failed on ", host, exitCode)
		}
	}

}