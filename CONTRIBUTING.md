# Contributing to HomeClip

Thank you for your interest in contributing to HomeClip! This document provides guidelines for contributing to the
project.

## Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/) for our commit messages. This enables automatic
versioning and changelog generation.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: A new feature (triggers minor version bump)
- `fix`: A bug fix (triggers patch version bump)
- `docs`: Documentation only changes
- `style`: Changes that don't affect code meaning (white-space, formatting)
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `test`: Adding or updating tests
- `build`: Changes to build system or dependencies
- `ci`: Changes to CI configuration files and scripts
- `chore`: Other changes that don't modify src or test files
- `revert`: Reverts a previous commit

### Breaking Changes

Add `BREAKING CHANGE:` in the commit footer or append `!` after the type/scope to trigger a major version bump.

```
feat!: remove support for Go 1.20

BREAKING CHANGE: Minimum Go version is now 1.24.6
```

### Examples

```
feat(api): add file upload endpoint

Implements multipart form data handling for file uploads
with configurable size limits.

Closes #123
```

```
fix(memory): prevent concurrent map access

Add mutex locks to protect document map from
concurrent read/write operations.
```

```
docs(readme): update Home Assistant integration examples

Simplify the configuration examples to show only
essential read/write operations.
```

## Pull Request Process

1. **Fork the repository** and create your branch from `main`
2. **Follow conventional commits** for all your commits
3. **Update documentation** if you're changing functionality
4. **Add tests** for new features
5. **Ensure all tests pass** with `make test`
6. **Update the README** if needed
7. **Submit a pull request** with a clear description

### PR Title

Pull request titles should also follow the conventional commit format, as they're used for squash merges:

- ✅ `feat: add Docker health check endpoint`
- ✅ `fix: resolve memory leak in file handler`
- ❌ `Added new feature`
- ❌ `bugfix`

## Development Setup

1. Install Go 1.24.6 or later
2. Clone the repository
3. Install development tools:
   ```bash
   go get -tool go.uber.org/mock/mockgen@latest
   ```
4. Generate mocks:
   ```bash
   go generate ./...
   ```
5. Run tests:
   ```bash
   make test
   ```

## Testing

- Write table-driven tests when possible
- Ensure all new code has test coverage
- Use mocks for external dependencies
- Run `make test` before submitting PR

## Code Style

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Keep functions small and focused
- Document exported functions and types
- No unnecessary comments in code

## Release Process

Releases are automated using semantic-release:

1. Merge PR to `main` with conventional commit message
2. GitHub Actions runs semantic-release
3. Version is bumped based on commit types
4. Changelog is generated automatically
5. Git tag is created
6. GitHub Release is published
7. Docker images are built and pushed to GHCR

The process is fully automated, just follow conventional commits!

## Questions?

Feel free to open an issue for any questions about contributing.