#!/bin/bash


# setup environment variables based on flags to see if we should build the docker containers again
M3_DOCKER="true"
NGINX_DOCKER="true"
VAULT_DOCKER="true"
PORTAL_DOCKER="true"

# evaluate flags
# `-m` for mkube
# `-n` for m3-nginx
# `-v` for m3-vault
# `-p` for portal-ui


while getopts ":m:n:v:p:" opt; do
  case $opt in
    m)
	  M3_DOCKER="$OPTARG"
      ;;
    n)
	  NGINX_DOCKER="$OPTARG"
      ;;
    v)
	  VAULT_DOCKER="$OPTARG"
      ;;
    p)
	  PORTAL_DOCKER="$OPTARG"
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      exit 1
      ;;
    :)
      echo "Option -$OPTARG requires an argument." >&2
      exit 1
      ;;
  esac
done

echo "Provisioning Kind"
kind create cluster --name m3cluster --config kind-cluster.yaml

echo "install metrics server"
kubectl apply -f deployments/metrics-dev.yaml

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

# Whether or not to build the portal container and load it to kind or just load it
if [[ $PORTAL_DOCKER == "true" ]]; then
	# Build portal-ui
  make --directory="../portal-ui" k8sdev
else
	kind load docker-image minio/m3-portal-frontend:dev --name m3cluster
fi

# Whether or not to build the m3-nginx container and load it to kind or just load it
if [[ $NGINX_DOCKER == "true" ]]; then
	# Build nginx
  make --directory="../nginx" k8sdev
else
	kind load docker-image minio/m3-nginx:edge --name m3cluster
fi


# Whether or not to build the m3-vault container and load it to kind or just load it
if [[ $VAULT_DOCKER == "true" ]]; then
	# Build vault
  make --directory="../vault" k8sdev
else
	kind load docker-image minio/m3-vault:edge --name m3cluster
fi

# Whether or not to build the m3 container and load it to kind or just load it
if [[ $M3_DOCKER == "true" ]]; then
	# Build mkube
  make --directory=".." k8sdev TAG=minio/m3:dev
else
	kind load docker-image minio/m3:dev --name m3cluster
fi

# install autocert
sleep 30

kubectl apply -f deployments/m3-autocert-deployment.yaml
kubectl run autocert-init -it --rm --image alevsk/autocert-init:dev --env="AUTO_START=true" --restart Never
kubectl label namespace default autocert.step.sm=enabled

sleep 5

# Setup development postgres
kubectl apply -f deployments/postgres-dev.yaml
# install etcd-operator
kubectl apply -f deployments/etcd-dev.yaml
# install prometheus
kubectl apply -f deployments/prometheus-dev.yaml

kubectl apply -f deployments/m3-portal-backend-deployment.yaml
kubectl apply -f deployments/m3-portal-frontend-deployment.yaml
kubectl apply -f deployments/m3-vault-deployment.yaml
kubectl apply -f deployments/portal-proxy-deployment.yaml

# Wait for etcd-operator custom resource to be ready, then create the etcd cluster
CR_COUNT=0
while [[ $CR_COUNT == 0 ]]; do CR_COUNT=$(kubectl get customresourcedefinitions.apiextensions.k8s.io | grep etcdclusters | wc -l) && echo "Waiting on etcd-operator" &&sleep 1; done
kubectl apply -f deployments/etcd-dev-cr.yaml

# Apply mkube
kubectl apply -f deployments/m3-deployment.yaml

# Extract and save vault token into config map
while [[ $(kubectl get pods -l app=m3-vault -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "waiting for pod" && sleep 1; done

sleep 5
VAULT_TOKEN=$(kubectl logs $(kubectl get pods | grep vault | awk '{print $1}') -c m3-vault | grep token: | sed 's/^.*: //')
kubectl get configmaps m3-env -o json | jq --arg vt "$VAULT_TOKEN" '.data["KMS_TOKEN"]=$vt' | kubectl apply -f -

# Extracts cluster CA certificate (the one used by autocert), you need to install manually this in your system in order to issue m3 commands
kubectl get configmap -n step certs -o json | jq -r '.data["root_ca.crt"]' > ../cluster-ca.crt

echo "done"
