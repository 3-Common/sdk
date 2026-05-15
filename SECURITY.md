# Security policy

## Reporting a vulnerability

Please report security vulnerabilities **privately** — do not open a public GitHub issue.

Email: **support@3common.com**

Include:

- A description of the issue and its potential impact.
- Steps to reproduce, or a proof-of-concept.
- Affected SDK(s) and version(s) (`@3common/sdk`, `threecommon`, or `github.com/3-Common/sdk/sdk-go`).
- Your contact info if you would like credit in the advisory.

We aim to acknowledge reports within **2 business days** and to issue a fix or mitigation within **30 days** for confirmed vulnerabilities. Critical issues are prioritized.

## Scope

In scope:

- Vulnerabilities in code published from this repository.
- Issues that allow API key leakage, request tampering, or response forgery against a user of the SDK.

Out of scope:

- Issues in the 3Common API server itself — please report those through 3Common support.
- Issues in third-party dependencies that have not been disclosed publicly. We track dependency advisories via Dependabot and patch as they land upstream.

## Supported versions

The latest minor release of each SDK is supported. Older minors receive critical-severity patches at maintainer discretion.

## Handling

We use [GitHub Security Advisories](https://docs.github.com/en/code-security/security-advisories) to coordinate disclosure. Once a fix lands, we publish a CVE and credit the reporter (with permission).
