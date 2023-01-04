# container-tag-exists
[![GoDoc](https://godoc.org/github.com/Hsn723/container-tag-exists?status.svg)](https://godoc.org/github.com/Hsn723/container-tag-exists) [![Go Report Card](https://goreportcard.com/badge/github.com/Hsn723/container-tag-exists)](https://goreportcard.com/report/github.com/Hsn723/container-tag-exists) ![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/Hsn723/container-tag-exists?label=latest%20version)

Check whether a container image with the given tag exists by querying the Registry API v2. In principle, any registry implementing the Docker Registry API v2 should be supported, but this has only been confirmed with `ghcr.io` and `quay.io`.

## Installation

Install from [releases](https://github.com/Hsn723/container-tag-exists/releases) or via `go install`

```sh
go install github.com/Hsn723/container-tag-exists@latest
```

## Usage

```sh
Usage:
  container-tag-exists IMAGE TAG [flags]
  container-tag-exists [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     show version

Flags:
  -h, --help               help for container-tag-exists
  -p, --platform strings   specify platforms in the format os/arch to look for in container images. Default behavior is to look for any platform.
```

If `IMAGE:TAG` exists, this simply outputs `found`. This is intended to be used in CI environments to automate checking for existing container images before pushing. By default, `container-tag-exists` looks for any existing container image with the given tag.

```sh
container-tag-exists ghcr.io/example 0.0.0
```

If you additionally need to check for specific platforms, specify platform strings, in the format `os/arch`, to check for.

```sh
container-tag-exists ghcr.io/example 0.0.0 -p linux/amd64,linux/arm64
container-tag-exists ghcr.io/example 0.0.0 -p linux/amd64 -p linux/arm64
```

## Configuration

`container-tag-exists` first tries to retrieve the given tag unauthenticated. For public container images, this is sufficient and no further configuration is needed.

For private container images, `container-tag-exists` looks for the following environment variable(s) in this order:

| Environment variable | Description |
|----------------------| ----------- |
| `${REGISTRY_NAME}_TOKEN` | The base64 encoded bearer token |
| `${REGISTRY_NAME}_AUTH` | The basic auth token. This is basically the base64 encoded form of `$user:$pass` |
| `${REGISTRY_NAME}_USER`, `${REGISTRY_NAME}_PASSWORD` | the username/password used to authenticate to the registry |
| `GITHUB_TOKEN` | As a special case, if the registry is `ghcr.io`, the `GITHUB_TOKEN` or PAT can be used with the Registry API, provided it has sufficient permissions (`read:packages`)

The `REGISTRY_NAME` value is inferred from the registry URL part of the image name, with some special characters (`.`, `:`, `-`) being replaced by `_` and capitalized. For instance, `ghcr.io` becomes `GHCR_IO` and `container-tag-exists` therefore looks for `GHCR_IO_TOKEN`, `GHCR_IO_AUTH`, etc.
