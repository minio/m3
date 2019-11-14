FROM golang:1.13.4

ADD go.mod /go/src/github.com/minio/m3/go.mod
ADD go.sum /go/src/github.com/minio/m3/go.sum
WORKDIR /go/src/github.com/minio/m3/
# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

ADD . /go/src/github.com/minio/m3/
WORKDIR /go/src/github.com/minio/m3/


RUN apt-get install -y ca-certificates
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -a -o m3 ./cmd/m3

FROM scratch
MAINTAINER MinIO Development "dev@min.io"
EXPOSE 9009

COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /go/src/github.com/minio/m3/m3    .
ADD ./cluster/templates /cluster/templates

CMD ["/m3"]
