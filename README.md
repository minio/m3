# m3 (mkube)
MinIO Kubernetes Cloud

## Installation

- Copy the template located at `./k8s/deployments/m3-deployment.yaml.example` into `./k8s/deployments/m3-deployment.yaml`
- Modify `./k8s/deployments/m3-deployment.yaml`
- Apply `kubectl apply -f ./k8s/deployments/m3-deployment.yaml`

## Creating a MinIO cluster

You can create a new MinIO Cluster by defining it's first zone.

A Zone is made of instances (`replicas`) of homogeneous characteristics. Each instance should have the same number of drives, for this we define a `PersistenVolumeClaim` format inside the `spec.nodeTemplate.volumes` of the zone.

```
apiVersion: "mkube.min.io/v1"
kind: "Zone"
metadata:
  name: "zone-1"
spec:
  cluster: my-cluster
  image: minio/minio:RELEASE.2020-02-27T00-23-05Z
  replicas: 4
  nodeTemplate:
    env:
      - name: MINIO_ACCESS_KEY
        value: minio
      - name: MINIO_SECRET_KEY
        value: minio123
    volumes:
      - metadata:
          name: data1
        spec:
          accessModes:
            - ReadWriteOnce
          storageClassName: standard
          resources:
            requests:
              storage: 1Gi
      - metadata:
          name: data2
        spec:
          accessModes:
            - ReadWriteOnce
          storageClassName: standard
          resources:
            requests:
              storage: 1Gi
      - metadata:
          name: data3
        spec:
          accessModes:
            - ReadWriteOnce
          storageClassName: standard
          resources:
            requests:
              storage: 1Gi
      - metadata:
          name: data4
        spec:
          accessModes:
            - ReadWriteOnce
          storageClassName: standard
          resources:
            requests:
              storage: 1Gi

```

A sample zone can be found at `k8s/crds/sample/`

# Development

If you want to do some development for `m3` please refer to our [Development](DEVELOPMENT.md) document


