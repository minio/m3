default: grpc

grpc:
	@echo "Generating grpc stubs"
	@protoc -I/usr/local/include -I=protos --go_out=plugins=grpc:stubs protos/*.proto

grpc-gateway:
	@echo "Generating grpc-gateway stubs"
	@protoc -I/usr/local/include -I=protos  --grpc-gateway_out=logtostderr=true,grpc_api_configuration=protos/public_api_rest.yaml:stubs protos/*.proto

swagger-def:
	@echo "Generating swagger-def stubs"
	@protoc -I/usr/local/include -I=protos --swagger_out=logtostderr=true,grpc_api_configuration=protos/public_api_rest.yaml:swagger protos/public_api.proto

m3:
	@echo "Building m3 binary to './m3'"
	@(cd cmd/m3; CGO_ENABLED=0 go build --ldflags "-s -w" -o ../../m3)
