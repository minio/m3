FROM golang:1.13

ADD go.mod /go/src/github.com/minio/m3-nginx/go.mod
ADD go.sum /go/src/github.com/minio/m3-nginx/go.sum
WORKDIR /go/src/github.com/minio/m3-nginx/

# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

ADD . /go/src/github.com/minio/m3-nginx/
WORKDIR /go/src/github.com/minio/m3-nginx/

ENV CGO_ENABLED=0
RUN go build -ldflags "-w -s" -o m3-nginx ./cmd/m3-nginx

FROM nginx
MAINTAINER MinIO Development "dev@min.io"
EXPOSE 80

COPY --from=0 /go/src/github.com/minio/m3-nginx/m3-nginx  .

CMD ["/m3-nginx"]
