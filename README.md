# Ezlab

## Host requirements

- Single-host Data Fabric (optional): 16 vCores, 64GB memory, 1x 100GB data disk
- One Orchestrator host: deployment is initiated here, 4 vCPUs, 32GB memory, 1x 500GB data disk
- One Controller host: Controlplane for Workload cluster, 4 vCPUs, 32GB memory, 1x 500GB data disk
- Three Worker hosts: CPU worker nodes, 32 vCPUs, 128GB memory, 2x 500GB data disks

<!-- ## Prerequisites - tool will take care of these

- Enable password authtentication for all hosts
```bash
sudo sed -i 's/^[^#]*PasswordAuthentication[[:space:]]no/PasswordAuthentication yes/' /etc/ssh/sshd_config
```
- Enable root login for all hosts
```bash
sudo sed -i's/^[^#]*PermitRootLogin[[:space:]]no/PermitRootLogin yes/' /etc/ssh/sshd_config
sudo sed -i's/^[^#]*PermitRootLogin[[:space:]]no/PermitRootLogin yes/' /etc/ssh/sshd_config.d/50-cloud-init.conf
```
- Restart sshd
```bash
sudo systemctl restart sshd
```
- Enable passwordless sudo for all hosts (replace username for your admin user)
```bash
echo "ezmeral ALL=(ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/010_ezlab
sudo chmod 0440 /etc/sudoers.d/010_ezlab
``` -->

### Install single-node Data Fabric:

TODO: Update `.wgetrc` for default repository access.

`ezlabctl df -c -i -u ezmeral -p Admin123. -r http://10.1.1.4/mapr/ -d /dev/sda`

## Deployment

<!-- - Download the latest release from [Github](https://github.com/erdincka/ua-rpm/releases) -->
<!-- - Install the binary to `/usr/local/bin` -->
- Download and install the rpm package `rpm -ivh https://github.com/erdincka/ua-rpm/releases/download/v0.1.2/ezlabctl-0.1.2-1.x86_64.rpm`
- Run `ezlabctl ua -c -s -t -w 10.1.1.33 -u ezmeral -p Admin123.` on the Orchestrator host to setup all hosts for readiness
```ini
[Parameters]
-c=configure host(s)
-s=configure storage
-t=update templates
-u=ssh user
-p=ssh password
-w=workers
```
- Run `ezlabctl storage` on the Orchestrator host to generate the necessary configuration files from Data Fabric
NOTE: S3 IAM Policy needs to be applied to EDF.
- Run `ezlabctl deploy` on the Orchestrator host to deploy the management and then the workload clusters

## Monitor

- Run `ezlabctl status` on the Orchestrator host to check the cluster status
- Run `ezlabctl kubeconf` on the Orchestrator host to get the Kubeconfig file for the workload cluster
- Run `ezlabctl ui` on the Orchestrator host to open the UI

## YAML Templates

https://github.hpe.com/hpe/fabricctl/blob/feature/fy24-q4/example/all-full-input.yaml
