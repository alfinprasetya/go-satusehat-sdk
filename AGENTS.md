# AGENTS.md

## Project overview

`go-satusehat-sdk` is a Go client library for Indonesia's SATUSEHAT FHIR R4 APIs. There is no long-running app or Docker stack in this repository—development means building the module and exercising the SDK (locally with mocks or against Kemkes-hosted staging/production with credentials).

## Common commands

| Task | Command |
|------|---------|
| Download dependencies | `go mod download` |
| Build | `go build ./...` |
| Vet (lint) | `go vet ./...` |
| Test | `go test ./...` |
| Format | `gofmt -w .` |

Go **1.26+** is required (`go.mod`). There are no `_test.go` files yet; `go test ./...` succeeds with `[no test files]`.

## Cursor Cloud specific instructions

- **No services to start** in tmux or Docker for local work. The SDK is imported by consumer apps; this repo only ships library packages.
- **Linting**: use `go vet ./...`. `golangci-lint` is not configured in the repo.
- **End-to-end against real SATUSEHAT** needs Kemkes credentials passed into constructors (not read from env in library code): OAuth token URL, client ID, client secret, organization ID, and FHIR base URL. Staging bases are documented in `README.md` (e.g. `https://api-satusehat-stg.dto.kemkes.go.id/oauth2/v1` and `.../fhir-r4/v1`).
- **Hello-world without secrets**: run a small program outside the repo with `replace github.com/alfinprasetya/go-satusehat-sdk => /workspace` and `httptest` servers for OAuth + FHIR, or use the smoke pattern under `/tmp/satusehat-smoke` during setup. Full live API calls require outbound HTTPS to Kemkes and valid SSP credentials.
- **Patient mapping**: `models.MapFHIRPatient` expects FHIR `meta.versionId` and `meta.lastUpdated` when mapping search results; minimal mocks must include `Meta` or mapping can panic.
