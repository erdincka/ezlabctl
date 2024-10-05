package internal

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// GetUAInput collects node details, credentials, and checks connectivity
func GetUAInput() (*AppConfig, error) {
	var orchestrator Node
	var controller Node
	var workers []Node

    // Read config file
    appConfig := GetAppConfiguration()

	// Get orchestrator node input (default to localhost)
    orchestratorInput := AskForInput("Enter the orchestrator node (IP or Hostname)", appConfig.Orchestrator.IP)

	// Get controller node input
    controllerInput := AskForInput("Enter the controller node (IP or Hostname)", appConfig.Controller.IP)

	// Get worker nodes input and validate connectivity for each
    workerIps := GetWorkerIPs()
    workerInput := AskForInput("Enter worker nodes (IP or Hostname, min 3, separated by comma)", workerIps)

    if len(strings.Split(workerInput, ",")) < 3 {
        log.Fatalln("minimum of three workers are required")
    }

    // Get credentials
    username := AskForInput("Enter SSH username", appConfig.Username)
    password := AskForInput("Enter SSH password", appConfig.Password)

    // Validate orchestrator, controller and worker nodes
    orchNode, err := ResolveNode(orchestratorInput)
    if err != nil {
        return nil, fmt.Errorf("failed to validate orchestrator node: %w", err)
    }
    orchestrator = *orchNode

    ctrlNode, err := ResolveNode(controllerInput)
    if err != nil {
        return nil, fmt.Errorf("failed to validate controller node: %w", err)
    }
    controller = *ctrlNode

    for _, node := range strings.Split(workerInput, ",") {
        workerNode, err := ResolveNode(node)
        if err != nil {
            return nil, fmt.Errorf("invalid worker node #%s: %w", node, err)
        }
        workers = append(workers, *workerNode)
    }

    // Get domain
    domain := AskForInput("Enter UA domain", appConfig.Domain)

    // Test connection to all nodes
    for _, node := range append(workers, controller, orchestrator) {
        go TestCredentials(node.IP, &username, &password)
    }

	registryUrl := AskForInput("Enter Registry URL (without http[s])", appConfig.RegistryUrl)
	registryUsername := AskForInput("Enter Registry Username", appConfig.RegistryUsername)
	registryPassword := AskForInput("Enter Registry Password", appConfig.RegistryPassword)
	registryInsecure := AskForInput("Is Registry Insecure", fmt.Sprint(appConfig.RegistryInsecure))

    // Save the configuration if all went well
	appConfig.Orchestrator = orchestrator
	appConfig.Controller = controller
	appConfig.Workers = workers
    appConfig.Username = username
    appConfig.Password = password
    appConfig.Domain =   domain
	appConfig.RegistryUrl = registryUrl
	appConfig.RegistryUsername = registryUsername
	appConfig.RegistryPassword = registryPassword
	appConfig.RegistryInsecure = registryInsecure == "true"

	// Save the updated config
	if err := saveConfig(appConfig); err != nil {
		log.Fatal("Error saving config:", err)
	} else {
		log.Println("Config saved.")
	}

	return appConfig, nil
}

func GetWorkerIPs() string {
	appConfig := GetAppConfiguration()
	return strings.Join(func(nodes []Node) []string {
        ips := []string{}
        for _, node := range nodes {
            ips = append(ips, node.IP)
        }
        return ips
    }(appConfig.Workers), ",")
}

// GetDFInput collects information for External Data Fabric configuration
func GetDFInput() (*AppConfig, error) {
    // Read config file
    appConfig := GetAppConfiguration()

	host := AskForInput("DF host", appConfig.DFHost)
	user := AskForInput("DF Cluster Admin", appConfig.DFAdmin)
	password := AskForInput("DF Cluster Admin password", appConfig.DFPass)

	appConfig.DFHost = host
	appConfig.DFAdmin = user
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
        log.Fatalf("Failed to get FQDN for %s\n", ip)
    }
	// log.Printf("Found FQDNs for %s: %s", ip, names)
	return strings.TrimRight(names[0], ".")

}
// ResolveNode validates and resolves the FQDN and IP for the given input
func ResolveNode(node string) (*Node, error) {
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

func GetAppConfiguration() *AppConfig {
    var config AppConfig
	// Attempt to open the config file
	file, err := os.Open("ezlab.json")
	if err != nil {
		fmt.Println("Using defaults")
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

	// for {
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
		// if response != "" {
			return response
		// }

		// If response is empty and there's no default value, ask the question again
		// fmt.Println("Input is empty with no default value. Returning empty string.")
	// }
}

// https://stackoverflow.com/a/18159705/7033031
// func ExecCommand(command string, result chan<- string) {
// 	log.Println("Executing command:", command)
//     args := strings.Split(command, " ")
//     cmd := exec.Command(args[0], args[1:]...)
//     var out bytes.Buffer
//     var stderr bytes.Buffer
//     cmd.Stdout = &out
//     cmd.Stderr = &stderr
//     err := cmd.Run()
//     if err != nil {
//         log.Fatal(fmt.Sprint(err) + ": " + stderr.String())
//     }
// 	// log.Println(command, "finished", out.String())
//     result <- strings.TrimSpace(out.String())
// 	close(result)
// }

func ReadFile(path string) []byte {
	// Open the file and defer closing it
	file, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	// Read the contents of the file into a string
	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatalln(err)
	}

	return data
}

func GetStringInput(cmd *cobra.Command, param, prompt, defaultValue string) string {
	var input string
	var err error
	if cmd.Flags().Changed(param) {
		input, err = cmd.Flags().GetString(param)
		if err != nil {
			log.Fatal("failed to get host value: %w", err)
		}
	} else {
		input = AskForInput(prompt, defaultValue)
	}
	return input
}