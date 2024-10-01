package internal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
)

var deploySteps = map[string]string{
	"prechecks":        "00-prechecks.yaml",
	"fabricinit":       "01-fabricctl-init.yaml",
	"workloadprepare":  "02-workload-prepare.yaml",
	"workloaddeploy":   "03-workload-deploy.yaml",
	"fabriccluster":    "04-ezfabric-cluster.yaml",
}

func GetDeploySteps() (map[string]string) {
	return deploySteps
}

func GetDeployConfig() (string, TemplateFiles, UADeployConfig) {
	appConfig := GetAppConfiguration()

	clusterName := strings.Split(appConfig.Domain, ".")[0]

	templateFiles := TemplateFiles{
		TemplateDirectory: "/tmp/ez-" + clusterName,
		OrchestratorKubeConfig: "/tmp/ez-" + clusterName + "/mgmt-kubeconfig",
		WorkloadKubeConfig: "/tmp/ez-" + clusterName + "/workload-kubeconfig",
	}

	// Setup auth data for admin user
	authData := map[string]interface{}{
		"admin_user": map[string]string{
			"fullname": "Ezmeral Admin",
			"email":    fmt.Sprintf("ezadmin@%s", appConfig.Domain),
			"username": "ezua",
			"password": appConfig.Password,
		},
	}
	// Convert authData to JSON
	authDataJSON, err := json.Marshal(authData)
	if err != nil {
		fmt.Println("Error converting authData to JSON:", err)
		return "", TemplateFiles{}, UADeployConfig{}
	}
	// log.Println("Auth data: " + string(authDataJSON))

	dfConfig := GetMaprConfig()
	// log.Printf("DEBUG tenantticket: %v", dfConfig.TenantTicket)

	uaConfig := UADeployConfig{
		Username: appConfig.Username,
		Password: appConfig.Password,
		// Password: base64.StdEncoding.EncodeToString([]byte(appConfig.Password)),
		Domain: appConfig.Domain,
		RegistryUrl: appConfig.RegistryUrl,
		RegistryInsecure: appConfig.RegistryInsecure,
		RegistryUsername: appConfig.RegistryUsername,
		RegistryPassword: appConfig.RegistryPassword,
		RegistryCa: "",
		// Orchestrator: GetOutboundIP(),
		Orchestrator: appConfig.Orchestrator.IP,
		Master: appConfig.Controller.IP,
		Workers: strings.Split(GetWorkerIPs(), ","),
		ClusterName: clusterName,
		AuthData: base64.StdEncoding.EncodeToString(authDataJSON),
		NoProxy: "10.96.0.0/12,10.224.0.0/16,10.43.0.0/16,192.168.0.0/16,.external.hpe.local,localhost,.cluster.local,.svc,.default.svc,127.0.0.1,169.254.169.254," + GetWorkerIPs() + "," + appConfig.Controller.IP + "," + appConfig.Orchestrator.IP + ",." + appConfig.Domain,
		DF: dfConfig,
	}

	return clusterName, templateFiles, uaConfig
}

func GetMaprConfig() DFConfig {
	// Define a regex pattern to match IPv4 addresses
	ipv4Pattern := `(\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b)`
	// Define regex patterns for accesskey and secretkey
	accessKeyPattern := `"accesskey":"([A-Z0-9]+)"`
	secretKeyPattern := `"secretkey":"([A-Z0-9]+)"`

	// Compile the regular expressions
	ipV4re, err := regexp.Compile(ipv4Pattern)
	if err != nil {
		log.Fatalf("failed to compile regex: %v", err)
	}
	accessKeyRe, err := regexp.Compile(accessKeyPattern)
	if err != nil {
		log.Fatalf("failed to compile accesskey regex: %v", err)
	}
	secretKeyRe, err := regexp.Compile(secretKeyPattern)
	if err != nil {
		log.Fatalf("failed to compile secretkey regex: %v", err)
	}

	dfConfig := DFConfig{}

	fileContent := ReadFile("/tmp/cldb_nodes.json")
	dfConfig.CldbNodes = ipV4re.FindString(string(fileContent))

	fileContent = ReadFile("/tmp/rest_nodes.json")
	dfConfig.RestNodes = ipV4re.FindString(string(fileContent))

	fileContent = ReadFile("/tmp/s3_nodes.json")
	dfConfig.S3Nodes = ipV4re.FindString(string(fileContent))

	fileContent = ReadFile("/tmp/s3_keys.json")
	accessKeyMatch := accessKeyRe.FindStringSubmatch(string(fileContent))
	if len(accessKeyMatch) < 2 {
		log.Fatalf("accesskey not found")
	}
	dfConfig.AccessKey = base64.StdEncoding.EncodeToString([]byte(accessKeyMatch[1]))

	secretKeyMatch := secretKeyRe.FindStringSubmatch(string(fileContent))
	if len(secretKeyMatch) < 2 {
		log.Fatalf("secretkey not found")
	}
	dfConfig.SecretKey = base64.StdEncoding.EncodeToString([]byte(secretKeyMatch[1]))

	fileContent = ReadFile("/tmp/maprtenantticket")
	dfConfig.ClusterName = strings.Split(string(fileContent), " ")[0]
	dfConfig.TenantTicket = base64.StdEncoding.EncodeToString([]byte(fileContent))
	log.Printf("Using Data Fabric Cluster: %s\n", dfConfig.ClusterName)

	return dfConfig
}