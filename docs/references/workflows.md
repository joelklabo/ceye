# GitHub Actions Workflows

This document describes all GitHub Actions workflows in the ceye project and how they work together to ensure code quality, security, and reliability.

## Workflow Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Code Push / PR                           │
└────────────────┬────────────────────────────────────────────────┘
                 │
    ┌────────────┼────────────┬────────────┬────────────┐
    │            │            │            │            │
    ▼            ▼            ▼            ▼            ▼
┌────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
│   CI   │ │  Tests   │ │ Security │ │ Quality  │ │   Docs   │
│ Basic  │ │ Complete │ │  Scans   │ │  Checks  │ │ Validate │
└────┬───┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘
     │          │            │            │            │
     └──────────┴────────────┴────────────┴────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │  All Checks   │
                    │    Pass?      │
                    └───────┬───────┘
                            │
                ┌───────────┴───────────┐
                │                       │
                ▼                       ▼
           ┌─────────┐           ┌──────────┐
           │  Merge  │           │  Block   │
           │   PR    │           │   PR     │
           └────┬────┘           └──────────┘
                │
                ▼
         ┌─────────────┐
         │  Tag: v1.0  │
         └──────┬──────┘
                │
                ▼
         ┌─────────────┐
         │   Release   │
         │  (GoReleaser)│
         └──────┬──────┘
                │
                ▼
         ┌─────────────┐
         │  Artifacts  │
         │  Published  │
         └─────────────┘
