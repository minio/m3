# Sets the build version based on the output of the following command, if we are building for a tag, that's the build else it uses the current git branch as the build
BUILD_VERSION:=$(shell git describe --exact-match --tags $(git log -n1 --pretty='%h') 2>/dev/null || git rev-parse --abbrev-ref HEAD 2>/dev/null)
BUILD_TIME:=$(shell date 2>/dev/null)
TAG ?= "minio/m3:$(VERSION)-dev"

default: m3

.PHONY: m3
m3:
	@echo "Building m3 binary to './m3'"
	@(cd cmd/m3; CGO_ENABLED=0 go build --ldflags "-s -w" -o ../../m3)

docker:
	@docker build -t minio/m3 --build-arg build_version=$(BUILD_VERSION) --build-arg build_time='$(BUILD_TIME)' .

k8sdev:
	@docker build -t $(TAG) --build-arg build_version=$(BUILD_VERSION) --build-arg build_time='$(BUILD_TIME)' .
	@kind load docker-image $(TAG)
	@echo "Done, now restart your m3 deployment"

swagger-gen:
	@echo "Generating swagger server code from yaml"
	@swagger generate server -A m3 --main-package=m3 --exclude-main -P models.Principal -f ./swagger.yml -r NOTICE

yamlgen:
	@./k8s/yamlgen.sh

test:
	@(GO111MODULE=on go test -race -v github.com/minio/m3/restapi/...)

coverage:
	@(GO111MODULE=on go test -v -coverprofile=coverage.out github.com/minio/m3/restapi/... && go tool cover -html=coverage.out && open coverage.html)

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@find . -name '*~' | xargs rm -fv
	@rm -rvf m3