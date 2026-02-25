# Contributing to git-retime

Thank you for your interest in contributing to git-retime! This document provides guidelines for contributing.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally
3. **Create a branch** for your changes: `git checkout -b feature/your-feature` or `git checkout -b fix/your-fix`
4. **Make your changes** and add tests where appropriate
5. **Run the test suite**: `go test ./...`
6. **Commit** with clear, descriptive messages
7. **Push** to your fork and open a pull request

## Development Setup

```bash
git clone https://github.com/erfnzdeh/git-retime.git
cd git-retime
go build
go test ./...
```

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep functions focused and reasonably sized
- Add tests for new behavior and bug fixes

## Pull Request Process

1. Ensure all tests pass
2. Update documentation if you change behavior or add features
3. Fill out the pull request template
4. Request review from maintainers

## Reporting Bugs

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md) when opening an issue. Include:

- Your OS and Go version
- Steps to reproduce
- Expected vs actual behavior

## Suggesting Features

Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md) for new ideas. Describe the use case and proposed behavior.

## License

By contributing, you agree that your contributions will be licensed under the GNU General Public License v3.0.
