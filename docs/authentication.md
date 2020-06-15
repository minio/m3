# Mkube authentication

Currently, Mkube support only bearer token authentication for secured endpoints,
clients that wishes to access resources behind a secured endpoint will need to provide
a jwt token using the HTTP `Authorization` header, ie: 

```
curl --location --request GET 'http://localhost:8787/api/v1/tenants' --header 'Authorization: Bearer eyJ...'
```

The provided `JWT token` corresponds to the `Kubernetes service account` that Mkube will use to run tasks on behalf of the client
ie: list, create, edit, delete tenants, etc.

# Development


If you are running mkube in your local environment and wish to connect with a kubernetes cluster for testing you can use the
`M3_K8S_API_SERVER` and `M3_K8S_API_SERVER_INSECURE` environment variables, ie:

```bash
M3_K8S_API_SERVER=https://localhost:52461 M3_K8S_API_SERVER_INSECURE=on ./m3 server
```

By default, if you don't provide `M3_K8S_API_SERVER` mkube will try to obtain the IP address of the k8s Api server (assuming is running inside
kubernetes cluster), if mkube can't do that then will use `http://localhost:8001` (commonly used when running `kubectl proxy`)

## Extract the Service account token

For local development you can use the jwt associated to the `m3-sa` service account, you can get the token running
the following command in your terminal:

```
kubectl get secret $(kubectl get serviceaccount m3-sa -o jsonpath="{.secrets[0].name}") -o jsonpath="{.data.token}" | base64 --decode
```

Then test the token works with `curl`
```
curl --location --request GET 'http://localhost:8787/api/v1/tenants' --header 'Authorization: Bearer eyJ...'
...
{
    "tenants": [
        {
            "creation_date": "2020-06-08 22:35:50 -0700 PDT",
            "currentState": "Ready",
            "instance_count": 4,
            "name": "minio",
            "volume_count": 16,
            "volume_size": 1099511627776,
            "zone_count": 1
        }
    ]
}
```