FROM golang:1.13.4

ADD go.mod /go/src/github.com/minio/m3-vault/go.mod
ADD go.sum /go/src/github.com/minio/m3-vault/go.sum
WORKDIR /go/src/github.com/minio/m3-vault/
# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

ADD . /go/src/github.com/minio/m3-vault/
WORKDIR /go/src/github.com/minio/m3-vault/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -a -o m3-vault ./cmd/m3-vault

FROM vault
MAINTAINER MinIO Development "dev@min.io"
EXPOSE 8200

COPY --from=0 /go/src/github.com/minio/m3-vault/m3-vault    .
COPY --from=0 /go/src/github.com/minio/m3-vault/vault-config.json    .
COPY --from=0 /go/src/github.com/minio/m3-vault/tls/    /tls/

ENTRYPOINT []
CMD ["/m3-vault"]
