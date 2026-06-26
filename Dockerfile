# This file is part of REANA.
# Copyright (C) 2026 CERN.
#
# REANA is free software; you can redistribute it and/or modify it
# under the terms of the MIT License; see LICENSE file for more details.

# Build sidecar binary
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.26.4-bookworm AS builder

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -trimpath \
    -ldflags "-s -w -X github.com/reanahub/reana-datastore-s3fs/internal/version.Version=${VERSION}" \
    -o /out/reana-datastore-s3fs \
    ./cmd/reana-datastore-s3fs

# Collect S3FS runtime dependencies for the target platform
FROM docker.io/library/debian:bookworm-slim AS s3fs-runtime

# hadolint ignore=DL3008,DL4006
RUN apt-get update -y && \
    apt-get install --no-install-recommends -y \
      ca-certificates \
      fuse3 \
      s3fs && \
    mkdir -p /runtime-root/usr/bin /runtime-root/bin /runtime-root/etc/ssl/certs /runtime-root/etc && \
    cp /usr/bin/s3fs /runtime-root/usr/bin/s3fs && \
    cp /bin/fusermount3 /runtime-root/bin/fusermount3 && \
    cp /etc/ssl/certs/ca-certificates.crt /runtime-root/etc/ssl/certs/ca-certificates.crt && \
    echo "user_allow_other" > /runtime-root/etc/fuse.conf && \
    for binary in /usr/bin/s3fs /bin/fusermount3; do \
      ldd "$binary" | awk '/=> \// { print $3 } /^\// { print $1 }' | sort -u | while read -r library; do \
        mkdir -p "/runtime-root$(dirname "$library")"; \
        cp -L "$library" "/runtime-root$library"; \
      done; \
    done && \
    for loader in \
      /lib64/ld-linux-x86-64.so.2 \
      /lib/ld-linux-aarch64.so.1 \
      /lib/aarch64-linux-gnu/ld-linux-aarch64.so.1 \
      /lib/x86_64-linux-gnu/ld-linux-x86-64.so.2; do \
      if [ -e "$loader" ]; then \
        mkdir -p "/runtime-root$(dirname "$loader")"; \
        cp -L "$loader" "/runtime-root$loader"; \
      fi; \
    done && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Run only the sidecar binary plus S3FS runtime dependencies
# hadolint ignore=DL3007
FROM gcr.io/distroless/base-debian12:latest

COPY --from=s3fs-runtime /runtime-root/ /
COPY --from=builder /out/reana-datastore-s3fs /usr/local/bin/reana-datastore-s3fs

# Set image labels
LABEL org.opencontainers.image.authors="team@reanahub.io"
LABEL org.opencontainers.image.created="2026-06-26"
LABEL org.opencontainers.image.description="Image for mounting S3-compatible object storage into REANA workloads"
LABEL org.opencontainers.image.documentation="https://github.com/reanahub/reana-datastore-s3fs/blob/master/README.md"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/reanahub/reana-datastore-s3fs"
LABEL org.opencontainers.image.title="reana-datastore-s3fs"
LABEL org.opencontainers.image.url="https://github.com/reanahub/reana-datastore-s3fs"
LABEL org.opencontainers.image.vendor="reanahub"
# x-release-please-start-version
LABEL org.opencontainers.image.version="1.0.0"
# x-release-please-end

# The FUSE sidecar needs root privileges to perform and tear down mounts.
# hadolint ignore=DL3002
USER 0:0

ENTRYPOINT ["/usr/local/bin/reana-datastore-s3fs"]
