# Ezlab

## Host requirements

- One Orchestrator host: where you run the `ezlabctl` command
- One Controller host: Controlplane for Workload cluster
- Three Worker hosts: CPU worker nodes

## Prerequisites

- Enable password authtentication for all hosts
```bash
sudo sed -i 's/^[^#]*PasswordAuthentication[[:space:]]no/PasswordAuthentication yes/' /etc/ssh/sshd_config
```
- Enable root login for all hosts
```bash
sudo sed -i's/^[^#]*PermitRootLogin[[:space:]]no/PermitRootLogin yes/' /etc/ssh/sshd_config
```
- Restart sshd
```bash
sudo systemctl restart sshd
```
- Enable passwordless sudo for all hosts (replace username for your admin user)
```bash
echo "ezmeral ALL=(ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/010_ezlab
sudo chmod 0440 /etc/sudoers.d/010_ezlab
```

## Deployment

- Download the latest release from [Github](https://github.com/erdincka/ua-rpm/releases)
- Install the binary to `/usr/local/bin`
- Run `ezlabctl prepare` on the Orchestrator host to setup all hosts for readiness
- Run `ezlabctl setup` on the Orchestrator host to generate the necessary configuration files from Data Fabric
- Run `ezlabctl deploy` on the Orchestrator host to deploy the management and then the workload clusters

## Monitor

- Run `ezlabctl status` on the Orchestrator host to check the cluster status
- Run `ezlabctl kubeconf` on the Orchestrator host to get the Kubeconfig file for the workload cluster
- Run `ezlabctl ui` on the Orchestrator host to open the UI

## YAML Templates

https://github.hpe.com/hpe/fabricctl/blob/feature/fy24-q4/example/all-full-input.yaml
