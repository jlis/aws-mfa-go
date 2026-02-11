# aws-mfa-go

Minimal AWS MFA helper for the CLI.

[![](https://img.shields.io/github/actions/workflow/status/jlis/aws-mfa-go/ci.yml?branch=master&longCache=true&label=Test&logo=github%20actions&logoColor=fff)](https://github.com/jlis/aws-mfa-go/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/jlis/aws-mfa-go)](https://goreportcard.com/report/github.com/jlis/aws-mfa-go)

`aws-mfa-go` refreshes temporary AWS credentials using MFA and writes them to `~/.aws/credentials`.

Inspired by [`broamski/aws-mfa`](https://github.com/broamski/aws-mfa), implemented in Go as a single binary.

## Quick start (2 minutes)

1) Install:

```bash
brew install jlis/tap/aws-mfa-go
```

2) Add long-term credentials to `~/.aws/credentials`:

```ini
[prod-long-term]
aws_access_key_id = YOUR_LONGTERM_KEY_ID
aws_secret_access_key = YOUR_LONGTERM_SECRET
aws_mfa_device = arn:aws:iam::123456789012:mfa/your-user
```

3) Run:

```bash
aws-mfa-go --profile prod
```

You’ll be prompted for your 6-digit MFA code, and short-term credentials will be written to `[prod]`.

## What it does

- Reads **long-term** credentials from `~/.aws/credentials`
- Prompts for an MFA token (or accepts `--token`)
- Calls AWS STS `GetSessionToken`
- Writes **short-term** credentials back into `~/.aws/credentials`
- Skips STS calls when existing short-term credentials are still valid (unless `--force`)

## Install

### Homebrew (macOS)

```bash
brew install jlis/tap/aws-mfa-go
```

### Upgrade (Homebrew)

```bash
brew update
brew upgrade aws-mfa-go
```

### Alternative: install with Go

```bash
go install github.com/jlis/aws-mfa-go/cmd/aws-mfa-go@latest
```

## Minimal configuration

```ini
[prod-long-term]
aws_access_key_id = YOUR_LONGTERM_KEY_ID
aws_secret_access_key = YOUR_LONGTERM_SECRET
aws_mfa_device = arn:aws:iam::123456789012:mfa/your-user
```

Required keys in your long-term section:
- `aws_access_key_id`
- `aws_secret_access_key`
- `aws_mfa_device`

Short-term credentials are written automatically to `[<profile>]` (for example `[prod]`).

## Common usage

Refresh credentials:

```bash
aws-mfa-go --profile prod
```

Force refresh:

```bash
aws-mfa-go --profile prod --force
```

Non-interactive:

```bash
aws-mfa-go --profile prod --token 123456
```

Show version:

```bash
aws-mfa-go --version
```

## Configuration precedence

`aws-mfa-go` uses:

**flags > environment variables > `~/.aws/credentials` values > defaults**

Environment variables:
- `AWS_PROFILE`
- `MFA_DEVICE`
- `MFA_STS_DURATION`
- `AWS_REGION` / `AWS_DEFAULT_REGION` (defaults to `us-east-1`)

## Advanced profile suffixes

By default:
- long-term section: `<profile>-long-term`
- short-term section: `<profile>`

Override this with:
- `--long-term-suffix` (default `long-term`, use `none` to use `<profile>`)
- `--short-term-suffix` (default `none`, when set uses `<profile>-<suffix>`)

Example:

```bash
aws-mfa-go --profile myorg --long-term-suffix none --short-term-suffix production
aws-mfa-go --profile myorg --long-term-suffix none --short-term-suffix staging
```

This writes short-term credentials to `[myorg-production]` and `[myorg-staging]`, while reading long-term credentials from `[myorg]`.

## Development

Build from local checkout:

```bash
make build
./aws-mfa-go --help
```

Run tests:

```bash
make test
```

Run linter:

```bash
make lint
```

Note: this repo builds with Go 1.20+, but CI runs on a supported Go release line for the latest security fixes.


