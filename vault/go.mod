module m3/vault

go 1.13

require (
	github.com/fatih/color v1.7.0
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/hashicorp/vault/api v1.0.5-0.20191216174727-9d51b36f3ae4
	github.com/minio/minio v0.0.0-20191230055646-8eba97da74ef
	k8s.io/api v0.16.4
	k8s.io/apimachinery v0.16.4
	k8s.io/client-go v0.16.4
	k8s.io/utils v0.0.0-20191218082557-f07c713de883 // indirect
)

replace github.com/gorilla/rpc v1.2.0+incompatible => github.com/gorilla/rpc v1.2.0

replace github.com/hashicorp/vault => github.com/hashicorp/vault v1.3.1
