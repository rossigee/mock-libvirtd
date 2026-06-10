---
title: Home
---

[![Build](https://github.com/rossigee/mock-libvirtd/actions/workflows/build.yaml/badge.svg)](https://github.com/rossigee/mock-libvirtd/actions/workflows/build.yaml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/rossigee/mock-libvirtd)](https://github.com/rossigee/mock-libvirtd)
[![Docker Image Size](https://img.shields.io/docker/image-size/rossigee/mock-libvirtd)](https://github.com/rossigee/mock-libvirtd/pkgs/container/mock-libvirtd)
[![License](https://img.shields.io/github/license/rossigee/mock-libvirtd)](LICENSE)

A mock libvirtd service for E2E testing of libvirtd-based applications without requiring actual hypervisor infrastructure. Provides HTTP/REST endpoints for managing virtual machines, networks, and storage pools.

Supports testing against libvirt-compatible clients including **Proxmox VE**, QEMU/KVM, and other libvirt-based virtualization platforms.

## Overview

This service emulates the libvirtd/libvirt daemon API for testing purposes. It supports:

- **Domains**: Virtual machine lifecycle management (create, list, get, update, delete)
- **Networks**: Virtual network management (create, list, get, delete)
- **Storage**: Storage pool management (create, list, get, delete)
- **Volumes**: Storage volume management (create, list, get, delete)

All data is stored in-memory and reset on service restart.

## Quick Links

- [Getting Started](/getting-started/) - Run your first VM
- [Configuration](/configuration/) - Environment variables
- [API Reference](/reference/) - All endpoints
- [State Machine](/state-machine/) - Domain lifecycle
- [OpenAPI Spec](/openapi.yaml) - Raw OpenAPI/Swagger spec

## Use Cases

### CI/CD Testing

Run the mock service in your CI pipeline to test libvirt client code without spinning up actual VMs:

```yaml
# .github/workflows/test.yaml
- name: Run mock-libvirtd
  run: docker run -d -p 8080:8080 ghcr.io/rossigee/mock-libvirtd
- name: Run tests
  run: go test -v ./...
```

### Local Development

Develop libvirt client applications without requiring KVM/Proxmox on your machine:

```bash
docker run -d -p 8080:8080 ghcr.io/rossigee/mock-libvirtd
# Your client code can now connect to localhost:8080
```

### Kubernetes

Deploy alongside your application in staging environments:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mock-libvirtd
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mock-libvirtd
  template:
    metadata:
      labels:
        app: mock-libvirtd
    spec:
      containers:
      - name: mock-libvirtd
        image: ghcr.io/rossigee/mock-libvirtd
        ports:
        - containerPort: 8080
```

## Features

- **REST API**: Full CRUD operations for domains, networks, storage pools, and volumes
- **State Machine**: Realistic VM lifecycle with boot delays and state transitions
- **Metrics**: Prometheus-compatible metrics endpoint
- **Tracing**: OpenTelemetry/OTLP support for distributed tracing
- **Rate Limiting**: Configurable request rate limiting
- **Input Validation**: Validate names, memory, CPU limits
- **CORS**: Configurable CORS for browser-based testing

## License

[MIT](LICENSE)