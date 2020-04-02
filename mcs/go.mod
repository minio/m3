module github.com/minio/m3/mcs

go 1.14

require (
	github.com/elazarl/go-bindata-assetfs v1.0.0
	github.com/go-bindata/go-bindata v3.1.2+incompatible // indirect
	github.com/go-openapi/errors v0.19.4
	github.com/go-openapi/loads v0.19.5
	github.com/go-openapi/runtime v0.19.12
	github.com/go-openapi/spec v0.19.7
	github.com/go-openapi/strfmt v0.19.5
	github.com/go-openapi/swag v0.19.8
	github.com/go-openapi/validate v0.19.7
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/jessevdk/go-flags v1.4.0
	github.com/minio/mc v0.0.0-20200311043454-128f81461c9e
	github.com/minio/minio v0.0.0-20200325062613-ef6304c5c2f0
	github.com/minio/minio-go/v6 v6.0.51-0.20200319192131-097caa7760c7
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	golang.org/x/sys v0.0.0-20200331124033-c3d80250170d // indirect
	google.golang.org/genproto v0.0.0-20200401122417-09ab7b7031d2 // indirect
	google.golang.org/grpc v1.28.0 // indirect
)

replace github.com/minio/mc => github.com/dvaldivia/mc v0.0.0-20200330203654-e8aaf0b56ebd
