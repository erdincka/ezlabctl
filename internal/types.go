package internal

// Node holds the FQDN, IP address
type Node struct {
	FQDN    string `json:"fqdn"`
	IP      string `json:"ip"`
}

// AppConfig holds the application settings and common credentials
type AppConfig struct {
	Orchestrator Node `json:"orchestrator"`
	Controller Node   `json:"controller"`
	Workers    []Node `json:"workers"`
	Username string `json:"username"`
	Password string `json:"password"`
    Domain string `json:"domain"`
    Timezone string `json:"timezone"`
	RegistryUrl string `json:"registryurl"`
	RegistryUsername string `json:"registryusername"`
	RegistryPassword string `json:"registrypassword"`
	RegistryInsecure bool   `json:"registryinsecure"`
	DFHost string `json:"dfhost"`
    DFAdmin string `json:"dfuser"`
    DFPass string `json:"dfpass"`
}

type DFConfig struct {
	CldbNodes string `json:"cldbnodes"`
	RestNodes string `json:"restnodes"`
	S3Nodes string `json:"s3nodes"`
	AccessKey string `json:"s3accesskey"`
	SecretKey string `json:"s3secretkey"`
	TenantTicket string `json:"tenantticket"`
	ClusterName string `json:"clustername"`
}

// UADeployConfig holds the controller and worker node details and common credentials
type UADeployConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Domain string `yaml:"host"`
	RegistryUrl string `yaml:"registryUrl"`
	RegistryInsecure bool `yaml:"registryInsecure"`
	RegistryUsername string `yaml:"registryUsername"`
	RegistryPassword string `yaml:"registryPassword"`
	RegistryCa string `yaml:"registryCa"`
	Orchestrator string `yaml:"orchestrator"`
	Master string `yaml:"master"`
	Workers []string `yaml:"workers"`
	ClusterName string `yaml:"clusterName"`
	AuthData string `yaml:"authdata"`
	Proxy string `yaml:"proxy"`
	NoProxy string `yaml:"noProxy"`
	DF DFConfig `yaml:"df"`
}