```

## Core Workflows

### 1. CI Workflow (`ci.yml`)

**Purpose:** Fast feedback loop for basic checks

**Triggers:**
- Push to `main` branch
- Pull requests to `main`

**Jobs:**
- **test** - Run all tests with race detector and coverage
- **lint** - Run golangci-lint for code quality
- **build** - Cross-compile for multiple OS/arch combinations

**Key Features:**
- Go module caching for speed
- Race detector enabled
- Coverage uploaded to Codecov
- Builds 6 platform combinations (linux/darwin/windows × amd64/arm64)
- Artifacts uploaded for each build

**Why:** Provides quick feedback (~3-5 minutes) on basic functionality before running more expensive tests.

### 2. Comprehensive Tests (`tests.yml`)

**Purpose:** Deep testing across all dimensions

**Triggers:**
- Push to `main` branch
- Pull requests to `main`
- Daily at 2am UTC (scheduled)

**Jobs:**
- **unit-tests** - Test on Go 1.21, 1.22, 1.23 (matrix)
- **integration-tests** - Test provider integration and demo mode
- **provider-tests** - Contract tests for all providers
- **alerting-tests** - Alert condition and notification tests
- **ui-tests** - TUI rendering and web server tests
- **storage-tests** - SQLite and database operations
- **config-tests** - Configuration loading and validation
- **cross-platform** - Build and test on Linux, macOS, Windows
- **benchmark** - Performance benchmarks
- **test-summary** - Aggregate results from all jobs

**Key Features:**
- Multi-version Go testing
- Provider contract validation
- UI snapshot capture (uploaded as artifact)
- Web server health checks
- Performance benchmarking
- Platform-specific testing

**Why:** Ensures comprehensive coverage across all code paths, providers, and platforms.

### 3. Security Checks (`security.yml`)

**Purpose:** Identify security vulnerabilities and secrets

**Triggers:**
- Push to `main` branch
- Pull requests to `main`
- Weekly on Mondays at 9am UTC

**Jobs:**
- **gosec** - Static security scanning with gosec
- **govulncheck** - Check for known vulnerabilities in dependencies
- **dependency-review** - Review dependency changes in PRs
- **trivy** - Filesystem security scanning
- **secrets-scan** - Detect leaked secrets with TruffleHog

**Key Features:**
- SARIF output for GitHub Security tab
- Automated dependency review on PRs
- Secret detection in commit history
- Weekly scheduled scans

**Why:** Proactive security prevents vulnerabilities from reaching production.

### 4. Code Quality (`code-quality.yml`)

**Purpose:** Maintain code quality standards

**Triggers:**
- Push to `main` branch
- Pull requests to `main`

**Jobs:**
- **golangci-lint** - Comprehensive linting (40+ linters)
- **staticcheck** - Advanced static analysis
- **go-fmt** - Verify code formatting
- **go-vet** - Standard Go checks
- **deadcode** - Find unused code
- **coverage-report** - Generate and enforce coverage (>70%)
- **test-coverage-comment** - Post coverage on PRs

**Key Features:**
- 10-minute timeout for thorough linting
- HTML coverage report (uploaded as artifact)
- Automatic PR comments with coverage
- Fails if coverage below 70%

**Why:** Consistent code quality makes maintenance easier and prevents bugs.

### 5. Documentation (`docs.yml`)

**Purpose:** Validate and maintain documentation

**Triggers:**
- Push to `main` branch
- Pull requests to `main`

**Jobs:**
- **validate-docs** - Check links, structure, symlinks
- **generate-api-docs** - Generate godoc HTML
- **spell-check** - Check spelling with typos
- **update-ui-snapshot** - Capture TUI demo (on push to main)

**Key Features:**
- Broken link detection
- Required file validation
- Symlink verification
- Automatic UI snapshot updates
- API documentation generation
- Spell checking

**Why:** Good documentation is critical for adoption and maintenance.

### 6. Release (`release.yml`)

**Purpose:** Automated release process

**Triggers:**
- Push of tags matching `v*` (e.g., `v1.0.0`)

**Jobs:**
- **release** - Build and publish release with GoReleaser

**Key Features:**
- Multi-platform binaries
- Automated changelog
- GitHub Release creation
- Checksum generation
- Archive creation (.tar.gz, .zip)

**Configuration:** Uses `.goreleaser.yml` for build configuration

**Why:** Consistent, automated releases reduce errors and save time.

### 7. Snapshot Demo (`snapshot.yml`)

**Purpose:** Capture TUI demo for documentation

**Triggers:**
- Push to `main` branch
- Pull requests

**Jobs:**
- **snapshot** - Build, run demo, capture output

**Key Features:**
- Uses tmux for capture
- Runs tests first
- Uploads snapshot as artifact
- Stored in `docs/ui-demo.txt`

**Why:** Visual documentation helps users understand the UI.

### 8. Workflow Health (`workflow-health.yml`)

**Purpose:** Self-check for workflow configuration

**Triggers:**
- Changes to `.github/workflows/**`
- Daily at 3am UTC

**Jobs:**
- **validate-workflows** - YAML syntax and required files
- **test-workflow-triggers** - Verify trigger configuration
- **workflow-documentation** - Generate workflow summary
- **workflow-metrics** - Count and report metrics

**Key Features:**
- Self-validating workflows
- Required workflow enforcement
- Deprecated action detection
- Automatic documentation generation
- Workflow metrics dashboard

**Why:** Ensures CI/CD infrastructure stays healthy and up-to-date.

## Workflow Dependencies

```
CI (fast) ────────────┐
                      ├──▶ Required for merge
Tests (thorough) ─────┤
                      │
Security (safety) ────┤
                      │
Quality (standards) ──┤
                      │
Docs (validate) ──────┘

Release (on tag) ──▶ Independent, requires passing tests

Health (self-check) ──▶ Validates other workflows
```

## Coverage and Protection

### Branch Protection Rules

**Recommended settings for `main` branch:**

```yaml
required_status_checks:
  strict: true
  contexts:
    - "Test"               # From ci.yml
    - "Lint"               # From ci.yml
    - "Build"              # From ci.yml
    - "Unit Tests"         # From tests.yml
    - "Integration Tests"  # From tests.yml
    - "Security Scan"      # From security.yml
    - "Go Linting"         # From code-quality.yml
    - "Validate Documentation" # From docs.yml

required_pull_request_reviews:
  required_approving_review_count: 1

enforce_admins: true
```

### Test Coverage Breakdown

| Component | Tested By | Coverage |
|-----------|-----------|----------|
| Core types | `tests.yml` (unit-tests) | 95%+ |
| Providers | `tests.yml` (provider-tests) | 90%+ |
| Store | `tests.yml` (unit-tests) | 95%+ |
| Alerting | `tests.yml` (alerting-tests) | 85%+ |
| UI (TUI) | `tests.yml` (ui-tests) | 75%+ |
| Web Server | `tests.yml` (ui-tests) | 80%+ |
| Storage | `tests.yml` (storage-tests) | 90%+ |
| Config | `tests.yml` (config-tests) | 85%+ |
| **Overall** | All workflows combined | **80%+** |

### Security Scanning

| Tool | Purpose | Coverage |
|------|---------|----------|
| gosec | Static analysis | Go code |
| govulncheck | Vulnerability DB | Dependencies |
| Trivy | Filesystem scan | All files |
| TruffleHog | Secret detection | Git history |
| Dependency Review | PR diff analysis | go.mod changes |

## Performance Characteristics

| Workflow | Typical Duration | Cost (compute minutes) |
|----------|------------------|----------------------|
| CI | 3-5 minutes | 3-5 |
| Tests | 15-20 minutes | 40-60 (matrix jobs) |
| Security | 5-10 minutes | 10-15 |
| Code Quality | 5-8 minutes | 10-15 |
| Docs | 2-3 minutes | 2-3 |
| Release | 5-10 minutes | 10-15 |
| Snapshot | 1-2 minutes | 1-2 |
| Workflow Health | 1-2 minutes | 1-2 |
| **Total per PR** | **~30-40 min** | **~70-120 min** |

**Optimization strategies:**
- Aggressive caching (Go modules)
- Matrix job parallelization
- Fast-fail on basic checks
- Conditional job execution
- Scheduled expensive jobs (daily/weekly)

## Scheduled Jobs

| Workflow | Schedule | Purpose |
|----------|----------|---------|
| Tests | Daily 2am UTC | Catch environment drift |
| Security | Weekly Mon 9am | Deep security scan |
| Workflow Health | Daily 3am UTC | Self-check |

## Artifacts Generated

| Workflow | Artifact | Purpose | Retention |
|----------|----------|---------|-----------|
| CI | Platform binaries | Cross-platform builds | 90 days |
| Tests | TUI snapshot | UI documentation | 90 days |
| Tests | Benchmark results | Performance tracking | 90 days |
| Quality | Coverage HTML | Coverage report | 30 days |
| Docs | API documentation | godoc HTML | 30 days |
| Docs | Workflow summary | Workflow inventory | 30 days |
| Release | Release binaries | Public downloads | Permanent |

## Adding a New Workflow

1. **Create workflow file** in `.github/workflows/`
2. **Add required triggers** (push, pull_request at minimum)
3. **Use latest action versions** (v4 for checkout, v5 for setup-go)
4. **Add caching** for Go modules
5. **Upload artifacts** if generating reports
6. **Update this document** with workflow details
7. **Add to branch protection** if critical
8. **Test locally** with `act` (GitHub Actions runner)

## Workflow Best Practices

✅ **DO:**
- Use semantic workflow names
- Cache dependencies aggressively
- Fail fast for basic checks
- Upload artifacts for debugging
- Use matrix jobs for parallelization
- Set appropriate timeouts
- Use SARIF for security findings
- Comment coverage on PRs

❌ **DON'T:**
- Run expensive jobs on every commit
- Forget to handle errors
- Hard-code secrets in workflows
- Use deprecated action versions
- Skip caching
- Run long jobs sequentially
- Ignore security warnings

## Monitoring Workflow Health

**Check workflow status:**
```bash
gh workflow list
gh run list --workflow=tests.yml
gh run view <run-id>
```

**View workflow logs:**
```bash
gh run view <run-id> --log
gh run view <run-id> --log-failed
```

**Re-run failed workflows:**
```bash
gh run rerun <run-id>
gh run rerun <run-id> --failed
```

## Troubleshooting

### Common Issues

**Workflow not triggering:**
- Check trigger configuration in YAML
- Verify branch name matches
- Check if workflow is disabled
- Review GitHub Actions logs

**Slow workflows:**
- Add/improve caching
- Parallelize with matrix jobs
- Move expensive jobs to scheduled runs
- Profile job duration

**Flaky tests:**
- Increase timeouts
- Add retry logic
- Check for race conditions
- Review test logs in artifacts

**High compute costs:**
- Optimize caching
- Reduce matrix dimensions
- Skip redundant jobs
- Use scheduled runs for expensive checks

## Future Improvements

- [ ] Add E2E tests with Playwright
- [ ] Performance regression detection
- [ ] Automatic dependency updates (Dependabot)
- [ ] Nightly builds
- [ ] Docker image builds
- [ ] Homebrew formula updates
- [ ] Changelog generation
- [ ] Release notes automation
- [ ] Deployment to demo environment
- [ ] Load testing in CI

---

**Last Updated:** 2025-11-17  
**Maintained By:** @joelklabo  
**Questions?** Open an issue or discussion on GitHub
