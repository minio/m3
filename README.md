# m3 (mkube)
MinIO Kubernetes Cloud

## Requirements

- [Docker](https://www.docker.com)

- [Kubefwd](https://github.com/txn2/kubefwd)

- [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

## Installation

- Install [`kind`](https://kind.sigs.k8s.io/docs/user/quick-start/)

```shell
$ GO111MODULE="on" go get sigs.k8s.io/kind@v0.5.1
```

## Setup a local kubernetes (`m3cluster`) using kind
Provision the local kubernetes cluster for test/development

inside `/k8s` run:

```shell

$ ./create-kind.sh
```

## Access Kubernetes dashboard

1. Configure `kubectl` to work with local kubernetes

```shell
$ export KUBECONFIG="$(kind get kubeconfig-path --name="m3cluster")"
```

2. Launch kubectl proxy to access kubernetes dashboard
```shell
$ kubectl proxy
```

3. Log in to the dashboard at  http://localhost:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/#!/login

4. To get the access token,

On *nix,
```shell
$ kubectl get secret $(kubectl get serviceaccount dashboard -o jsonpath="{.secrets[0].name}") -o jsonpath="{.data.token}" | base64 --decode
```

On Mac OS,
```shell
$ kubectl get secret $(kubectl get serviceaccount dashboard -o jsonpath="{.secrets[0].name}") -o jsonpath="{.data.token}" | base64 --decode | pbcopy
```

## Setup `m3`
(The following instructions assume that you are in the top-level directory of this repository)
1. Build `m3` locally
```shell
   go build ./cmd/m3
```

2. Run `m3 setup` on the local kubernetes
```shell
   ./m3 setup
```
At the moment, you would see the following error message,
```
Running Migrations
2019/10/17 12:02:50 error connecting to database or reading migrations
2019/10/17 12:02:50 dial tcp 127.0.0.1:5432: connect: connection refused
```

This is benign and can be fixed with the following steps,

```shell
  kubectl port-forward -n m3 svc/postgres 5432
  ./m3 setup db
```

## Creating a Storage Group
```shell
  ./m3 cluster sc sg add --name my-dc-rack-1
```

## Adding a tenant
```shell
  ./m3 tenant add company-name
```

If the company name is not url-friendly a short name will be generated, but it can also be specified.

```shell
  ./m3 tenant add "Commpany® Inc." --short_name company-inc
```

---
For development we need to port-forward the kubernetes pods after we add a new tenant by running:

```shell
  sudo -E kubefwd svc -n default
```
---
## Making a bucket on a tenant
```shell
  ./m3 tenant bucket add tenant-short-name bucket-name
```
or 
```shell
  ./m3 tenant bucket add --tenant_name tenant-short-name --bucket_name bucket-name
```

## Adding a user to a tenant's database
```shell
  ./m3 tenant user add --tenant company-inc --email user@acme.com --password user1234
``` 
or 
```shell
  ./m3 tenant user add company-inc user@acme.com user1234
```

## Adding a service account
```shell
  ./m3 tenant service-account add tenant-short-name service-account-name
```
or 
```shell
  ./m3 tenant service-account add --tenant_name tenant-short-name --name service-account-name --description "optional"
```

## Accessing the tenant MinIO service via browser UI
```shell
  kubectl port-forward svc/nginx-resolver 1337:80
```
Then in your browser go to: http://company-short-name.s3.localhost:1337/, you can add more tenants and access them via a subdomain in localhost for now.
