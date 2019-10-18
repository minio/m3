#!/bin/bash

echo "Provisioning Kind"
kind create cluster --name m3cluster --config kind-cluster.yaml
echo "exporting KUBECONFIG"
export KUBECONFIG="$(kind get kubeconfig-path --name="m3cluster")"
echo "installing dashboard"
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v1.10.1/src/deploy/recommended/kubernetes-dashboard.yaml
echo "creating service account"
kubectl create serviceaccount dashboard -n default
kubectl create clusterrolebinding dashboard-admin -n default --clusterrole=cluster-admin --serviceaccount=default:dashboard

echo "done"

# uncomment this if you want to use the static files
#kubectl apply -f deployments/v3
