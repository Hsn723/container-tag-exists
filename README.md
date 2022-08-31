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
container-tag-exists IMAGE TAG
```

If `IMAGE:TAG` exists, this simply outputs `found`. This is intended to be used in CI environments to automate checking for existing container images before pushing.

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
