# m3 (mkube)
MinIO Kubernetes Cloud

## Prerequisites

- [Docker](https://docs.docker.com/install/)

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

A binary release can be downloaded via 

```
curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl
```

- [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway)

```
go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
go get github.com/golang/protobuf/protoc-gen-go
```

## Installation

- Install [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)

```
go get sigs.k8s.io/kind@v0.6.0
```

## Setup a local kubernetes (`m3cluster`) using kind
Provision the local kubernetes cluster for test/development

inside `/k8s` run:

```
cd k8s/; ./create-kind.sh
```

## Access Kubernetes dashboard

- Launch kubectl proxy to access kubernetes dashboard

```
kubectl proxy
```

- Log in to the dashboard at `http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/#!/login`


- To get the access token

```
kubectl get secret $(kubectl get serviceaccount dashboard -o jsonpath="{.secrets[0].name}") -o jsonpath="{.data.token}" | base64 --decode
```

## Setup `m3`
(The following instructions assume that you are in the top-level directory of this repository)

- Build `m3` locally

```
make m3
```

- Build `m3` local docker image and push it to your local kubernetes

```
make k8sdev
```

- Modify `./k8s/deployments/m3-deployment.yaml`

Replace all the <TOKENS> with their corresponding values, for example <DEV_EMAIL> with your personal email.
A valid `smtp` account is needed, if you don't have one we recommend you create a gmail account and enable [Less Secure Apps access](https://support.google.com/accounts/answer/6010255?hl=en)

- Install the m3 deployment on kubernetes
```
kubectl apply -f k8s/deployments/m3-deployment.yaml
``` 

- Start m3 development environment

```
./m3 dev
```

- You should get an email with your activation command, execute it
```
./m3 set-password <YOUR_TOKEN>
```
- Finally, perform login to the cluster so you can run all the commands
```
./m3 login
```

## Creating a new Storage Cluster

```
./m3 cluster sc add --name my-dc-rack-1
```

## Creating a nodes to store the information

```
./m3 cluster nodes add --name node-1 --k8s_label m3cluster-worker --volumes /mnt/disk{1...4}
```

For development add 3 additional nodes

```
./m3 cluster nodes add --name node-2 --k8s_label m3cluster-worker2 --volumes /mnt/disk{1...4}
./m3 cluster nodes add --name node-3 --k8s_label m3cluster-worker3 --volumes /mnt/disk{1...4}
./m3 cluster nodes add --name node-4 --k8s_label m3cluster-worker4 --volumes /mnt/disk{1...4}
```

## Associate the nodes to a storage cluster

```
./m3 cluster nodes assign --storage_cluster my-dc-rack-1 --node node-1
```

For development assign the remaining 3 nodes
```
./m3 cluster nodes assign --storage_cluster my-dc-rack-1 --node node-2
./m3 cluster nodes assign --storage_cluster my-dc-rack-1 --node node-3
./m3 cluster nodes assign --storage_cluster my-dc-rack-1 --node node-4
```

## Creating a new Storage Group

```
./m3 cluster sc sg add --storage_cluster my-dc-rack-1 --name group-1
```

## Adding a new tenant
```
./m3 tenant add company-name --admin_name="John Doe" --admin_email="email@domain.com"
```

If the company name is not url-friendly a short name will be generated, but it can also be specified as shown below.

```
./m3 tenant add "Commpany® Inc." --short_name company-inc --admin_name="John Doe" --admin_email="email@domain.com"
```

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

## Adding a new permission
To grant `write` permission to `bucketA` and `bucketB` 

`./m3 tenant permission add acme SAMPLE allow write "bucketA,bucketB"`

## Accessing the tenant MinIO service via browser UI

The nginx router should be exposed on your local on port `9000` after doing `./m3 dev`

Modify your `/etc/hosts` and add the following record

```
127.0.0.1   s3.localhost
```

Then in your browser go to: http://company-short-name.s3.localhost:9000/, you can add more tenants and access them via a subdomain in localhost for now.
