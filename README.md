# REANA-Datastore-S3FS

[![image](https://github.com/reanahub/reana-datastore-s3fs/actions/workflows/ci.yml/badge.svg)](https://github.com/reanahub/reana-datastore-s3fs/actions)
[![image](https://codecov.io/gh/reanahub/reana-datastore-s3fs/branch/master/graph/badge.svg)](https://codecov.io/gh/reanahub/reana-datastore-s3fs)
[![image](https://img.shields.io/badge/discourse-forum-blue.svg)](https://forum.reana.io)
[![image](https://img.shields.io/github/license/reanahub/reana-datastore-s3fs.svg)](https://github.com/reanahub/reana-datastore-s3fs/blob/master/LICENSE)

## About

REANA-Datastore-S3FS is a component of the [REANA](https://www.reana.io/)
reusable and reproducible research data analysis platform. It provides a
datastore sidecar container image for mounting S3-compatible object storage into
REANA workloads by using S3FS.

## Features

- mount S3-compatible object storage into REANA workload pods
- expose mounted data to interactive sessions and Kubernetes jobs
- provide a reusable datastore sidecar image for REANA cluster components

## Usage

The datastore sidecar automatically mounts S3-compatible object storage into
REANA workload pods based on environment variables. It exposes a health check
endpoint and handles graceful shutdown with proper unmounting.

### Environment Variables

The sidecar reads configuration from environment variables with the following
pattern:

```bash
S3_TO_LOCAL_{alias}_ALIAS={alias}
S3_TO_LOCAL_{alias}_BUCKET={bucket-name}
S3_TO_LOCAL_{alias}_HOST={s3-host-url}
S3_TO_LOCAL_{alias}_REGION={aws-region}
S3_TO_LOCAL_{alias}_ACCESS_KEY={access-key}
S3_TO_LOCAL_{alias}_SECRET_KEY={secret-key}
```

Example:

```bash
S3_TO_LOCAL_mydata_ALIAS=mydata
S3_TO_LOCAL_mydata_BUCKET=my-s3-bucket
S3_TO_LOCAL_mydata_HOST=https://s3.example.com
S3_TO_LOCAL_mydata_REGION=us-east-1
S3_TO_LOCAL_mydata_ACCESS_KEY=my-access-key
S3_TO_LOCAL_mydata_SECRET_KEY=my-secret-key
```

The sidecar will mount the bucket at `/s3-data/{alias}/{bucket}`.

### Endpoints

- `GET /health` - Returns HTTP 200 with `{"status": "ready"}` when mounts are
  complete, or HTTP 503 with `{"status": "mounting"}` during initialization.
- `POST /shutdown` - Initiates graceful shutdown, triggering unmount of all S3FS
  mounts.

For detailed information on how to install and use REANA, see
[docs.reana.io](https://docs.reana.io).

## Useful links

- [REANA project home page](https://www.reana.io/)
- [REANA user documentation](https://docs.reana.io)
- [REANA user support forum](https://forum.reana.io)
- [REANA-Datastore-S3FS docker images](https://hub.docker.com/r/reanahub/reana-datastore-s3fs)
- [REANA-Datastore-S3FS known issues](https://github.com/reanahub/reana-datastore-s3fs/issues)
- [REANA-Datastore-S3FS source code](https://github.com/reanahub/reana-datastore-s3fs)
