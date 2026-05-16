# Security Policy

## Supported Versions

| Version | Supported |
|---|---|
| `main` (latest) | ✓ |
| Older releases | ✗ |

Only the latest version on `main` receives security fixes. Older releases are not patched.

---

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

To report a vulnerability, email directly:

**iqbaleff214@gmail.com**

Include the following in your report:
- A description of the vulnerability and its potential impact
- Steps to reproduce or a proof-of-concept (if safe to share)
- Affected component (endpoint, package, configuration)
- Your suggested fix, if any

### Response Timeline

| Stage | Target |
|---|---|
| Acknowledgement | Within 72 hours |
| Initial assessment | Within 7 days |
| Fix or mitigation | Within 30 days (critical issues may be faster) |
| Public disclosure | After fix is deployed, coordinated with reporter |

We follow responsible disclosure. You will be credited in the release notes unless you prefer anonymity.

---

## Scope

The following are in scope:

- Authentication and authorization bypass
- SQL injection or query manipulation
- JWT vulnerabilities (forging, replay, algorithm confusion)
- Sensitive data exposure (user credentials, tokens)
- Rate limiting bypass
- Server-side request forgery (SSRF) via OpenRouter integration

The following are out of scope:

- Vulnerabilities in third-party dependencies (report directly to maintainers)
- Denial-of-service attacks requiring exceptional resources
- Issues in unreleased or experimental branches
