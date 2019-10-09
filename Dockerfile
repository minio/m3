FROM golang:1.13.1

#ENV GO111MODULE=on
#ENV GOPROXY=https://proxy.golang.org
#ENV GOPROXY=direct

ADD go.mod /go/src/github.com/minio/m3/go.mod
ADD go.sum /go/src/github.com/minio/m3/go.sum
WORKDIR /go/src/github.com/minio/m3/
# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

ADD . /go/src/github.com/minio/m3/
WORKDIR /go/src/github.com/minio/m3/
WORKDIR /

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o main ./go/src/github.com/minio/m3/

FROM scratch
MAINTAINER MinIO Development "dev@min.io"
EXPOSE 10105

COPY --from=0 /main    .

CMD ["/main"]
