default: docker

docker:
	@docker build -t minio/m3-nginx .

k8sdev:
	@docker build -t minio/m3-nginx:edge .
	@kind load docker-image minio/m3-nginx:edge --name m3cluster
	@echo "Done, now restart your nginx-resolver deployment"
