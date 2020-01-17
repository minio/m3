#!/bin/bash

echo "Provisioning Kind"
kind create cluster --name m3cluster --config kind-cluster.yaml
echo "installing dashboard"
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.0.0-rc1/aio/deploy/recommended.yaml
echo "creating service account"
kubectl create serviceaccount dashboard -n default
kubectl create clusterrolebinding dashboard-admin -n default --clusterrole=cluster-admin --serviceaccount=default:dashboard
# pre-load MinIO, postgres, etcd, etcd-operator to speed up setup
docker pull minio/minio:RELEASE.2020-01-16T22-40-29Z
docker pull postgres:12
docker pull quay.io/coreos/etcd-operator:v0.9.4
docker pull quay.io/coreos/etcd:v3.4.0
docker pull quay.io/prometheus/prometheus:v2.14.0
kind load docker-image minio/minio:RELEASE.2020-01-16T22-40-29Z --name m3cluster
kind load docker-image postgres:12 --name m3cluster
kind load docker-image quay.io/coreos/etcd-operator:v0.9.4 --name m3cluster
kind load docker-image quay.io/coreos/etcd:v3.4.0 --name m3cluster
kind load docker-image quay.io/prometheus/prometheus:v2.14.0 --name m3cluster

make --directory="../portal-ui" k8sdev
make --directory="../nginx" k8sdev
make --directory="../vault" k8sdev

kubectl apply -f deployments/m3-portal-backend-deployment.yaml
kubectl apply -f deployments/m3-portal-frontend-deployment.yaml
kubectl apply -f deployments/m3-vault-deployment.yaml
kubectl apply -f deployments/portal-proxy-deployment.yaml

# Build mkube
make --directory=".." k8sdev TAG=minio/m3:dev

# Apply mkube
kubectl apply -f deployments/m3-deployment.yaml

# Extract and save vault token into config map
while [[ $(kubectl get pods -l app=m3-vault -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "waiting for pod" && sleep 1; done

sleep 5
VAULT_TOKEN=$(kubectl logs $(kubectl get pods | grep vault | awk '{print $1}') | grep token:| sed 's/^.*: //')
kubectl get configmaps m3-env -o json | jq --arg vt "$VAULT_TOKEN" '.data["KMS_TOKEN"]=$vt' | kubectl apply -f -

echo "done"
