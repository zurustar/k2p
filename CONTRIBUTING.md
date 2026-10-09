# Contributing to k2p

Thank you for your interest in contributing to k2p! This document provides guidelines for developing, testing, and submitting changes.

## Development Workflow

### 1. Development Process
Follow [.agent/workflows/start-work.md](.agent/workflows/start-work.md): update docs first, write tests first, then implement.

### 2. Git Branching
We follow the **Git Flow** strategy defined in `.agent/workflows/git-flow.md`.

- **Main Branch**: `main` (stable)
- **Feature Branches**: `feature/your-feature-name`
- **Bug Fixes**: `fix/issue-description`

**Rule**: Never commit directly to `main`. Always create a branch.

### 3. Project Structure
- `cmd/k2p-gui`: Main application entry point (GUI).
- `internal/`: Private library code (File Manager, Automation, PDF Gen, etc.).
- `docs/`: Project documentation.
- `test/`: Integration and E2E tests.

## Building and Testing

### Prerequisites
- Go 1.21+
- Make
- fyne CLI (used by `make build` to create the `.app` bundle):
  ```bash
  go install fyne.io/tools/cmd/fyne@v1.7.0
  ```
  Note: the legacy `fyne.io/fyne/v2/cmd/fyne` does not support the `--app-id` flag used by the Makefile.

### Build
```bash
make build
```

### Run Tests
```bash
# Run unit tests
make test-unit

# Run integration tests (requires MacOS)
make test-integration

# Run all tests
make test
```

## Code Style
- Follow standard Go idioms.
- Run `go fmt` before committing.
- Ensure all exported functions have GoDoc comments.

### Test Artifacts & Cleanup
- Any temporary files (especially images) generated during testing **MUST** be cleaned up automatically.
- Use `defer os.Remove(...)`, `defer os.RemoveAll(...)`, or `t.Cleanup(...)` in Go tests.
- **Do not** leave debug artifacts (like `.png` files) in the repository.

## Release Process
1. Ensure completed tasks have been removed from `docs/tasks.md`.
2. Verify all tests pass.
3. Push a `v*` tag on `main` (e.g. `v0.1.0`). The Release workflow builds the `.app` and publishes it to GitHub Releases.
