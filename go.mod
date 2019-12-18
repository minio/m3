module github.com/minio/m3

go 1.13

require (
	github.com/coreos/etcd v3.3.12+incompatible
	github.com/coreos/go-oidc v2.0.0+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20191011121108-aa519ddbe484 // indirect
	github.com/fatih/color v1.7.0
	github.com/golang-migrate/migrate/v4 v4.7.0
	github.com/golang/protobuf v1.3.2
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gosimple/slug v1.9.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/grpc-gateway v1.12.1
	github.com/lib/pq v1.1.1
	github.com/minio/cli v1.22.0
	github.com/minio/mc v0.0.0-20191012041914-735aa139b19c
	github.com/minio/minio v0.0.0-20191213205711-c3d4c1f58435
	github.com/minio/minio-go/v6 v6.0.44
	github.com/pelletier/go-toml v1.6.0
	github.com/rs/cors v1.6.0
	github.com/satori/go.uuid v1.2.0
	golang.org/x/crypto v0.0.0-20191117063200-497ca9f6d64f
	golang.org/x/net v0.0.0-20191101175033-0deb6923b6d9
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/sys v0.0.0-20191029155521-f43be2a4598c // indirect
	google.golang.org/genproto v0.0.0-20191115221424-83cc0476cb11 // indirect
	google.golang.org/grpc v1.24.0
	gopkg.in/yaml.v2 v2.2.7 // indirect
	k8s.io/api v0.0.0-20190313115550-3c12c96769cc
	k8s.io/apimachinery v0.0.0-20190313115320-c9defaaddf6f
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20190923111123-69764acb6e8e // indirect
)

// Added for go1.13 migration https://github.com/golang/go/issues/32805
replace github.com/gorilla/rpc v1.2.0+incompatible => github.com/gorilla/rpc v1.2.0
