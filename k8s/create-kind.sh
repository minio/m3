#!/bin/bash

echo "Provisioning Kind"
kind create cluster --name m3cluster --image m3kind --config kind-cluster.yaml 
echo "exporting KUBECONFIG"
export KUBECONFIG="$(kind get kubeconfig-path --name="m3cluster")"
echo "installing dashboard"
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v1.10.1/src/deploy/recommended/kubernetes-dashboard.yaml 
echo "creating service account"
kubectl create serviceaccount dashboard -n default
kubectl create clusterrolebinding dashboard-admin -n default --clusterrole=cluster-admin --serviceaccount=default:dashboard

echo "done"

echo "Creating: tenant-1"
./add-volume-tenant.sh tenant-1
echo "Creating: tenant-2"
./add-volume-tenant.sh tenant-2
kubectl apply -f deployments/v3
