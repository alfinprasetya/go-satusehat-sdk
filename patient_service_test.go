package satusehat

import (
	"context"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newbornSearchBundle is the success response from Postman request
// "Patient - Bayi Search NIK Ibu" in docs/postman/collections/.
const newbornSearchBundle = `{
    "entry": [
        {
            "fullUrl": "https://api-satusehat-stg.dto.kemkes.go.id/fhir-r4/v1/Patient/P20394967125",
            "resource": {
                "active": true,
                "birthDate": "2024-12-09",
                "id": "P20394967125",
                "identifier": [
                    {
                        "system": "https://fhir.kemkes.go.id/id/nik-ibu",
                        "use": "official",
                        "value": "9104025209000006"
                    },
                    {
                        "system": "https://fhir.kemkes.go.id/id/ihs-number",
                        "use": "official",
                        "value": "P20394967125"
                    }
                ],
                "meta": {
                    "lastUpdated": "2024-12-09T05:07:55.249926+00:00",
                    "profile": [
                        "https://fhir.kemkes.go.id/r4/StructureDefinition/Patient"
                    ],
                    "versionId": "MTczMzcyMDg3NTI0OTkyNjAwMA"
                },
                "multipleBirthInteger": 0,
                "name": [
                    {
                        "text": "LOUISA MINGAME",
                        "use": "official"
                    }
                ],
                "resourceType": "Patient"
            }
        }
    ],
    "resourceType": "Bundle",
    "total": 1,
    "type": "searchset"
}`

// newbornMultipleSearchBundle is from Postman response
// "Patient - Bayi Search NIK Ibu" (without birthdate) in docs/postman/collections/.
//
//go:embed testdata/patient_bayi_search_nik_ibu.json
var newbornMultipleSearchBundle []byte

const emptySearchBundle = `{
    "resourceType": "Bundle",
    "total": 0,
    "type": "searchset"
}`

func TestSearchNewbornsByMotherNIK_Success(t *testing.T) {
	t.Parallel()

	const (
		motherNIK = "9104025209000006"
		birthdate = "2024-12-09"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: got %s, want GET", r.Method)
		}

		if r.URL.Path != "/Patient" {
			t.Errorf("path: got %s, want /Patient", r.URL.Path)
		}

		identifier := r.URL.Query().Get("identifier")
		wantIdentifier := "https://fhir.kemkes.go.id/id/nik-ibu|" + motherNIK
		if identifier != wantIdentifier {
			t.Errorf("identifier: got %q, want %q", identifier, wantIdentifier)
		}

		if got := r.URL.Query().Get("birthdate"); got != birthdate {
			t.Errorf("birthdate: got %q, want %q", got, birthdate)
		}

		w.Header().Set("Content-Type", "application/fhir+json")
		_, _ = w.Write([]byte(newbornSearchBundle))
	}))
	defer server.Close()

	client := NewClient("org-id", server.URL, nil)
	client.httpClient = server.Client()

	patients, err := client.Patients.SearchNewbornsByMotherNIK(
		context.Background(),
		motherNIK,
		birthdate,
	)
	if err != nil {
		t.Fatalf("SearchNewbornsByMotherNIK: %v", err)
	}

	if len(patients) != 1 {
		t.Fatalf("len(patients): got %d, want 1", len(patients))
	}

	patient := patients[0]
	if patient.MotherNIK != motherNIK {
		t.Errorf("MotherNIK: got %q, want %q", patient.MotherNIK, motherNIK)
	}
	if patient.IHSNumber != "P20394967125" {
		t.Errorf("IHSNumber: got %q, want P20394967125", patient.IHSNumber)
	}
	if patient.FullName != "LOUISA MINGAME" {
		t.Errorf("FullName: got %q, want LOUISA MINGAME", patient.FullName)
	}
	if patient.BirthDate != birthdate {
		t.Errorf("BirthDate: got %q, want %q", patient.BirthDate, birthdate)
	}
	if !patient.Active {
		t.Error("Active: got false, want true")
	}
}

func TestSearchNewbornsByMotherNIK_Multiple(t *testing.T) {
	t.Parallel()

	const motherNIK = "9104025209000006"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identifier := r.URL.Query().Get("identifier")
		wantIdentifier := "https://fhir.kemkes.go.id/id/nik-ibu|" + motherNIK
		if identifier != wantIdentifier {
			t.Errorf("identifier: got %q, want %q", identifier, wantIdentifier)
		}

		if got := r.URL.Query().Get("birthdate"); got != "" {
			t.Errorf("birthdate: got %q, want empty", got)
		}

		w.Header().Set("Content-Type", "application/fhir+json")
		_, _ = w.Write(newbornMultipleSearchBundle)
	}))
	defer server.Close()

	client := NewClient("org-id", server.URL, nil)
	client.httpClient = server.Client()

	patients, err := client.Patients.SearchNewbornsByMotherNIK(
		context.Background(),
		motherNIK,
		"",
	)
	if err != nil {
		t.Fatalf("SearchNewbornsByMotherNIK: %v", err)
	}

	if len(patients) != 73 {
		t.Fatalf("len(patients): got %d, want 73", len(patients))
	}

	first := patients[0]
	if first.IHSNumber != "P20395871417" {
		t.Errorf("patients[0].IHSNumber: got %q, want P20395871417", first.IHSNumber)
	}
	if first.FullName != "John Smith saria" {
		t.Errorf("patients[0].FullName: got %q, want John Smith saria", first.FullName)
	}
	if first.BirthDate != "2026-05-09" {
		t.Errorf("patients[0].BirthDate: got %q, want 2026-05-09", first.BirthDate)
	}
	if first.MotherNIK != motherNIK {
		t.Errorf("patients[0].MotherNIK: got %q, want %q", first.MotherNIK, motherNIK)
	}

	// Entry from Postman "Patient - Bayi Search NIK Ibu + Birthdate" example.
	louisa := patients[68]
	if louisa.IHSNumber != "P20394967125" {
		t.Errorf("patients[68].IHSNumber: got %q, want P20394967125", louisa.IHSNumber)
	}
	if louisa.FullName != "LOUISA MINGAME" {
		t.Errorf("patients[68].FullName: got %q, want LOUISA MINGAME", louisa.FullName)
	}
	if louisa.BirthDate != "2024-12-09" {
		t.Errorf("patients[68].BirthDate: got %q, want 2024-12-09", louisa.BirthDate)
	}
}

func TestSearchNewbornsByMotherNIK_Empty(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/fhir+json")
		_, _ = w.Write([]byte(emptySearchBundle))
	}))
	defer server.Close()

	client := NewClient("org-id", server.URL, nil)
	client.httpClient = server.Client()

	patients, err := client.Patients.SearchNewbornsByMotherNIK(
		context.Background(),
		"9104025209000006",
		"",
	)
	if err != nil {
		t.Fatalf("SearchNewbornsByMotherNIK: %v", err)
	}
	if len(patients) != 0 {
		t.Errorf("len(patients): got %d, want 0", len(patients))
	}
}

func TestSearchNewbornsByMotherNIK_Validation(t *testing.T) {
	t.Parallel()

	client := NewClient("org-id", "http://example.com", nil)

	_, err := client.Patients.SearchNewbornsByMotherNIK(
		context.Background(),
		"",
		"2024-12-09",
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mother NIK is required") {
		t.Errorf("error: got %q, want containing %q", err.Error(), "mother NIK is required")
	}
}
