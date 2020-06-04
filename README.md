# m3 (mkube)
MinIO Kubernetes Cloud

## Installation

You can apply all the files located in `k8s/base/` to install `m3`

```bash
kubectl apply -f k8s/base/
```

Or you can use [kustomize](https://github.com/kubernetes-sigs/kustomize) to build a single file to apply which supports customizations

```bash
kustomize build k8s/base/ | kubectl apply -f -
```

# Development

If you want to do some development for `m3` please refer to our [Development](DEVELOPMENT.md) document


