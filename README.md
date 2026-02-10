# aws-mfa-go

Minimal, fast AWS MFA helper for the CLI.

[![](https://img.shields.io/github/actions/workflow/status/jlis/aws-mfa-go/ci.yml?branch=master&longCache=true&label=Test&logo=github%20actions&logoColor=fff)](https://github.com/jlis/aws-mfa-go/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/jlis/aws-mfa-go)](https://goreportcard.com/report/github.com/jlis/aws-mfa-go)

This is inspired by (and aims to be compatible with the core workflow of) [`broamski/aws-mfa`](https://github.com/broamski/aws-mfa), but implemented in Go so you can install a single binary.

## What it does

- Reads **long-term** credentials from `~/.aws/credentials`
- Prompts for an MFA token (or accepts `--token`)
- Calls AWS STS `GetSessionToken`
- Writes **short-term** credentials back into `~/.aws/credentials`
- Skips STS calls when existing short-term credentials are still valid (unless `--force`)

## Install

### Homebrew (macOS)

```bash
brew tap jlis/tap
brew install aws-mfa-go
```

### From source

Install the latest version with Go:

```bash
go install github.com/jlis/aws-mfa-go/cmd/aws-mfa-go@latest
aws-mfa-go --help
```

Build from a local checkout:

```bash
go test ./...
go build -o aws-mfa-go ./cmd/aws-mfa-go
./aws-mfa-go --help
```

## Credentials file setup

`aws-mfa-go` follows the same convention as `aws-mfa`:

- Long-term section: `<profile>-long-term` (default)
- Short-term section: `<profile>` (default)

Example (`~/.aws/credentials`):

```ini
[prod-long-term]
aws_access_key_id = YOUR_LONGTERM_KEY_ID
aws_secret_access_key = YOUR_LONGTERM_SECRET
aws_mfa_device = arn:aws:iam::123456789012:mfa/your-user

[prod]
# short-term keys will be written here by aws-mfa-go
```

## Usage

Refresh credentials for a profile:

```bash
aws-mfa-go --profile prod
```

Show version:

```bash
aws-mfa-go --version
```

Non-interactive (CI-ish) usage:

```bash
aws-mfa-go --profile prod --token 123456
```

Force refresh even if credentials are still valid:

```bash
aws-mfa-go --profile prod --force
```

## Configuration precedence

Like upstream `aws-mfa`, precedence is:

**flags > environment variables > `~/.aws/credentials` values > defaults**

Supported environment variables:

- `AWS_PROFILE`
- `MFA_DEVICE`
- `MFA_STS_DURATION`
- `AWS_REGION` / `AWS_DEFAULT_REGION` (region used for STS client; defaults to `us-east-1`)

## Profile suffixes (advanced)

Override how long-term/short-term sections are derived:

- `--long-term-suffix` (default `long-term`, use `none` to use `<profile>` as the long-term section)
- `--short-term-suffix` (default `none`, when set uses `<profile>-<suffix>`)

Example: share one long-term section and mint multiple short-term sections:

```bash
aws-mfa-go --profile myorg --long-term-suffix none --short-term-suffix production
aws-mfa-go --profile myorg --long-term-suffix none --short-term-suffix staging
```

This will write short-term credentials into:

- `[myorg-production]`
- `[myorg-staging]`

…while reading long-term credentials from `[myorg]`.

## Shell completion

Cobra provides a built-in completion command:

```bash
aws-mfa-go completion zsh
aws-mfa-go completion bash
```

## Development

Run tests:

```bash
make test
```

Run linter:

```bash
golangci-lint run
```

Note: this repo builds with Go 1.20+, but CI runs on a supported Go release line for the latest security fixes.

If you install `golangci-lint`, use a recent version that supports your Go version.


