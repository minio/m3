#!/bin/bash

# setup environment variables based on flags to see if we should build the docker containers again
M3_DOCKER="true"

# evaluate flags
# `-m` for mkube
# `-v` for m3-vault
# `-p` for portal-ui


while getopts ":m:" opt; do
  case $opt in
    m)
	  M3_DOCKER="$OPTARG"
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
echo "Remove Master Taint"
kubectl taint nodes --all node-role.kubernetes.io/master-
echo "Install Contour"
kubectl apply -f https://projectcontour.io/quickstart/contour.yaml
kubectl patch daemonsets -n projectcontour envoy -p '{"spec":{"template":{"spec":{"nodeSelector":{"ingress-ready":"true"},"tolerations":[{"key":"node-role.kubernetes.io/master","operator":"Equal","effect":"NoSchedule"}]}}}}'
echo "install metrics server"
kubectl apply -f deployments/metrics-dev.yaml


# Whether or not to build the m3 container and load it to kind or just load it
if [[ $M3_DOCKER == "true" ]]; then
	# Build mkube
  make --directory=".." k8sdev TAG=minio/m3:dev
else
	kind load docker-image minio/m3:dev
fi


# Apply mkube
kubectl apply -f deployments/m3-deployment.yaml

echo "done"
