package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
)

func PrepareCommands(hostname string) []string {

	commands := []string{
		// TODO: Better handling for various resolvers (/etc/[hosts|resolv.conf], resolvectl etc)
		// Disable cloud-init if it manages /etc/hosts
		"sudo sed -i '/manage_etc_hosts/d' /etc/cloud/cloud.cfg",
		"sudo sed -i 's/preserve_hostname:.*/preserve_hostname: true\\nmanage_etc_hosts: false/' /etc/cloud/cloud.cfg",
		"sudo sed -i 's/ssh_pwauth:.*/ssh_pwauth: true/' /etc/cloud/cloud.cfg",
		"sudo sed -i 's/^[^#]*PasswordAuthentication[[:space:]]no/PasswordAuthentication yes/' /etc/ssh/sshd_config",
		// following file is in RHEL9 not RHEL8
		// "sudo sed -i 's/^[^#]*PasswordAuthentication[[:space:]]no/PasswordAuthentication yes/' /etc/ssh/sshd_config.d/50-cloud-init.conf || true",
		"sudo systemctl restart sshd",
		// set host resolution for IPv4
		"sudo sed -i '/^::1/d' /etc/hosts",
		// Disable NetworkManager if it manages /etc/hosts or /etc/resolv.conf
		"sudo sed -i '/" + strings.Split(hostname, ".")[0] + "/d' /etc/hosts",
		"sudo sed -i 's/myhostname//g' /etc/nsswitch.conf",
		"sudo hostnamectl set-hostname " + hostname,
		"sudo dnf install -yq glibc-langpack-en",
		"sudo localectl set-locale LANG=en_US.UTF-8",
		// "echo '" + user + " ALL=(ALL) NOPASSWD:ALL' | sudo tee /etc/sudoers.d/" + user,
		// Rocky 9 repository workaround, not needed for RHEL8
		// "sudo rpm -e openssl-fips-provider --nodeps || true",
		"sudo dnf upgrade -yq",
	}

	return commands
}

func Preinstall(hostname string, extraCommands []string, wg *sync.WaitGroup, dryrun bool) {
	defer wg.Done()

	commands := PrepareCommands(hostname)
	for _, command := range append(commands, extraCommands...) {
		if dryrun {
			log.Printf("[ %s ]: Skipped: %s", hostname, command)
		} else {
			// log.Printf("%s: %s", hostname, command)
			exitCode, err := RunCommand(command)
			if err != nil {
				log.Fatal("Error running preinstall: ", err)
			}
			if exitCode > 0 {
				log.Fatal("Pre-install configuration failed on ", hostname, exitCode)
			}
		}
	}
}

func PreinstallOverSsh(hostname, sshuser, sshpass string, wg *sync.WaitGroup, dryrun bool) {
	defer wg.Done()
	commands := PrepareCommands(hostname)
	if dryrun {
		log.Printf("[%s]: Skipped: %v", hostname, strings.Join(commands, "\n"))
	} else {
		err := SshCommands(hostname, sshuser, sshpass, commands)
		if err != nil {
			log.Fatal("Error running preinstall: ", err)
		}
	}
}

