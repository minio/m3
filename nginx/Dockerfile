FROM golang:1.13.4

ADD go.mod /go/src/github.com/minio/m3-nginx/go.mod
ADD go.sum /go/src/github.com/minio/m3-nginx/go.sum
WORKDIR /go/src/github.com/minio/m3-nginx/
# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

ADD . /go/src/github.com/minio/m3-nginx/
WORKDIR /go/src/github.com/minio/m3-nginx/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -a -o m3-nginx ./cmd/m3-nginx

FROM nginx
MAINTAINER MinIO Development "dev@min.io"
EXPOSE 80

COPY --from=0 /go/src/github.com/minio/m3-nginx/m3-nginx    .

CMD ["/m3-nginx"]
