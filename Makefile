default: grpc

grpc:
	@echo "Generating grpc stubs"
	@protoc -I=protos protos/*.proto --go_out=plugins=grpc:portal/stubs

grpc-gateway:
	@echo "Generating grpc-gateway stubs"
	@protoc -I=protos --grpc-gateway_out=logtostderr=true:. protos/*.proto

swagger-def:
	@echo "Generating swagger-def stubs"
	@protoc -I=protos --swagger_out=logtostderr=true:. protos/public_api.proto

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
	@docker build -t minio/m3:dev .	
	@kind load docker-image minio/m3:dev --name m3cluster
	@echo "Done, now restart your m3 deployment"