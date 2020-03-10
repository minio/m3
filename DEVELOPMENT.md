# m3 (mkube)
To do development for m3 we recommend setup a local Kubernetes cluster. In this guide we do steps on how to use [kind](https://kind.sigs.k8s.io/).

## Prerequisites

- [Docker](https://docs.docker.com/install/)

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

A binary release can be downloaded via 

```
curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl
```

## Installation

- Install [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)

```
go get sigs.k8s.io/kind@v0.7.0
```

## Setup a local kubernetes using kind
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
make k8sdev TAG=minio/m3:dev
```
- Copy the template located at `./k8s/deployments/m3-deployment.yaml.example` into `./k8s/deployments/m3-deployment.yaml`
- Modify `./k8s/deployments/m3-deployment.yaml`

Replace all the `<TOKENS>` with their corresponding values, for example `<DEV_EMAIL>` with your personal email.
A valid `smtp` account is needed, if you don't have one we recommend you create a gmail account and enable [Less Secure Apps access](https://support.google.com/accounts/answer/6010255?hl=en)

- Install the m3 deployment on kubernetes
```
kubectl apply -f k8s/deployments/m3-deployment.yaml
``` 

- Start m3 development environment

```
./m3 dev
```


## Using a custom CA for the KMS

Mkube supports the use of custom CA for the kms if need it.
Load the custom `ca certificate` using a `configmap`

```
kubectl create configmap kms-ca-cert --from-file=customCA.crt
```

Then set the `configmap` name and `certificate file name` on `m3-deployment.yaml` by uncommenting the following env variables
``` 
KMS_CA_CERT_CONFIG_MAP: "kms-ca-cert"
KMS_CA_CERT_FILE_NAME: "customCA.crt"
```

Apply changes for `mkube` running `kubectl apply -f k8s/deployment/m3-deployment.yaml`
