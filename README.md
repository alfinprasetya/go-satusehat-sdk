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

	name := "Dr. Alan Bagus Prasetya"
	birthdate := "1977-09-03"
	nik := "9104223107000004"
	gender := models.GenderMale
	patient, err := client.Patients.Search(context.Background(), satusehat.PatientSearchParams{
		Name:      &name,
		Birthdate: &birthdate,
		Gender:    &gender,
		NIK:       &nik,
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", patient)
}
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
