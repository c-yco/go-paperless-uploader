# Quick Reference Guide

## Creating a Release

```bash
# Method 1: Git tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# Method 2: GitHub CLI
gh release create v1.0.0 --generate-notes
```

## Running Security Scans Locally

```bash
# Install tools
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest

# Run scans
gosec ./...
govulncheck ./...
```

## Running Linter Locally

```bash
# Install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run
golangci-lint run

# Auto-fix
golangci-lint run --fix
```

## Running Tests

```bash
# All tests
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Docker Commands

```bash
# Build
docker build -t paperless-uploader:latest .

# Run
docker run -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/consume:/app/consume \
  -v $(pwd)/processed:/app/processed \
  paperless-uploader:latest

# Pull from registry (after release)
docker pull ghcr.io/c-yco/go-paperless-uploader:latest
```

## Workflow Status

Check workflow status:
- CI: https://github.com/c-yco/go-paperless-uploader/actions/workflows/ci.yml
- Release: https://github.com/c-yco/go-paperless-uploader/actions/workflows/release.yml

## Security Dashboard

View security findings:
- https://github.com/c-yco/go-paperless-uploader/security

## Releases

View and download releases:
- https://github.com/c-yco/go-paperless-uploader/releases
