VERSION ?= $(shell git describe --tags)
TAG ?= "minio/m3:$(VERSION)-dev"

default: m3

grpc:
	@echo "Generating grpc stubs"
# 	@protoc -I=protos protos/*.proto --go_out=plugins=grpc:api/stubs
	@protoc -I=protos --grpc-gateway_out=logtostderr=true,grpc_api_configuration=protos/public_api_rest.yaml:api/stubs --go_out=plugins=grpc:api/stubs protos/*.proto

grpc-gateway:
	@echo "Generating grpc-gateway stubs"
	@protoc -I=protos --grpc-gateway_out=logtostderr=true:. protos/*.proto

swagger-def:
	@echo "Generating swagger-def stubs"
	@protoc -I=protos --swagger_out=logtostderr=true:. protos/public_api.proto

.PHONY: m3
m3:
	@echo "Building m3 binary to './m3'"
	@(cd cmd/m3; CGO_ENABLED=0 go build --ldflags "-s -w" -o ../../m3)

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@find . -name '*~' | xargs rm -fv
	@rm -rvf m3

docker:
	@docker build -t minio/m3 .

k8sdev:
	@docker build -t $(TAG) .
	@kind load docker-image $(TAG) --name m3cluster
	@echo "Done, now restart your m3 deployment"
