# Security Policy

## Reporting a Vulnerability

If you believe you have found a security issue in `infakt-go-sdk`, please report it **privately**. Do not open a public GitHub issue or pull request describing the vulnerability.

Preferred channel:

- **GitHub Security Advisories** — open a private report via the repository's *Security* tab (`Security` → `Report a vulnerability`). This keeps the discussion private until a fix is ready and lets us coordinate a CVE if appropriate.

If you would rather contact the maintainer directly, see the repository metadata (commit history, `go.mod`, GitHub profile) for current contact details.

When you report, please include:

- A description of the issue and its impact.
- Steps to reproduce, ideally with a minimal Go program.
- The version (commit SHA or tag) where you observed the issue.
- Any suggested mitigation, if you have one.

We will acknowledge receipt within a reasonable time, work with you on a fix, and credit you in the release notes unless you prefer to remain anonymous.

## Scope

This policy covers vulnerabilities in **this SDK's source code**: the Go client implementation, its handling of credentials, and its interaction with the inFakt HTTP API.

Out of scope:

- Vulnerabilities or outages in the inFakt service itself (`api.infakt.pl`, the web app, account security, billing, etc.). Please report those directly to inFakt support.
- Issues in third-party Go modules — please report those to their respective maintainers. (This SDK has zero external runtime dependencies, so this is unlikely to apply.)

## Supported Versions

While this module is pre-1.0, only the latest minor release of the `v0.x` line receives security fixes. Older minor versions are not patched; please upgrade to the most recent release. Once `v1.0.0` is published this policy will be updated.

## Handling API Keys

The inFakt API key is the credential that protects your inFakt account. Treat it like a password.

- **Never commit API keys** to version control. Use environment variables, a secrets manager, or a per-developer config file that is git-ignored.
- **Scope keys to least privilege** if inFakt offers per-key permissions or environments — production keys should not be reused for development.
- **Rotate immediately** if a key may have leaked (committed by accident, shared in chat or logs, exposed in a screenshot, etc.). Rotation in the inFakt dashboard invalidates the old key.
- **Avoid logging full request headers**. The SDK itself does not log the API key, but if you wrap the `*http.Client` with logging or tracing middleware, make sure the `X-inFakt-ApiKey` header is redacted before it is written to logs, traces, or error reports.
- Use `WithHTTPClient` to inject a TLS-configured `*http.Client` if your environment requires custom TLS settings; do not transmit the key over plain HTTP.

If you discover a key has been exposed in this repository's history, please report it through the channels above so we can coordinate revocation and history rewriting if needed.
