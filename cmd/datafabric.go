package cmd

import (
	"ezlabctl/internal"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var datafabricCmd = &cobra.Command{
    Use:   "datafabric host",
	Short: "Install Data Fabric, run pre-install configuration if -c flag is set, deploy installer if -i flag is set",
	Args: cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {

		host := args[0]
		dfNode, err := internal.ResolveNode(host)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to validate host: %w", err))
		}
		df := *dfNode

		var username, password, repo string

		if cmd.Flags().Changed("username") {
			username, err = cmd.Flags().GetString("username")
			if err != nil {
				log.Fatal("failed to get username value: %w", err)
			}
		} else {
			username = internal.AskForInput("Username: ", "root")
		}

		if cmd.Flags().Changed("password") {
			password, err = cmd.Flags().GetString("password")
			if err != nil {
				log.Fatal("failed to get password value: %w", err)
			}
		} else {
			password = internal.AskForInput("Password: ", "")
		}

		if cmd.Flags().Changed("repo") {
			repo, err = cmd.Flags().GetString("repo")
			if err != nil {
				log.Fatal("failed to get repo value: %w", err)
			}
		} else {
			repo = internal.AskForInput("MapR Repo: ", "https://package.ezmeral.hpe.com/releases")
		}

		// log.Fatalf("User: %s, pass: %s", username, password)
		if cmd.Flags().Changed("configure") {
			log.Println("Preinstall setup...")
			Preinstall(df.FQDN, username, password, internal.GetCommandOutput("timedatectl show --property=Timezone --value"))
		}

		if cmd.Flags().Changed("installer") {
			log.Println("Deploying mapr-installer...")
			Installer(df.FQDN, username, password, repo)
		}

		commands := []string{
			"[ -f /opt/mapr/conf/mapr-clusters.conf ] || sudo /opt/mapr/installer/bin/mapr-installer-cli install -nvpf -t /tmp/mapr-stanza.yaml -u mapr:mapr@127.0.0.1:9443",
		}
		// Run the commands
		for _, command := range commands {
			// log.Printf("%s: %s", host, command)
			err := internal.SSHCommand(host, username, password, command) // Launch the SSH command
			if err != nil {
				log.Fatal("Error running preinstall configuration:", err)
			}
		}

		log.Printf("EDF deployed on %s\n", host)
    },
}

func init() {

	rootCmd.AddCommand(datafabricCmd)
	datafabricCmd.Flags().BoolP("configure", "c", false, "Run pre-install configuration steps")
    datafabricCmd.Flags().BoolP("installer", "i", false, "Deploy mapr-installer")
    datafabricCmd.Flags().StringP("username", "u", "", "Username")
    datafabricCmd.Flags().StringP("password", "p", "", "Password")
    datafabricCmd.Flags().StringP("repo", "r", "https://package.ezmeral.hpe.com/releases", "MapR Repo (use your HPE Passport credentials in ~/.wgetrc if using default)")
    // datafabricCmd.Flags().String("host", "", "IP/FQDN to install")
}

func Preinstall(host string, username string, password string, timezone string) {
	commands := []string{
		"sudo hostnamectl set-hostname " + host,
		"sudo dnf install -y glibc-langpack-en",
		"sudo localectl set-locale LANG=en_US.UTF-8",
		"sudo timedatectl set-timezone " + timezone,
		"echo '" + username + " ALL=(ALL) NOPASSWD:ALL' | sudo tee /etc/sudoers.d/" + username,
		"sudo sed -i '/^::1/d' /etc/hosts",
		"sudo sed -i '/" + host + "/d' /etc/hosts",
		"sudo sed -i 's/myhostname//g' /etc/nsswitch.conf",
		"echo Updating packages, might take a while...",
		"sudo dnf upgrade -yq",
		"echo preinstall finished for " + host,
	}
	// Run the commands
	for _, command := range commands {
		// log.Printf("%s: %s", host, command)
		err := internal.SSHCommand(host, username, password, command) // Launch the SSH command
		if err != nil {
			log.Fatal("Error running preinstall: ", err)
		}
	}
}

func Installer(host string, username string, password string, repo string) {
	internal.ProcessTemplate("./templates/df-stanza.yaml", "/tmp/mapr-stanza.yaml", internal.UAConfig{ Master: host, Username: username, Password: password })
	internal.SCPPutFile(host, username, password, "/tmp/mapr-stanza.yaml", "/tmp/mapr-stanza.yaml")

	log.Printf("Using repo: %s\n", repo)
	commands := []string{
		"command -v wget || sudo dnf install -yq wget",
		"[ -f /tmp/mapr-setup.sh ] || wget -nv -O /tmp/mapr-setup.sh " + repo + "/installer/redhat/mapr-setup.sh",
		"chmod +x /tmp/mapr-setup.sh",
		"sudo cp /home/" + username + "/.wgetrc /root/.wgetrc || true",
		"[ -f /opt/mapr/installer/bin/mapr-installer-cli ] || ( sudo bash /tmp/mapr-setup.sh -y -r " + repo + " && sleep 120 )",
		"echo Installer ready",
	}

	// Run the commands
	for _, command := range commands {
		// log.Printf("%s: %s", host, command)
		err := internal.SSHCommand(host, username, password, command) // Launch the SSH command
		if err != nil {
			log.Fatal("Error running preinstall configuration:", err)
		}
	}

}