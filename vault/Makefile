default: docker

.PHONY: m3-vault
m3-vault:
	@echo "Building m3-vault binary to './m3-vault'"
	@(cd cmd/m3-vault; CGO_ENABLED=0 go build --ldflags "-s -w" -o ../../m3-vault)

docker:
	@docker build -t minio/m3-vault .

k8sdev:
	@docker build -t minio/m3-vault:edge .
	@kind load docker-image minio/m3-vault:edge --name m3cluster
	@echo "Done, now restart your m3-vault deployment"
