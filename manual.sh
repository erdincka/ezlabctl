#!/bin/bash

kubectl --kubeconfig=/tmp/ez-ua15/mgmt-kubeconfig create ns ua15

kubectl --kubeconfig=/tmp/ez-ua15/mgmt-kubeconfig apply -f /tmp/ez-ua15/02-workload-prepare.yaml

kubectl --kubeconfig=/tmp/ez-ua15/mgmt-kubeconfig apply -f /tmp/ez-ua15/03-workload-deploy.yaml

kubectl --kubeconfig=/tmp/ez-ua15/mgmt-kubeconfig apply -f /tmp/ez-ua15/04-ezfabric-cluster.yaml

kubectl --kubeconfig=/tmp/ez-ua15/mgmt-kubeconfig get secret ua15-kubeconfig -n ua15 -o json | jq -r '.data.value' | base64 -d > /tmp/ez-ua15/workload-kubeconfig

kubectl --kubeconfig=/tmp/ez-ua15/mgmt-kubeconfig get ezfabriccluster/ua15 -n ua15 --watch