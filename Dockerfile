FROM golang:1.13.7

ADD go.mod /go/src/github.com/minio/m3/go.mod
ADD go.sum /go/src/github.com/minio/m3/go.sum
WORKDIR /go/src/github.com/minio/m3/
# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

ADD . /go/src/github.com/minio/m3/
WORKDIR /go/src/github.com/minio/m3/

ENV CGO_ENABLED=0

ARG build_version
ARG build_time
ENV env_build_version=$build_version
ENV env_build_time=$build_time

RUN apt-get update -y && apt-get install -y ca-certificates
RUN go build -ldflags "-w -s -X 'github.com/minio/m3/version.BuildTime=$env_build_time' -X 'github.com/minio/m3/version.BuildVersion=$env_build_version'" -a -o m3 ./cmd/m3

FROM scratch
MAINTAINER MinIO Development "dev@min.io"
EXPOSE 9009

COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /go/src/github.com/minio/m3/m3 .
ADD ./cluster/templates /cluster/templates
ADD ./cluster/migrations /cluster/migrations
ADD ./cluster/tenant-migrations /cluster/tenant-migrations

CMD ["/m3"]
