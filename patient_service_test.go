package satusehat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alfinprasetya/go-satusehat-sdk/models"
)

func TestPatientServiceSearchByNIKUsesEnvironmentBaseURL(t *testing.T) {
	nik := envOrDefault("SATUSEHAT_TEST_PATIENT_NIK", "3175090101900001")
	token := envOrDefault("SATUSEHAT_TEST_BEARER_TOKEN", "mock-bearer-token")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want %s", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/Patient" {
			t.Fatalf("path = %s, want /Patient", r.URL.Path)
		}
		wantIdentifier := "https://fhir.kemkes.go.id/id/nik|" + nik
		if got := r.URL.Query().Get("identifier"); got != wantIdentifier {
			t.Fatalf("identifier = %s, want %s", got, wantIdentifier)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer "+token {
			t.Fatalf("Authorization = %s, want Bearer %s", got, token)
		}

		w.Header().Set("Content-Type", "application/fhir+json")
		_, _ = w.Write([]byte(`{
			"resourceType": "Bundle",
			"type": "searchset",
			"entry": [{
				"resource": {
					"resourceType": "Patient",
					"id": "patient-1",
					"meta": {
						"versionId": "1",
						"lastUpdated": "2024-01-02T03:04:05Z"
					},
					"identifier": [
						{
							"system": "https://fhir.kemkes.go.id/id/nik",
							"value": "3175090101900001"
						},
						{
							"system": "https://fhir.kemkes.go.id/id/ihs-number",
							"value": "P123456789"
						}
					],
					"active": true,
					"name": [{
						"use": "official",
						"text": "Budi Santoso"
					}],
					"gender": "male",
					"birthDate": "1990-01-01"
				}
			}]
		}`))
	}))
	defer server.Close()

	t.Setenv("FHIR_URL", server.URL)

	client := NewClient(
		envOrDefault("ORG_ID", "test-org"),
		os.Getenv("FHIR_URL"),
		&stubTokenProvider{token: token},
	)
	client.httpClient = server.Client()

	patient, err := client.Patients.Search(context.Background(), PatientSearchParams{NIK: &nik})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if patient.NIK != "3175090101900001" {
		t.Fatalf("NIK = %s, want 3175090101900001", patient.NIK)
	}
	if patient.ID != "P123456789" {
		t.Fatalf("ID = %s, want P123456789", patient.ID)
	}
	if patient.FullName != "Budi Santoso" {
		t.Fatalf("FullName = %s, want Budi Santoso", patient.FullName)
	}
	if patient.Gender != models.GenderMale {
		t.Fatalf("Gender = %s, want %s", patient.Gender, models.GenderMale)
	}
	if patient.BirthDate != "1990-01-01" {
		t.Fatalf("BirthDate = %s, want 1990-01-01", patient.BirthDate)
	}
	if patient.Meta == nil || patient.Meta.VersionID != "1" {
		t.Fatalf("Meta = %#v, want version 1", patient.Meta)
	}
}

func TestPatientServiceSearchRejectsMissingParams(t *testing.T) {
	client := NewClient("test-org", "https://fhir.example.test", nil)

	_, err := client.Patients.Search(context.Background(), PatientSearchParams{})
	if err == nil {
		t.Fatal("Search returned nil error, want invalid search params error")
	}
	if !strings.Contains(err.Error(), "invalid search params") {
		t.Fatalf("error = %v, want invalid search params", err)
	}
}

func TestPatientServiceSearchFromEnvironment(t *testing.T) {
	tokenURL := envOrSkip(t, "AUTH_URL")
	clientID := envOrSkip(t, "CLIENT_ID")
	clientSecret := envOrSkip(t, "CLIENT_SECRET")
	orgID := envOrSkip(t, "ORG_ID")
	fhirURL := envOrSkip(t, "FHIR_URL")

	params := patientSearchParamsFromEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	auth := NewOAuth2Provider(tokenURL, clientID, clientSecret)
	client := NewClient(orgID, fhirURL, auth)

	patient, err := client.Patients.Search(ctx, params)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if patient == nil {
		t.Fatal("patient is nil")
	}
	if patient.ID == "" {
		t.Fatal("patient ID is empty")
	}
}

func patientSearchParamsFromEnv(t *testing.T) PatientSearchParams {
	t.Helper()

	nik := strings.TrimSpace(os.Getenv("SATUSEHAT_TEST_PATIENT_NIK"))
	name := strings.TrimSpace(os.Getenv("SATUSEHAT_TEST_PATIENT_NAME"))
	birthdate := strings.TrimSpace(os.Getenv("SATUSEHAT_TEST_PATIENT_BIRTHDATE"))
	gender := models.Gender(strings.TrimSpace(os.Getenv("SATUSEHAT_TEST_PATIENT_GENDER")))

	switch {
	case name != "" && birthdate != "" && nik != "":
		return PatientSearchParams{Name: &name, Birthdate: &birthdate, NIK: &nik}
	case name != "" && birthdate != "" && gender != "":
		return PatientSearchParams{Name: &name, Birthdate: &birthdate, Gender: &gender}
	case name != "" && nik != "":
		return PatientSearchParams{Name: &name, NIK: &nik}
	case nik != "":
		return PatientSearchParams{NIK: &nik}
	default:
		t.Skip("set SATUSEHAT_TEST_PATIENT_NIK or SATUSEHAT_TEST_PATIENT_NAME plus SATUSEHAT_TEST_PATIENT_BIRTHDATE and SATUSEHAT_TEST_PATIENT_GENDER")
		return PatientSearchParams{}
	}
}
