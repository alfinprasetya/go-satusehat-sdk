# go-satusehat-sdk

A high-performance, type-safe Go SDK for interacting with the Indonesian SATUSEHAT FHIR APIs.

Designed for backend services and healthcare applications with a clean, idiomatic Go API.

[![Go Reference](https://pkg.go.dev/badge/github.com/alfinprasetya/go-satusehat-sdk.svg)](https://pkg.go.dev/github.com/alfinprasetya/go-satusehat-sdk)

---

## Features

- OAuth2 client credentials authentication
- Automatic token management and refresh
- Type-safe FHIR resource handling
- Modular service-based architecture
- Full `context.Context` support
- Lightweight and dependency-minimal design
- Built on top of official FHIR Go models

---

## Installation

```bash
go get github.com/alfinprasetya/go-satusehat-sdk
```

---

## Quick Start

```go
package main

import (
	"context"
	"log"

	satusehat "github.com/alfinprasetya/go-satusehat-sdk"
	"github.com/alfinprasetya/go-satusehat-sdk/models"
)

func main() {
	auth := satusehat.NewOAuth2Provider(
		"<auth-url>",
		"<client-key>",
		"<client-secret>",
	)

	client := satusehat.NewClient(
		"<organization-id>",
		"<fhir-url>",
		auth,
	)

	// Use one search variant per call (do not combine Gender and NIK).
	name := "Salsabilla Anjani Rizki"
	birthdate := "2001-04-16"
	gender := models.GenderFemale

	patient, err := client.Patients.Search(context.Background(), satusehat.PatientSearchParams{
		Name:      &name,
		Birthdate: &birthdate,
		Gender:    &gender,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", patient)
}
```

`PatientSearchParams` supports four SATUSEHAT search shapes (one per call):

| Fields set | Postman equivalent |
|------------|-------------------|
| `Name`, `Birthdate`, `NIK` | Patient - Search Name, Birthdate, NIK |
| `Name`, `Birthdate`, `Gender` | Patient - Search Name, Birthdate, Gender |
| `Name`, `NIK` | Patient - Search Name, NIK |
| `NIK` only | Patient - Search NIK |

```go
// Name + Birthdate + NIK (returns full profile with real NIK)
nik := "9104025209000006"
patient, err := client.Patients.Search(ctx, satusehat.PatientSearchParams{
	Name: &name, Birthdate: &birthdate, NIK: &nik,
})

// Name + NIK (minimal profile, masked NIK)
patient, err = client.Patients.Search(ctx, satusehat.PatientSearchParams{
	Name: &name, NIK: &nik,
})

// NIK only
patient, err = client.Patients.Search(ctx, satusehat.PatientSearchParams{NIK: &nik})
```

### Search newborn by mother NIK

```go
birthdate := "2024-12-09"
patients, err := client.Patients.SearchNewbornsByMotherNIK(
	context.Background(),
	"9104025209000006", // mother NIK (required)
	&birthdate,         // newborn birthdate (optional; pass nil to omit)
)
if err != nil {
	log.Fatal(err)
}

for _, patient := range patients {
	log.Printf("mother NIK: %s, IHS: %s, name: %s", patient.MotherNIK, patient.IHSNumber, patient.FullName)
}
```

---

## Testing

Unit tests use `httptest` and Postman example bundles under `testdata/`:

```bash
make test        # go test ./...
make test-cover  # unit tests + coverage/unit.html (open in browser)
make vet         # go vet ./...
make check       # vet + unit tests
```

Live staging smoke tests (requires Kemkes credentials as environment variables). Integration tests cover `Patients.Search` (four Postman variants) and `SearchNewbornsByMotherNIK` (mother NIK with and without birthdate):

```bash
# Option A: .env in repo root (gitignored) with AUTH_URL, CLIENT_ID, CLIENT_SECRET, ORG_ID, FHIR_URL
make test-integration

# Option B: export variables yourself
export AUTH_URL="https://api-satusehat-stg.dto.kemkes.go.id/oauth2/v1"
export CLIENT_ID="..."
export CLIENT_SECRET="..."
export ORG_ID="..."
export FHIR_URL="https://api-satusehat-stg.dto.kemkes.go.id/fhir-r4/v1"
make test-integration

make test-integration-cover  # integration tests + coverage/integration.html

make test-all  # unit + integration
```

Coverage reports are written under `coverage/` (`unit.html`, `integration.html`, and per-function summaries in `*.txt`). Open the HTML files in a browser for line-by-line highlighting.

Run `make help` for all targets.

If gopls reports “No packages found” on `patient_service_integration_test.go`, reload the window after opening the repo (workspace `.vscode/settings.json` sets `gopls.buildFlags` to `-tags=integration`).

---

## Project Structure

```text
.
├── auth.go
├── client.go
├── docs
│   └── postman
│       ├── collections
│       │   ├── 00. FHIR Resource - Contoh Penggunaan.postman_collection.json
│       │   └── README.md
│       └── README.md
├── go.mod
├── go.sum
├── models
│   └── patients.go
├── patient_service.go
├── patient_service_test.go
└── README.md
```

---

## Postman Guidelines

Use the SATUSEHAT FHIR Postman collection JSON in `docs/postman/collections/` as a guideline for SDK resource implementation. Keep real credentials, tokens, and production patient data out of committed Postman files.

---

## Design Principles

- Idiomatic Go API design
- Strong type safety
- Minimal dependencies
- Easy SATUSEHAT integration
- Backend-service oriented architecture
- Production-ready authentication flow

---

## Roadmap

### Authentication

| Feature                   | Status |
| ------------------------- | ------ |
| OAuth2 Client Credentials | ✅     |
| Automatic Token Refresh   | ✅     |
| Thread-safe Token Cache   | ✅     |

---

### Patient Resource

| Feature                           | Status |
| --------------------------------- | ------ |
| GET Search Patient                | ✅     |
| GET Newborn Patient by Mother NIK | ✅     |
| POST Create Patient               | ⏳     |
| POST Create Newborn Patient       | ⏳     |
| PATCH Patient                     | ⏳     |

---

### Encounter Resource

| Feature                          | Status |
| -------------------------------- | ------ |
| GET Encounter                    | ⏳     |
| POST Encounter                   | ⏳     |
| PUT Encounter                    | ⏳     |
| Encounter Status History Helpers | ⏳     |

---

### Practitioner Resource

| Feature             | Status |
| ------------------- | ------ |
| GET Practitioner    | ⏳     |
| Search Practitioner | ⏳     |

---

### Organization Resource

| Feature             | Status |
| ------------------- | ------ |
| GET Organization    | ⏳     |
| Search Organization | ⏳     |

---

### Clinical Resources

| Resource           | Status |
| ------------------ | ------ |
| Condition          | ⏳     |
| Observation        | ⏳     |
| Procedure          | ⏳     |
| Medication         | ⏳     |
| MedicationRequest  | ⏳     |
| DiagnosticReport   | ⏳     |
| ServiceRequest     | ⏳     |
| AllergyIntolerance | ⏳     |
| Immunization       | ⏳     |
| CarePlan           | ⏳     |

---

### Imaging & Diagnostic Resources

| Resource     | Status |
| ------------ | ------ |
| Media        | ⏳     |
| ImagingStudy | ⏳     |
| Specimen     | ⏳     |

---

<!-- ### SDK Features

| Feature                      | Status |
| ---------------------------- | ------ |
| Typed Request Builders       | 🚧     |
| Typed Response Models        | 🚧     |
| FHIR Bundle Helpers          | ⏳     |
| Pagination Helpers           | ⏳     |
| Retry Middleware             | ⏳     |
| Structured Error Handling    | 🚧     |
| Logging Hooks                | ⏳     |
| Custom HTTP Client Support   | ⏳     |
| Context Cancellation Support | ✅     |

--- -->

### Legend

- ✅ Implemented
- 🚧 In Progress
- ⏳ Planned

---

## Status

> Alpha — APIs may change before stable release.

This project is currently under active development and intended for early integration testing.

---

## License

MIT License.

Copyright (c) 2026 Alfin Prasetya
