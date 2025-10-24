# Troubleshooting Guide

## Common Issues and Solutions

### gosec SARIF Upload Issues

**Issue**: `Unable to upload "gosec-results.sarif" as it is not valid SARIF`

**Explanation**: gosec's SARIF output format sometimes includes fields that are not fully compatible with GitHub's SARIF specification validator, particularly the `fixes[0].artifactChanges` field format.

**Solutions Implemented**:

1. **Continue on Error**: The workflows use `continue-on-error: true` so that SARIF upload failures don't block the pipeline
2. **Multiple Formats**: gosec runs in both JSON and text formats, with results displayed in workflow logs
3. **Artifact Upload**: JSON results are uploaded as workflow artifacts for manual review
4. **Trivy Backup**: The CI workflow also runs Trivy security scanner as a backup

**How to View Security Results**:

1. **Workflow Logs**: Check the "Run gosec Security Scanner" step output in the Actions tab
2. **Artifacts**: Download the `gosec-results.json` artifact from the workflow run
3. **Trivy Results**: Check the Security tab for Trivy SARIF uploads (CI workflow)
4. **Local Scan**: Run `gosec ./...` locally to see results

**Running gosec Locally**:

```bash
# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run with text output
gosec ./...

# Run with JSON output for detailed analysis
gosec -fmt json -out gosec-results.json ./...

# Run with detailed severity levels
gosec -severity medium ./...

# Exclude specific issues
gosec -exclude=G104 ./...
```

### govulncheck Failures

**Issue**: Build fails due to known vulnerabilities in dependencies

**Solution**:
```bash
# Check locally
govulncheck ./...

# Update vulnerable dependencies
go get -u <vulnerable-package>
go mod tidy

# If vulnerability is in indirect dependency
go get -u all
go mod tidy
```

### Trivy Scan Failures

**Issue**: Trivy detects vulnerabilities in dependencies or Docker images

**Solution**:
1. Review the Security tab for details
2. Update affected dependencies
3. Check for available patches
4. If no fix available, consider risk acceptance with documentation

### Cross-Platform Build Failures

**Issue**: Code builds on one platform but fails on another

**Common Causes**:
- Platform-specific imports (e.g., `golang.org/x/sys/unix` vs `windows`)
- File path handling differences
- Line ending issues (CRLF vs LF)

**Solution**:
```bash
# Test builds locally for all platforms
GOOS=linux GOARCH=amd64 go build ./cmd/paperless-uploader
GOOS=windows GOARCH=amd64 go build ./cmd/paperless-uploader
GOOS=darwin GOARCH=amd64 go build ./cmd/paperless-uploader

# Use build tags for platform-specific code
// +build windows
// +build !windows
```

### Docker Build Failures

**Issue**: Docker image build fails in GitHub Actions

**Solutions**:
1. Test locally: `docker build -t test .`
2. Check `.dockerignore` isn't excluding required files
3. Verify base image is available
4. Check multi-arch build compatibility

### Renovate Not Creating PRs

**Issue**: Renovate bot is not creating pull requests

**Checklist**:
1. ✅ Renovate GitHub App is installed on the repository
2. ✅ `renovate.json` is in the repository root
3. ✅ Check Renovate dashboard issue for errors
4. ✅ Verify branch protection rules aren't blocking Renovate

**Debug**:
- Check the Dependency Dashboard issue for Renovate logs
- Validate `renovate.json` syntax at https://www.schemastore.org/json/

### Release Workflow Not Triggering

**Issue**: Pushing a tag doesn't trigger the release workflow

**Checklist**:
1. Tag format must match `v*.*.*` (e.g., `v1.0.0`, not `1.0.0`)
2. Ensure workflows are enabled in Settings → Actions
3. Check if the workflow file has syntax errors

**Commands**:
```bash
# Verify tag format
git tag -l

# Delete and recreate tag if needed
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### CI Tests Failing

**Issue**: Tests pass locally but fail in CI

**Common Causes**:
1. **Race conditions**: Use `go test -race` locally
2. **Timezone differences**: Use UTC in tests
3. **File permissions**: Check file creation/access
4. **Path separators**: Use `filepath.Join()` instead of hardcoded `/` or `\`

**Debug**:
```bash
# Run tests with verbose output
go test -v ./...

# Run tests with race detector
go test -race ./...

# Run specific test
go test -v -run TestName ./...

# Check for parallelism issues
go test -parallel 1 ./...
```

### golangci-lint Failures

**Issue**: Linter reports issues in CI but not locally

**Solution**:
```bash
# Install same version as CI
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run with same config
golangci-lint run

# Auto-fix issues
golangci-lint run --fix

# Run specific linters
golangci-lint run --disable-all --enable=gosec
```

### Secrets and Environment Variables

**Issue**: Workflow fails due to missing secrets

**Required Secrets** (if using):
- `GITHUB_TOKEN` - Automatically provided by GitHub
- `CODECOV_TOKEN` - Only needed if using Codecov (optional)

**Setting Secrets**:
1. Go to Settings → Secrets and variables → Actions
2. Click "New repository secret"
3. Add name and value

### Artifact Upload/Download Issues

**Issue**: Artifacts not available or corrupted

**Solutions**:
1. Check artifact retention period (default: 7 days for CI, 90 for releases)
2. Verify file paths are correct in workflow
3. Ensure files are created before upload step

### Performance Issues

**Issue**: Workflows taking too long

**Optimizations**:
1. Enable Go module caching (already configured)
2. Use `GOOS` and `GOARCH` matrix sparingly
3. Run linting and security scans in parallel
4. Consider caching Docker layers

### GitHub Security Tab Not Showing Results

**Issue**: Security scans run but results don't appear

**Checklist**:
1. ✅ SARIF files are being uploaded (check workflow logs)
2. ✅ Repository has "Code scanning alerts" enabled
3. ✅ You have appropriate permissions to view security tab
4. ✅ Results may take a few minutes to appear

**Alternative**:
- Download JSON artifacts from workflow runs
- View results in workflow logs

## Getting Help

If issues persist:

1. **Check Workflow Logs**: Actions tab → Select workflow run → Click on failed job
2. **Review Security Tab**: Security → Code scanning alerts
3. **Dependency Dashboard**: Check Renovate's dashboard issue
4. **Local Testing**: Run commands locally to reproduce
5. **GitHub Actions Documentation**: https://docs.github.com/en/actions

## Useful Commands

```bash
# Validate workflow YAML syntax
cat .github/workflows/ci.yml | docker run --rm -i ghcr.io/rhysd/actionlint:latest -

# Test Docker build
docker build --no-cache -t test .

# Clean Go cache
go clean -cache -modcache -testcache

# Update all dependencies
go get -u all && go mod tidy

# Check for outdated dependencies
go list -u -m all
```
