# Gardening Assistant Backend

This repository contains a lightweight HTTP service that proxies chat requests to OpenAI's APIs and provides a few helper utilities (e.g. token counting).

## Running tests and quality checks

The CI pipeline runs the commands below. Run them locally before pushing changes to keep development and CI aligned:

```bash
# Run unit and integration tests
go test ./...

# Run vet and lint checks
go vet ./...
golangci-lint run

# Ensure source files are formatted and imports are tidy
gofmt -l . | tee /tmp/gofmt.out
[ ! -s /tmp/gofmt.out ] # fails if gofmt would modify files

goimports -l . | tee /tmp/goimports.out
[ ! -s /tmp/goimports.out ]
```

> Tip: you can replace the temporary file locations above with paths that suit your workflow.
