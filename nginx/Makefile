default: build

build:
	@echo "Building frontend static assets to 'build'"
	yarn build

clean:
	@echo "Cleaning up all the generated files"
	@find . -name '*.test' | xargs rm -fv
	@find . -name '*~' | xargs rm -fv
	@rm -rvf m3

docker:
	@docker build -t minio/m3-portal-frontend .

k8sdev:
	@docker build -t minio/m3-portal-frontend:dev .
	@kind load docker-image minio/m3-portal-frontend:dev --name m3cluster
	@echo "Done, now restart your m3-portal-frontend deployment"