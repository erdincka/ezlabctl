#!/bin/bash

go run ezlabctl deploy

# or

kubectl --kubeconfig=/tmp/ez-ua/mgmt-kubeconfig create ns ua

kubectl --kubeconfig=/tmp/ez-ua/mgmt-kubeconfig apply -f /tmp/ez-ua/02-workload-prepare.yaml

kubectl --kubeconfig=/tmp/ez-ua/mgmt-kubeconfig apply -f /tmp/ez-ua/03-workload-deploy.yaml

kubectl --kubeconfig=/tmp/ez-ua/mgmt-kubeconfig apply -f /tmp/ez-ua/04-ezfabric-cluster.yaml

while true; do
    if [[ `go run ezlabctl kubeconf` ]]; then
        break
    fi
    sleep 10
done

kubectl --kubeconfig=/tmp/ez-ua/workload-kubeconfig get nodes --watch


go run ezlabctl status -w
# kubectl --kubeconfig=/tmp/ez-ua/mgmt-kubeconfig get secret ua-kubeconfig -n ua -o json | jq -r '.data.value' | base64 -d > /tmp/ez-ua/workload-kubeconfig

# kubectl --kubeconfig=/tmp/ez-ua/mgmt-kubeconfig get ezfabriccluster/ua -n ua --watch


### DELETE
# kubectl --kubeconfig=/tmp/ez-ua/mgmt-kubeconfig delete -f /tmp/ez-ua/04-ezfabric-cluster.yaml