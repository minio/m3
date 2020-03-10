#!/bin/bash


# setup environment variables based on flags to see if we should build the docker containers again
M3_DOCKER="true"
VAULT_DOCKER="true"
PORTAL_DOCKER="true"

# evaluate flags
# `-m` for mkube
# `-v` for m3-vault
# `-p` for portal-ui


while getopts ":m:v:p:" opt; do
  case $opt in
    m)
	  M3_DOCKER="$OPTARG"
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
kind create cluster  --config kind-cluster.yaml
#echo "Remove Master Taint"
kubectl taint nodes --all node-role.kubernetes.io/master-
echo "install metrics server"
kubectl apply -f deployments/metrics-dev.yaml
#echo "installing dashboard"
#kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.0.0-rc5/aio/deploy/recommended.yaml
#echo "creating service account"
#kubectl create serviceaccount dashboard -n default
#kubectl create clusterrolebinding dashboard-admin -n default --clusterrole=cluster-admin --serviceaccount=default:dashboard
# pre-load MinIO, postgres to speed up setup
docker pull minio/minio:RELEASE.2020-02-07T23-28-16Z
kind load docker-image minio/minio:RELEASE.2020-02-07T23-28-16Z

# Whether or not to build the portal container and load it to kind or just load it
if [[ $PORTAL_DOCKER == "true" ]]; then
	# Build portal-ui
  make --directory="../portal-ui" k8sdev
else
	kind load docker-image minio/m3-portal-frontend:dev
fi


# Whether or not to build the m3-vault container and load it to kind or just load it
#if [[ $VAULT_DOCKER == "true" ]]; then
#	# Build vault
#  make --directory="../vault" k8sdev
#else
#	kind load docker-image minio/m3-vault:edge
#fi

# Setup development postgres
#kubectl apply -f deployments/m3-vault-deployment.yaml
#kubectl apply -f deployments/portal-proxy-deployment.yaml

# Whether or not to build the m3 container and load it to kind or just load it
if [[ $M3_DOCKER == "true" ]]; then
	# Build mkube
  make --directory=".." k8sdev TAG=minio/m3:dev
else
	kind load docker-image minio/m3:dev
fi


# Apply mkube
kubectl apply -f deployments/m3-deployment.yaml

## Extract and save vault token into config map
#while [[ $(kubectl get pods -l app=m3-vault -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "waiting for pod" && sleep 1; done

#sleep 5
#VAULT_TOKEN=$(kubectl logs $(kubectl get pods | grep vault | awk '{print $1}') | grep token:| sed 's/^.*: //')
#kubectl get configmaps m3-env -o json | jq --arg vt "$VAULT_TOKEN" '.data["KMS_TOKEN"]=$vt' | kubectl apply -f -

echo "done"
