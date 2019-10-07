export KUBECONFIG="$(kind get kubeconfig-path --name="m3cluster")"
kubectl cluster-info
