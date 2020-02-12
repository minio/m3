# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Start using "changelog" file.
- Cli command to get storage group tenant's bucket's average usage in a defined period `./m3 cluster sc sg usage`.
- Cli command to get storage group tenant's summary in a defined period `./m3 cluster sc sg summary`.
- Support for MinIO's `admin:ServerTrace` via the `trace` action on permissions, can be added using CLI `./m3 tenant permission add acme permName Allow trace "arn:aws:s3:::*"`.

### Changed

- Moved development `etcd` and `prometheus`  to `etcd-dev.yaml` and `prometheus-dev.yaml` respectivelys
- the development `postgres` was moved to  `postgres-dev.yaml`
- `nginx-resolver` service and deployment are now exposed on the `m3-deployment.yaml`
- `Mkube` will honor the expiration token time when operators authenticate via `IDP` 

### Removed

