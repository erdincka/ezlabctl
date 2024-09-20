package internal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Node holds the FQDN, IP address
type Node struct {
	FQDN    string `json:"fqdn"`
	IP      string `json:"ip"`
}

// AppConfig holds the controller and worker node details and common credentials
type AppConfig struct {
	Controller Node   `json:"controller"`
	Workers    []Node `json:"workers"`
	Username string `json:"username"`
	Password string `json:"password"`
    Domain string `json:"domain"`
    Timezone string `json:"timezone"`
	DFHost string `json:"dfhost"`
    DFUser string `json:"dfuser"`
    DFPass string `json:"dfpass"`
	// Repository string `json:"repository"`
}

// GetUAInput collects node details, credentials, and checks connectivity
func GetUAInput() (*AppConfig, error) {
	var controller Node
	var workers []Node

    // Read config file
    appConfig := loadConfig()

	// Get controller node input
    controllerInput := AskForInput("Enter the controller node (IP or Hostname)", appConfig.Controller.IP)

	// Get worker nodes input and validate connectivity for each
    workerIps := strings.Join(func(nodes []Node) []string {
        ips := []string{}
        for _, node := range nodes {
            ips = append(ips, node.IP)
        }
        return ips
    }(appConfig.Workers), ",")

    workerInput := AskForInput("Enter worker nodes (IP or Hostname, min 3, separated by comma)", workerIps)

    if len(strings.Split(workerInput, ",")) < 3 {
        log.Fatalln("minimum of three workers are required")
    }

    // Get credentials
    username := AskForInput("Enter SSH username", appConfig.Username)
    password := AskForInput("Enter SSH password", appConfig.Password)

    // Validate controller and worker nodes
    ctrlNode, err := resolveNode(controllerInput)
    if err != nil {
        return nil, fmt.Errorf("failed to validate controller node: %w", err)
    }
    controller = *ctrlNode

    for _, node := range strings.Split(workerInput, ",") {
        workerNode, err := resolveNode(node)
        if err != nil {
            return nil, fmt.Errorf("invalid worker node #%s: %w", node, err)
        }
        workers = append(workers, *workerNode)
    }

    // Get domain
    domain := AskForInput("Enter UA domain", appConfig.Domain)

	var wg sync.WaitGroup // Create a WaitGroup
    // Test connection to all nodes
    for _, node := range append(workers, controller) {
        wg.Add(1)
        go testCredentials(&node, &username, &password, &wg)
    }
    wg.Wait()

	commandChannel := make(chan string)
    go ExecCommand("timedatectl show --property=Timezone --value", commandChannel)
	// Wait for command result
	tz := <-commandChannel

    // Save the configuration if all went well
	appConfig.Controller = controller
	appConfig.Workers = workers
    appConfig.Username = username
    appConfig.Password = password
    appConfig.Domain =   domain
    appConfig.Timezone = string(tz)

	// Save the updated config
	if err := saveConfig(appConfig); err != nil {
		fmt.Println("Error saving config:", err)
	} else {
		fmt.Println("Config saved.")
	}

	return appConfig, nil
}


// GetDFInput collects information for External Data Fabric configuration
func GetDFInput() (*AppConfig, error) {
    // Read config file
    appConfig := loadConfig()

	host := AskForInput("DF host", appConfig.DFHost)
	user := AskForInput("DF user", appConfig.DFUser)
	password := AskForInput("DF password", appConfig.DFPass)

	appConfig.DFHost = host
	appConfig.DFUser = user
	appConfig.DFPass = password

	// Save the node config to a file
	if err := saveConfig(appConfig); err != nil {
		return nil, err
	}

	return appConfig, nil

}

func getFqdn(ip string) (string) {
    names, err := net.LookupAddr(ip)
    if err != nil || len(names) == 0 {
        fmt.Printf("Failed to get FQDN for %s", ip)
    }
    // fmt.Printf("Found fqdn for %s: %s", ip, names[0])
    return strings.TrimRight(names[0], ".")

}
// resolveNode validates and resolves the FQDN and IP for the given input
func resolveNode(node string) (*Node, error) {
	var fqdn, ip string

	// Check if it's an IP address
	if net.ParseIP(node) != nil {
		// If it's an IP, resolve it to a hostname
        ip = node
        fqdn = getFqdn(ip)
		// fmt.Printf("Resolved IP to hostname: %s\n", fqdn)
	} else {
		// If it's a hostname, resolve it to an IP
		ips, err := net.LookupIP(node)
		if err != nil || len(ips) == 0 {
			return nil, errors.New("hostname does not resolve to an IP address")
		}

        ip = ips[0].String()
        fqdn = getFqdn(ip)

		// fmt.Printf("Resolved hostname to IP: %s\n", ip)
	}

	return &Node{
		FQDN: fqdn,
		IP:   ip,
	}, nil
}

// testCredentials tests SSH connectivity
func testCredentials(node *Node, username *string, password *string, wg *sync.WaitGroup) error {
    defer wg.Done()

    for {
		// Test SSH connection and sudo access
		err := testSSHAndSudo(node, *username, *password)
		if err != nil {
			log.Fatalf("Connection failed: %v. Please re-enter the credentials.\n", err)
		} else {
			log.Printf("Connection to %s successful and passwordless sudo validated!\n", node.FQDN)
			break
		}
	}
	return nil
}

// testSSHAndSudo checks if the node can be accessed via SSH and sudo
func testSSHAndSudo(node *Node, username string, password string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(node.IP, "22"), config)
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %w", err)
	}
	defer client.Close()

	// Run a command to test sudo access
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	cmd := "sudo -n true"
	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("sudo failed, passwordless sudo not enabled for %s: %w", node.IP, err)
	}

	return nil
}

// saveConfig saves the node configuration to a JSON file
func saveConfig(config *AppConfig) error {
	file, err := os.Create("ezlab.json")
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write config to file: %w", err)
	}

	// fmt.Println("Node configuration saved to ezlab_config.json")
	return nil
}

func loadConfig() (*AppConfig) {
    var config AppConfig
	// Attempt to open the config file
	file, err := os.Open("ezlab.json")
	if err != nil {
		fmt.Printf("Warning: failed to open config file: %v, using defaults\n", err)
		return &config // Return default struct if file can't be opened
	}
	defer file.Close()
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&config); err != nil {
		fmt.Printf("Warning: failed to parse config file: %v, using defaults\n", err)
		return &config // Return default struct if file can't be parsed
    }
    return &config
}

// AskForInput prompts the user for input, provides a default value, and returns the user's input or the default value.
func AskForInput(prompt string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Display the prompt with the default value
		if defaultValue != "" {
			fmt.Printf("%s [%s]: ", prompt, defaultValue)
		} else {
			fmt.Printf("%s: ", prompt)
		}

		// Read user input
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response) // Remove any trailing newlines or spaces

		// If response is empty and we have a default value, use the default
		if response == "" && defaultValue != "" {
			return defaultValue
		}

		// If response is not empty, return the user input
		if response != "" {
			return response
		}

		// If response is empty and there's no default value, ask the question again
		fmt.Println("Input cannot be empty. Please provide a valid response.")
	}
}

// https://stackoverflow.com/a/18159705/7033031
func ExecCommand(command string, result chan<- string) {
	log.Println("Executing command:", command)
    args := strings.Split(command, " ")
    cmd := exec.Command(args[0], args[1:]...)
    var out bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &stderr
    err := cmd.Run()
    if err != nil {
        log.Fatal(fmt.Sprint(err) + ": " + stderr.String())
    }
	// log.Println(command, "finished", out.String())
    result <- strings.TrimSpace(out.String())
}