func DfSetupForUACommands(dfuser, dfpass string, files []string) []string {
	commands := []string{
        "id ezua || sudo useradd -m -U ezua",
        "echo ezua:" + dfpass + " | sudo chpasswd",
        "[ -f /tmp/maprticket_$(id -u) ] || echo " + dfpass + " | maprlogin password -user " + dfuser,
        "maprcli acl edit -type cluster -user ezua:login,cv",
        "maprcli volume create -name ezua-base-volume-ezua -path /ezua -type rw -json -rootdiruser ezua -rootdirgroup ezua -createparent true || true",
        "[ -f /tmp/ezua-maprtenantticket ] || maprlogin generateticket -type tenant -user ezua -out /tmp/ezua-maprtenantticket",
        "[ -f /tmp/ezua-cldb-nodes.json ] || /opt/mapr/bin/maprcli node list -columns hn -filter svc==cldb -json > /tmp/ezua-cldb-nodes.json",
        "[ -f /tmp/ezua-rest-nodes.json ] || /opt/mapr/bin/maprcli node list -columns hn -filter svc==apiserver -json > /tmp/ezua-rest-nodes.json",
        "[ -f /tmp/ezua-s3-nodes.json ] || /opt/mapr/bin/maprcli node list -columns hn -filter svc==s3server -json > /tmp/ezua-s3-nodes.json",
        "[ -f /tmp/ezua-s3-keys.json ] || maprcli s3keys generate -domainname primary -accountname default -username ezua -json > /tmp/ezua-s3-keys.json",
		"[ -f /tmp/ezua-chain-ca.pem ] || cp /opt/mapr/conf/ca/chain-ca.pem /tmp/ezua-chain-ca.pem",
        "sudo chown mapr:mapr /tmp/" + strings.Join(files, " /tmp/"),
        // "[ -d /mapr ] || sudo mkdir /mapr",
        // "mount | grep -q /mapr || sudo mount -t nfs -o vers=3,nolock " + dfhost + ":/mapr/ /mapr",
    }

	return commands
}
func DfInstallerCommands(username, repo string) []string {
	commands := []string{
		"command -v wget || sudo dnf install -yq wget",
		"[ -f /tmp/mapr-setup.sh ] || wget -nv -O /tmp/mapr-setup.sh " + repo + "/installer/redhat/mapr-setup.sh",
		"chmod +x /tmp/mapr-setup.sh",
		// FIX: ask user for auth credentials if needed
		"sudo cp /home/" + username + "/.wgetrc /root/.wgetrc || true",
		"[ -f /opt/mapr/installer/bin/mapr-installer-cli ] || ( sudo bash /tmp/mapr-setup.sh -y -r " + repo + " && sleep 120 )",
	}
	return commands
}


func GetCommandOutput(command string) string {
	args := strings.Split(command, " ")
	cmd := exec.Command(args[0], args[1:]...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	// Set up stdout and stderr redirection to the respective buffers
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(fmt.Sprint(err) + ": " + stderr.String())
		return ""
	}

	// Just log the stderr if any
	if stderr.Len() > 0 {
		log.Printf("STDERR: %s", stderr.String())
	}
	return strings.TrimSpace(out.String())
}

// runCommand streams the output of a command while it's running and logs the exit code after completion.
func RunCommand(command string) (int, error) {
	// Split the command into name and arguments
	// args := strings.Fields(command)
	// cmdName := args[0]
	// cmdArgs := strings.Join(args[1:], " ")

	// Create the command
	// cmd := exec.Command(cmdName, cmdArgs...)
	log.Println("[LOCAL]:", command)
	cmd := exec.Command("/bin/bash", "-c", command)

	// Get stdout and stderr pipes
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return 0, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return 0, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start command: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	// Stream stdout and stderr concurrently
	// Stream stdout and stderr concurrently
	go func() {
		defer wg.Done()
		streamOutput(stdoutPipe, "STDOUT")
	}()

	go func() {
		defer wg.Done()
		streamOutput(stderrPipe, "STDERR")
	}()

	// Wait for stdout and stderr to be fully read
	wg.Wait()

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Command exited with a non-zero status code
			return exitErr.ExitCode(), fmt.Errorf("command failed with exit code %d", exitErr.ExitCode())
		}
		return 0, fmt.Errorf("command failed to complete: %w", err)
	}

	// Command finished successfully
	return cmd.ProcessState.ExitCode(), nil
}

// streamOutput reads and logs the output from the provided io.Reader.
func streamOutput(pipe io.ReadCloser, pipeName string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		// fmt.Printf("[%s] %s\n", pipeName, scanner.Text())
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading %s: %v\n", pipeName, err)
	}
}

// func IfRoot(prompt string) {
// 	if os.Geteuid() == 0 {
// 		log.Println(prompt)
// 	} else {
// 		log.Fatalf("You must be root to run this command. Now using UID: %d", os.Geteuid())
// 	}
// }