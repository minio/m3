## Installation

Run m3-vault binary

```
go build ./cmd/m3-vault
./m3-vault
```

Run `m3-vault` docker image locally

```
docker run --cap-add=IPC_LOCK --rm -p 8200:8200 -e "TOTAL_INIT_RETRIES=5" minio/m3-vault:latest
```

Build a new docker image

```
make docker
```

Push `m3-vault` image into `m3cluster`

```
make k8sdev
```