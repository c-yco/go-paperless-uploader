# CI/CD and Release Documentation

This document describes the CI/CD pipelines and release process for the Go Paperless Uploader project.

## Overview

The project uses GitHub Actions for continuous integration, security scanning, and automated releases. Renovate is configured to automatically update dependencies.

## Workflows

### 1. CI Workflow (`.github/workflows/ci.yml`)

Runs on every push and pull request to `main` and `develop` branches.

#### Jobs:

- **Lint**: Runs `golangci-lint` to check code quality
- **Security**: Performs security scanning with:
  - `gosec`: Go security checker
  - `govulncheck`: Go vulnerability database checker
  - `trivy`: Container and filesystem vulnerability scanner
- **Test**: Runs tests on Linux, Windows, and macOS with race detection
- **Build**: Creates binaries for Linux and Windows to verify buildability

#### SARIF Reports:
Security findings are uploaded to GitHub Security tab for easy tracking.

### 2. Release Workflow (`.github/workflows/release.yml`)

Triggers automatically when you push a version tag (e.g., `v1.0.0`) or can be triggered manually.

#### Jobs:

1. **Security Scan**: Pre-release security validation
2. **Test**: Comprehensive test suite with coverage reporting
3. **Build**: Creates binaries for multiple platforms:
   - Linux (amd64, arm64)
   - Windows (amd64)
   - macOS (Intel and Apple Silicon)
4. **Release**: Creates a GitHub release with:
   - All binary artifacts
   - SHA256 checksums
   - Auto-generated changelog
   - Release notes
5. **Docker** (optional): Builds and pushes multi-arch Docker images to GitHub Container Registry

## Creating a Release

### Method 1: Using Git Tags

```bash
# Create and push a version tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

### Method 2: Using GitHub CLI

```bash
gh release create v1.0.0 --generate-notes --title "v1.0.0"
```

### Method 3: Manual Workflow Dispatch

1. Go to Actions tab in GitHub
2. Select "Release" workflow
3. Click "Run workflow"
4. Enter version tag (optional)

## Security Scanning

### Tools Used:

1. **gosec**: Static security analyzer for Go
   - Scans for common security issues
   - Results uploaded as SARIF to GitHub Security

2. **govulncheck**: Official Go vulnerability checker
   - Checks dependencies against the Go vulnerability database
   - Fails build if critical vulnerabilities found

3. **trivy**: Multi-purpose security scanner
   - Scans for vulnerabilities in dependencies
   - Checks for misconfigurations
   - Results uploaded as SARIF

### Viewing Security Results:

Navigate to the **Security** tab in your GitHub repository to view:
- Code scanning alerts
- Dependabot alerts
- Secret scanning alerts

## Dependency Management with Renovate

### Configuration (`renovate.json`)

Renovate automatically:
- Creates PRs for dependency updates
- Groups minor and patch updates
- Auto-merges non-major updates after tests pass
- Schedules updates weekly (Monday mornings)
- Prioritizes security updates

### Setup Instructions:

1. Install the [Renovate GitHub App](https://github.com/apps/renovate) on your repository
2. The configuration in `renovate.json` will be automatically detected
3. Renovate will create a "Dependency Dashboard" issue for tracking

### Update Categories:

- **Go dependencies (non-major)**: Auto-merged after CI passes
- **Go dependencies (major)**: Requires manual review
- **GitHub Actions**: Auto-merged
- **Docker images**: Auto-merged
- **Security updates**: Processed immediately

## Code Quality

### golangci-lint Configuration (`.golangci.yml`)

Enabled linters:
- `errcheck`: Error handling verification
- `gosimple`: Simplification suggestions
- `govet`: Go vet issues
- `gosec`: Security issues
- `gofmt`/`goimports`: Code formatting
- `misspell`: Spelling errors
- And more...

### Running Locally:

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Run with auto-fix
golangci-lint run --fix
```

## Docker Support

### Building Docker Image:

```bash
docker build -t paperless-uploader:latest .
```

### Running Docker Container:

```bash
docker run -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/consume:/app/consume \
  -v $(pwd)/processed:/app/processed \
  paperless-uploader:latest
```

### Multi-arch Images:

The release workflow automatically builds and pushes images for:
- `linux/amd64`
- `linux/arm64`

Images are available at: `ghcr.io/c-yco/go-paperless-uploader`

## Binary Artifacts

### Download Locations:

- **Releases**: Check the [Releases page](https://github.com/c-yco/go-paperless-uploader/releases)
- **CI Artifacts**: Available in Actions tab for 7 days

### Verification:

```bash
# Verify checksums
sha256sum -c paperless-uploader-linux-amd64.sha256

# Make executable (Linux/macOS)
chmod +x paperless-uploader-linux-amd64

# Run
./paperless-uploader-linux-amd64
```

## Environment Variables

Set these in GitHub repository secrets if needed:

- `GITHUB_TOKEN`: Automatically provided by GitHub Actions
- `CODECOV_TOKEN`: Optional, for Codecov integration

## Troubleshooting

### Release Workflow Not Triggering:

- Ensure tag follows semantic versioning: `v*.*.*`
- Check if tag was pushed: `git push --tags`

### Security Scan Failures:

- Check Security tab for details
- Review SARIF files in workflow artifacts
- Update vulnerable dependencies

### Build Failures:

- Check if code compiles locally for target platforms
- Verify Go version compatibility
- Review error logs in Actions tab

## Best Practices

1. **Always run tests locally** before pushing:
   ```bash
   go test -race ./...
   ```

2. **Run linter before committing**:
   ```bash
   golangci-lint run
   ```

3. **Check for vulnerabilities**:
   ```bash
   govulncheck ./...
   ```

4. **Review Renovate PRs promptly** to keep dependencies up-to-date

5. **Use semantic versioning** for releases:
   - MAJOR: Breaking changes
   - MINOR: New features (backward compatible)
   - PATCH: Bug fixes

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Renovate Documentation](https://docs.renovatebot.com/)
- [golangci-lint Documentation](https://golangci-lint.run/)
- [gosec Documentation](https://github.com/securego/gosec)
- [govulncheck Documentation](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)
