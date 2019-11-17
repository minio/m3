module github.com/minio/m3

go 1.13

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fatih/color v1.7.0
	github.com/golang-migrate/migrate/v4 v4.6.2
	github.com/golang/protobuf v1.3.2
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/lib/pq v1.0.0
	github.com/minio/cli v1.21.0
	github.com/minio/mc v0.0.0-20190908212443-54ee3a280031
	github.com/minio/minio v0.0.0-20190920231956-112729386357
	github.com/minio/minio-go v0.0.0-20190327203652-5325257a208f
	github.com/minio/minio-go/v6 v6.0.35
	github.com/pelletier/go-toml v1.6.0
	github.com/satori/go.uuid v1.2.0
	golang.org/x/crypto v0.0.0-20190820162420-60c769a6c586
	golang.org/x/net v0.0.0-20191101175033-0deb6923b6d9
	golang.org/x/sys v0.0.0-20191029155521-f43be2a4598c // indirect
	google.golang.org/genproto v0.0.0-20191028173616-919d9bdd9fe6 // indirect
	google.golang.org/grpc v1.24.0
	k8s.io/api v0.0.0-20190313115550-3c12c96769cc
	k8s.io/apimachinery v0.0.0-20190313115320-c9defaaddf6f
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20190923111123-69764acb6e8e // indirect
)

// Added for go1.13 migration https://github.com/golang/go/issues/32805
replace github.com/gorilla/rpc v1.2.0+incompatible => github.com/gorilla/rpc v1.2.0
