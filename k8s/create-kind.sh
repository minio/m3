#!/bin/bash

echo "Provisioning Kind"
kind create cluster --name m3cluster --config kind-cluster.yaml
echo "installing dashboard"
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.0.0-beta6/aio/deploy/recommended.yaml
echo "creating service account"
kubectl create serviceaccount dashboard -n default
kubectl create clusterrolebinding dashboard-admin -n default --clusterrole=cluster-admin --serviceaccount=default:dashboard
# pre-load MinIO, postgres, etcd, etcd-operator to speed up setup
docker pull minio/minio:RELEASE.2019-10-12T01-39-57Z
docker pull postgres:12
docker pull quay.io/coreos/etcd-operator:v0.9.4
docker pull quay.io/coreos/etcd:v3.4.0
kind load docker-image minio/minio:RELEASE.2019-10-12T01-39-57Z --name m3cluster
kind load docker-image postgres:12 --name m3cluster
kind load docker-image quay.io/coreos/etcd-operator:v0.9.4 --name m3cluster
kind load docker-image quay.io/coreos/etcd:v3.4.0 --name m3cluster

echo "done"

# uncomment this if you want to use the static files
#kubectl apply -f deployments/v3
