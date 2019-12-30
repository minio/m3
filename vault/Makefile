default: docker

docker:
	@docker build -t minio/m3-vault .

k8sdev:
	@docker build -t minio/m3-vault:edge .
	@kind load docker-image minio/m3-vault:edge --name m3cluster
	@echo "Done, now restart your m3-vault deployment"
