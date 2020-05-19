#!/bin/bash
# Get's the latest deployment file from MinIO Operator and replaces it in the m3-deployment.template.yaml
cat k8s/deployments/m3-deployment.yaml | OPERATORYAML=$(curl https://raw.githubusercontent.com/minio/minio-operator/master/minio-operator.yaml) awk '{gsub(/^\!~operator_file~\!$/, ENVIRON["OPERATORYAML"]); print}' > m3.yaml
