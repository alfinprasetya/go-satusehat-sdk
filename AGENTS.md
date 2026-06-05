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
- **End-to-end against real SATUSEHAT** needs Kemkes credentials passed into constructors (the library does not read env itself). In Cursor Cloud, these secrets are typically injected as: `AUTH_URL`, `CLIENT_ID`, `CLIENT_SECRET`, `ORG_ID`, `FHIR_URL`. Staging bases are in `README.md` (e.g. `https://api-satusehat-stg.dto.kemkes.go.id/oauth2/v1` and `.../fhir-r4/v1`).
- **Live smoke test** (from repo root, with secrets set): `go run /tmp/satusehat-live` after creating `/tmp/satusehat-live` with `replace github.com/alfinprasetya/go-satusehat-sdk => /workspace` (or any small `main` that wires the five env vars into `NewOAuth2Provider` / `NewClient` and calls `client.Patients.Search`).
- **Hello-world without secrets**: use `httptest` mocks with `replace ... => /workspace` (see prior setup notes). Full live API calls need outbound HTTPS to Kemkes and valid SSP credentials.
- **Patient mapping**: `models.MapFHIRPatient` expects FHIR `meta.versionId` and `meta.lastUpdated` when mapping search results; minimal mocks must include `Meta` or mapping can panic.
