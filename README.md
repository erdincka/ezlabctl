# Ezlab

## Host requirements

- Single-host Data Fabric (optional): 16 vCores, 64GB memory, 1x 100GB data disk
- One Orchestrator host: deployment is initiated here, 8 vCPUs, 32GB memory, 1x 500GB data disk
- One Controller host: Controlplane for Workload cluster, 8 vCPUs, 32GB memory, 1x 500GB data disk
- Three Worker hosts: CPU worker nodes, 32 vCPUs, 128GB memory, 2x 500GB data disks

## Prerequisites

Enable password authtentication on all hosts
```bash
sudo sed -i 's/^[^#]*PasswordAuthentication[[:space:]]no/PasswordAuthentication yes/' /etc/ssh/sshd_config; sudo systemctl restart sshd
```

Enable passwordless sudo for all hosts (replace username for your admin user)
```bash
echo "ezmeral ALL=(ALL) NOPASSWD: ALL" | sudo tee /etc/sudoers.d/010_ezlab
sudo chmod 0600 /etc/sudoers.d/010_ezlab
```

### Install the latest release
<!-- - Download the latest release from [Github](https://github.com/erdincka/ua-rpm/releases) -->
<!-- - Install the binary to `/usr/local/bin` -->
Download and install the rpm package on the Orchestrator node

`rpm -ivh https://github.com/erdincka/ua-rpm/releases/download/v0.1.2/ezlabctl-0.1.2-1.x86_64.rpm`


### Install A Single-node Ezmeral Data Fabric if needed

TODO: Update `.wgetrc` for default repository access.

`ezlabctl df -c -i -u ezmeral -p Admin123. -r http://10.1.1.4/mapr/ -d /dev/sda`

Parameters:

`df` - subcommand for Data Fabric

`--configure` | `-c`: Configure host for installation readiness

`--installer` | `-i`: Install MapR Installer

`--username` | `-u`: SSH username to connect to host for installation

`--password` | `-p`: SSH password for host

`--repo` | `-r`: Repository URL for Data Fabric installer (default: https://package.ezmeral.hpe.com/releases, requires auth credentials to be set in ~/.wgetrc file)

`--disk` | `-d`: Data disk for installation (default: /dev/sdb)

## Ezmeral Unified Analytics Installation

Run on the Orchestrator host

`ezlabctl ua -c -s -t -m 10.1.1.33 -w 10.1.1.34,10.1.1.35,10.1.1.36 -u ezmeral -p Admin123.`


Parameters:

`ua` - subcommand for Unified Analytics

`--configure` | `-c`: Configure node(s) for deployment (use `--master` and `--worker` to configure remote hosts)

`--attach` | `-a`: Configure and attach to MapR for UA with:

  - `--dfhost`: MapR host

  - `--dfuser`: MapR admin user

  - `--dfpass`: MapR admin password

    **NOTE: S3 IAM Policy needs to be applied to EDF.**

`--template` | `-t`: (re-)write templates for deployment

  - `domain`, `master` and `worker` are required

`--validate` | `-v`: Validate node readiness (prechecks)

  - `--master` and `--worker` are required

`--orchinit` | `-o`: Initialize orchestrator cluster on this node

  - TODO: required params

`--confirm`: Confirm workload cluster deployment

  - TODO: required params

`--username` | `-u`: SSH username to connect to hosts for installation (default: ezmeral)

`--password` | `-p` : SSH password for hosts

`--master` | `-m`: Master node for workload cluster

  - `--username` and `--password` are required

`--worker` | `-w`: Worker node for workload cluster (minimum of 3)

  - `--username` and `--password` are required

`--domain` | `-d`: Domain name for workload cluster

`--registryUrl`: Airgap registry for UA

`--registryUsername`: Registry username

`--registryPassword`: Registry password

`--registryCa`: Registry CA pem in base64

`--registryInsecure`: Set registry to insecure for http access (default: true)


## Monitor

- Run `ezlabctl status` on the Orchestrator host to check the cluster status
- Run `ezlabctl kubeconf` on the Orchestrator host to get the Kubeconfig file for the workload cluster
- Run `ezlabctl ui` on the Orchestrator host to open the UI
