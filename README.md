# m3 (mkube)
MinIO Kubernetes Cloud

## Prerequisites

- [Docker](https://docs.docker.com/install/)

- [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

```
curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl
```

- [Kubefwd](https://github.com/txn2/kubefwd)

```
go get github.com/txn2/kubefwd/cmd/kubefwd
```

- [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway)

```
go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
go get github.com/golang/protobuf/protoc-gen-go
```

## Installation

- Install [`kind`](https://kind.sigs.k8s.io/docs/user/quick-start/)

```
go get sigs.k8s.io/kind@v0.5.1
```

## Setup a local kubernetes (`m3cluster`) using kind
Provision the local kubernetes cluster for test/development

inside `/k8s` run:

```
cd k8s/; ./create-kind.sh
```

## Access Kubernetes dashboard

- Configure `kubectl` to work with local kubernetes

```
export KUBECONFIG=$(kind get kubeconfig-path --name="m3cluster")
```

- Launch kubectl proxy to access kubernetes dashboard

```
kubectl proxy
```

- Log in to the dashboard at  http://localhost:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/#!/login

- To get the access token

On *nix:
```
kubectl get secret $(kubectl get serviceaccount dashboard -o jsonpath="{.secrets[0].name}") -o jsonpath="{.data.token}" | base64 --decode
```

On macOS:

```
kubectl get secret $(kubectl get serviceaccount dashboard -o jsonpath="{.secrets[0].name}") -o jsonpath="{.data.token}" | base64 --decode | pbcopy
```

## Setup `m3`
(The following instructions assume that you are in the top-level directory of this repository)

- Build `m3` locally

```
make m3
```

- Run `m3 setup` on the local kubernetes

```
./m3 setup
```

- Make postgres reachable from host OS

```
kubectl port-forward -n m3 svc/postgres 5432
```

- To setup db

```
./m3 setup db
```

You may see the following error message since at times the postgres container is not running yet.
```
Running Migrations
2019/10/17 12:02:50 error connecting to database or reading migrations
2019/10/17 12:02:50 dial tcp 127.0.0.1:5432: connect: connection refused
```

If you do, try again in some time.

- To create the first admin account for `m3`
```
./m3 admin add "Admin" "your_real@email.com"
```

## Creating a new Storage Group

```
./m3 cluster sc sg add --name my-dc-rack-1
```

## Adding a new tenant
```
./m3 tenant add company-name
```

If the company name is not url-friendly a short name will be generated, but it can also be specified as shown below.

```
./m3 tenant add "Commpany® Inc." --short_name company-inc
```

> For development we need to port-forward the kubernetes pods after we add a new tenant by running:
> ```
> sudo -E kubefwd svc -n default
> ```

## Adding an Admin User

```
./m3 admin add "Admin Name" admin@email.com
```

## Making a bucket on a tenant
```
./m3 tenant bucket add tenant-short-name bucket-name
```

or

```
./m3 tenant bucket add --tenant_name tenant-short-name --bucket_name bucket-name
```

## Adding a user to a tenant's database

```
./m3 tenant user add --tenant company-inc --name somename --email user@acme.com --password user1234
```

or

```
./m3 tenant user add somename company-inc user@acme.com user1234
```

## Adding a service account

```
./m3 tenant service-account add tenant-short-name service-account-name
```

or

```
./m3 tenant service-account add --tenant_name tenant-short-name --name service-account-name --description "optional"
```

## Accessing the tenant MinIO service via browser UI

```
kubectl port-forward svc/nginx-resolver 1337:80
```

Then in your browser go to: http://company-short-name.s3.localhost:1337/, you can add more tenants and access them via a subdomain in localhost for now.
