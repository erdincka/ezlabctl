package internal

type UADeployConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Domain string `yaml:"host"`
	RegistryUsername string `yaml:"registryUsername"`
	RegistryPassword string `yaml:"registryPassword"`
	RegistryUrl string `yaml:"registryUrl"`
	RegistryInsecure string `yaml:"registryInsecure"`
	Orchestrator string `yaml:"orchestrator"`
	Master string `yaml:"master"`
	Workers []string `yaml:"workers"`
	ClusterName string `yaml:"clusterName"`
	AuthData string `yaml:"authdata"`
	Proxy string `yaml:"proxy"`
	NoProxy string `yaml:"noProxy"`
}
