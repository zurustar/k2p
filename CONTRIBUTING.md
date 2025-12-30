# Contributing to k2p

Thank you for your interest in contributing to k2p! This document provides guidelines for developing, testing, and submitting changes.

## Development Workflow

We strictly follow a structured workflow to maintain code quality and documentation consistency.

### 1. Mandatory Pre-Work
Before starting ANY implementation or bug fix, you **MUST** run:

```bash
/start-work
```

This ensures you read `.agent/workflows/start-work.md`. **Documentation updates (tasks.md, design.md) must typically happen BEFORE code changes.**

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
1. Ensure `docs/tasks.md` is fully updated.
2. Verify all tests pass.
3. Update version in `cmd/k2p/main.go` (or via build tag).
4. Tag the release in Git.

## AI Agents
If you are an AI agent working on this repo:
- **ALWAYS** check `docs/tasks.md` first.
- **ALWAYS** update `docs/tasks.md` as you progress.
- **NEVER** skip the `/start-work` workflow.
- **Clean up** any test images/artifacts you generate immediately after verification.
