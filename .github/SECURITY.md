# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < latest| :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in git-retime, please report it responsibly.

**Please do not open a public issue for security vulnerabilities.**

Instead:

1. **Email** the maintainers with details of the vulnerability
2. Include steps to reproduce and any proof-of-concept if available
3. Allow a reasonable time for a fix before any public disclosure

We will acknowledge your report and work on a fix. We appreciate your help in keeping git-retime and its users safe.

## Security Considerations

git-retime operates on your local Git repository and does not transmit data over the network. The main security considerations are:

- **Repository integrity**: The tool rewrites commit history. Ensure you understand the implications before running it on shared branches.
- **Editor invocation**: git-retime opens your configured `$GIT_EDITOR`. Ensure your editor and environment are trusted.
