#!/bin/bash

# Install MapR
go run ezlabctl datafabric -u ezmeral -p Admin123. -i -r http://10.1.1.4/mapr 10.1.1.11

# Install UA
go run ezlabctl prepare

go run ezlabctl storage -s 10.1.1.11

go run ezlabctl deploy -p -i -t # including prechecks & orch init

# Following can be used instead of deploy

# kubectl --kubeconfig=/tmp/ez-ua/mgmt-kubeconfig create ns ua

# kubectl --kubeconfig=/tmp/ez-ua/mgmt-kubeconfig apply -f /tmp/ez-ua/02-workload-prepare.yaml

# kubectl --kubeconfig=/tmp/ez-ua/mgmt-kubeconfig apply -f /tmp/ez-ua/03-workload-deploy.yaml

# kubectl --kubeconfig=/tmp/ez-ua/mgmt-kubeconfig apply -f /tmp/ez-ua/04-ezfabric-cluster.yaml

# while true; do
#     if [[ `go run ezlabctl kubeconf` ]]; then
#         break
#     fi
#     sleep 10
# done

# kubectl --kubeconfig=/tmp/ez-ua/workload-kubeconfig get nodes --watch


# go run ezlabctl status -w

# kubectl --kubeconfig=/tmp/ez-ua/mgmt-kubeconfig get secret ua-kubeconfig -n ua -o json | jq -r '.data.value' | base64 -d > /tmp/ez-ua/workload-kubeconfig

# kubectl --kubeconfig=/tmp/ez-ua/mgmt-kubeconfig get ezfabriccluster/ua -n ua --watch


### DELETE
# kubectl --kubeconfig=/tmp/ez-ua/mgmt-kubeconfig delete -f /tmp/ez-ua/04-ezfabric-cluster.yaml