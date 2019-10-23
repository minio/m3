default: grpc

grpc:
	protoc -I/usr/local/include -I=protos --go_out=plugins=grpc:stubs protos/*.proto

grpc-gateway:
	protoc -I/usr/local/include -I=protos  --grpc-gateway_out=logtostderr=true,grpc_api_configuration=protos/public_api_rest.yaml:stubs protos/*.proto

swagger-def:
	protoc -I/usr/local/include -I=protos --swagger_out=logtostderr=true,grpc_api_configuration=protos/public_api_rest.yaml:swagger protos/public_api.proto
