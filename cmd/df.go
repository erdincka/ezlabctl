package cmd

import (
	"ezlabctl/internal"
	"fmt"
	"log"
	"sync"

	"github.com/spf13/cobra"
)

var repo, disk string

func init() {
	rootCmd.AddCommand(datafabricCmd)
	datafabricCmd.Flags().BoolP("configure", "c", false, "Run pre-install configuration steps")
    datafabricCmd.Flags().BoolP("installer", "i", false, "Deploy mapr-installer")
    datafabricCmd.Flags().StringVarP(&sshuser, "username", "u", "", "SSH User")
    datafabricCmd.Flags().StringVarP(&sshpass,"password", "p", "", "SSH Password")
    datafabricCmd.Flags().StringVarP(&repo, "repo", "r", "https://package.ezmeral.hpe.com/releases", "MapR Repo (use your HPE Passport credentials in ~/.wgetrc if using default)")
    datafabricCmd.Flags().StringVarP(&disk, "disk", "d", "/dev/sdb", "MapR data disk")
    // datafabricCmd.Flags().String("host", "", "IP/FQDN to install")
}

var datafabricCmd = &cobra.Command{
    Use:   "df",
	Short: "Install Data Fabric, runs pre-install configuration if -c flag is set, deploys installer if -i flag is set",
    Run: func(cmd *cobra.Command, args []string) {

		var err error = nil
		host := internal.GetOutboundIP()

		// Check root privileges
		internal.IfRoot("Setting up host for Data Fabric installation")

		dfNode, err := internal.ResolveNode(host)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to validate host: %w", err))
		}
		df := *dfNode

		// if cmd.Flags().Changed("configure") {
		if cmd.Flags().Changed("configure") {
			log.Println("Preinstall setup...")

			// Commands required only for DF
			extraCommands := []string {}
			wg := sync.WaitGroup{}
			wg.Add(1)
			internal.Preinstall(df.FQDN, extraCommands, &wg)
			wg.Wait()
		}

		if cmd.Flags().Changed("installer") {
			// Need credentials and repo if installer is set
			sshuser = internal.GetStringInput(cmd, "username", "Username", "root")
			sshpass = internal.GetStringInput(cmd, "password", "Password", "")
			disk = internal.GetStringInput(cmd, "disk", "Data Disk", "/dev/sdb")
			repo = internal.GetStringInput(cmd, "repo", "MapR Repo", "https://package.ezmeral.hpe.com/releases")
			// TODO: check if repo is accessible and ask for credentials if auth needed

			log.Println("Deploying mapr-installer...")
			Installer(df.FQDN, sshuser, sshpass, repo, disk)
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