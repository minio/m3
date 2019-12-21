FROM golang:1.13.4

ADD go.mod /go/src/github.com/minio/kubenginx/go.mod
ADD go.sum /go/src/github.com/minio/kubenginx/go.sum
WORKDIR /go/src/github.com/minio/kubenginx/
# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

ADD . /go/src/github.com/minio/kubenginx/
WORKDIR /go/src/github.com/minio/kubenginx/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -a -o kubenginx ./cmd/kubenginx

FROM nginx
MAINTAINER MinIO Development "dev@min.io"
EXPOSE 80

COPY --from=0 /go/src/github.com/minio/kubenginx/kubenginx    .

CMD ["/kubenginx"]
