Getting started
=====

Requirements
---

`kubectl`

install
---

Install kind
`GO111MODULE="on" go get sigs.k8s.io/kind@v0.5.1`
then create cluster and sample tenant
```
chmod +x create-kind-cluster.sh
./create-kind-cluster.sh
```


configure `kubectl`
```
export KUBECONFIG="$(kind get kubeconfig-path --name="m3cluster")"
```


forward the service to your local
```
kubectl port-forward service/tenant-1 9001
```