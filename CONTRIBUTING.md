# Contributing to HomeClip

Thank you for your interest in contributing to HomeClip! This document provides guidelines for contributing to the
project.

## Commit Guidelines

Write clear, descriptive commit messages that explain what changed and why. While we don't enforce strict commit conventions, clear communication helps maintain the project.

## Pull Request Process

1. **Fork the repository** and create your branch from `main`
2. **Update documentation** if you're changing functionality
3. **Add tests** for new features
4. **Ensure all tests pass** with `make test`
5. **Update the README** if needed
6. **Submit a pull request** with a clear description

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

Releases are automated:

1. Merge PR to `main`
2. GitHub Actions automatically:
   - Increments version number
   - Creates a Git tag
   - Publishes GitHub Release
   - Builds and pushes Docker images to GHCR

Every merge to main creates a new release with Docker images!

## Questions?

Feel free to open an issue for any questions about contributing.