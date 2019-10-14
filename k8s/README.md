Getting started
=====

Requirements
---
- [Docker](https://www.docker.com)

	Build docker image:

	```
	$ docker build -t m3kind .
 	```

- Install [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

Installation
---

#### Install `kind`

```
$ GO111MODULE="on" go get sigs.k8s.io/kind@v0.5.1
```
> Note: If the message `kind: command not found` appears when running those commands add the `$GOPATH` variable to the `$PATH` variable with: `$ export PATH=$PATH:$(go env GOPATH)/bin`
else, refer to [kind docs](https://kind.sigs.k8s.io/docs/user/quick-start/)

then create cluster and sample tenant

```
$ chmod +x create-kind.sh
$ ./create-kind.sh
```


#### Configure `kubectl`
```
$ export KUBECONFIG="$(kind get kubeconfig-path --name="m3cluster")"
```


forward the service to your local
```
$ kubectl port-forward service/tenant-1 9001
```

To get the access token 

```
$ kubectl get secret $(kubectl get serviceaccount dashboard -o jsonpath="{.secrets[0].name}") -o jsonpath="{.data.token}" | base64 --decode | pbcopy
```

Launch proxy to access kubernetes dashboard
```
$ kubectl proxy
```
After that go to http://localhost:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/#!/login and enter the authentication token
