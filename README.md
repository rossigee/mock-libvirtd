# Mock Libvirtd

[![Build](https://github.com/rossigee/mock-libvirtd/actions/workflows/build.yaml/badge.svg)](https://github.com/rossigee/mock-libvirtd/actions/workflows/build.yaml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/rossigee/mock-libvirtd)](https://github.com/rossigee/mock-libvirtd)
[![Docker Image Size](https://img.shields.io/docker/image-size/rossigee/mock-libvirtd)](https://github.com/rossigee/mock-libvirtd/pkgs/container/mock-libvirtd)
[![License](https://img.shields.io/github/license/rossigee/mock-libvirtd)](LICENSE)

A mock libvirtd service for E2E testing of libvirtd-based applications without requiring actual hypervisor infrastructure.

## Quick Start

```bash
docker run -d -p 8080:8080 ghcr.io/rossigee/mock-libvirtd
```

## Documentation

Full documentation is available at: **https://rossigee.github.io/mock-libvirtd/**

- [Getting Started](https://rossigee.github.io/mock-libvirtd/getting-started/)
- [Configuration](https://rossigee.github.io/mock-libvirtd/configuration/)
- [API Reference](https://rossigee.github.io/mock-libvirtd/reference/)
- [State Machine](https://rossigee.github.io/mock-libvirtd/state-machine/)

## Development

```bash
go mod tidy
go run ./cmd/main
go test ./...
```

## License

[MIT](LICENSE)