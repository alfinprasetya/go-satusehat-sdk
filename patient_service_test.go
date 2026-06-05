package satusehat

import (
	"context"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alfinprasetya/go-satusehat-sdk/models"
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

	const motherNIK = "9104025209000006"
	birthdate := "2024-12-09"

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
		&birthdate,
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
		nil,
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
		nil,
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

	birthdate := "2024-12-09"
	_, err := client.Patients.SearchNewbornsByMotherNIK(
		context.Background(),
		"",
		&birthdate,
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "mother NIK is required") {
		t.Errorf("error: got %q, want containing %q", err.Error(), "mother NIK is required")
	}
}

// Postman fixture patient (Patient - Search * examples).
const (
	patientSearchName      = "Salsabilla Anjani Rizki"
	patientSearchBirthdate = "2001-04-16"
	patientSearchNIK       = "9104025209000006"
	patientSearchIHS       = "P02280547535"
	patientSearchMaskedNIK = "################"

	// Postman fixture newborn (Patient - Bayi Search NIK Ibu *).
	newbornMotherNIK       = "9104025209000006"
	newbornLouisaBirthdate = "2024-12-09"
	newbornLouisaIHS       = "P20394967125"
	newbornLouisaName      = "LOUISA MINGAME"
)

var patientSearchGender = models.GenderFemale

const patientNIKIdentifier = "https://fhir.kemkes.go.id/id/nik|" + patientSearchNIK

//go:embed testdata/patient_search_name_birthdate_gender.json
var patientSearchNameBirthdateGenderBundle []byte

//go:embed testdata/patient_search_name_birthdate_nik.json
var patientSearchNameBirthdateNIKBundle []byte

//go:embed testdata/patient_search_name_nik.json
var patientSearchNameNIKBundle []byte

//go:embed testdata/patient_search_nik.json
var patientSearchNIKOnlyBundle []byte

//go:embed testdata/patient_search_not_found.json
var patientSearchNotFoundBundle []byte

func newPatientTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := NewClient("org-id", server.URL, nil)
	client.httpClient = server.Client()
	return client
}

func TestPatientSearch_QueryParams(t *testing.T) {
	t.Parallel()

	name := patientSearchName
	birthdate := patientSearchBirthdate
	nik := patientSearchNIK
	gender := patientSearchGender

	tests := []struct {
		name       string
		params     PatientSearchParams
		wantName   string
		wantBD     string
		wantGender string
		wantIdent  string
	}{
		{
			name: "name_birthdate_nik",
			params: PatientSearchParams{
				Name:      &name,
				Birthdate: &birthdate,
				NIK:       &nik,
			},
			wantName:  name,
			wantBD:    birthdate,
			wantIdent: patientNIKIdentifier,
		},
		{
			name: "name_birthdate_gender",
			params: PatientSearchParams{
				Name:      &name,
				Birthdate: &birthdate,
				Gender:    &gender,
			},
			wantName:   name,
			wantBD:     birthdate,
			wantGender: string(gender),
		},
		{
			name: "name_nik",
			params: PatientSearchParams{
				Name: &name,
				NIK:  &nik,
			},
			wantName:  name,
			wantIdent: patientNIKIdentifier,
		},
		{
			name: "nik_only",
			params: PatientSearchParams{
				NIK: &nik,
			},
			wantIdent: patientNIKIdentifier,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := newPatientTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method: got %s, want GET", r.Method)
				}
				if r.URL.Path != "/Patient" {
					t.Errorf("path: got %s, want /Patient", r.URL.Path)
				}

				q := r.URL.Query()
				if got := q.Get("name"); got != tt.wantName {
					t.Errorf("name: got %q, want %q", got, tt.wantName)
				}
				if got := q.Get("birthdate"); got != tt.wantBD {
					t.Errorf("birthdate: got %q, want %q", got, tt.wantBD)
				}
				if got := q.Get("gender"); got != tt.wantGender {
					t.Errorf("gender: got %q, want %q", got, tt.wantGender)
				}
				if got := q.Get("identifier"); got != tt.wantIdent {
					t.Errorf("identifier: got %q, want %q", got, tt.wantIdent)
				}

				w.Header().Set("Content-Type", "application/fhir+json")
				_, _ = w.Write(patientSearchNameBirthdateNIKBundle)
			})

			_, err := client.Patients.Search(context.Background(), tt.params)
			if err != nil {
				t.Fatalf("Search: %v", err)
			}
		})
	}
}

func TestPatientSearch_InvalidParams(t *testing.T) {
	t.Parallel()

	name := patientSearchName
	birthdate := patientSearchBirthdate

	client := NewClient("org-id", "http://example.com", nil)

	_, err := client.Patients.Search(context.Background(), PatientSearchParams{
		Name:      &name,
		Birthdate: &birthdate,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid search params") {
		t.Errorf("error: got %q, want containing %q", err.Error(), "invalid search params")
	}
}

func TestPatientSearch_NotFound(t *testing.T) {
	t.Parallel()

	nik := patientSearchNIK

	client := newPatientTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/fhir+json")
		_, _ = w.Write(patientSearchNotFoundBundle)
	})

	_, err := client.Patients.Search(context.Background(), PatientSearchParams{NIK: &nik})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "patient not found") {
		t.Errorf("error: got %q, want containing %q", err.Error(), "patient not found")
	}
}

func TestPatientSearch_HTTPError(t *testing.T) {
	t.Parallel()

	nik := patientSearchNIK

	client := newPatientTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := client.Patients.Search(context.Background(), PatientSearchParams{NIK: &nik})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "fhir api error") {
		t.Errorf("error: got %q, want containing %q", err.Error(), "fhir api error")
	}
}

func assertPatientSearchFullProfile(t *testing.T, patient *models.Patient, wantNIK string) {
	t.Helper()

	if patient.FullName != patientSearchName {
		t.Errorf("FullName: got %q, want %q", patient.FullName, patientSearchName)
	}
	if patient.NIK != wantNIK {
		t.Errorf("NIK: got %q, want %q", patient.NIK, wantNIK)
	}
	if patient.IHSNumber != patientSearchIHS {
		t.Errorf("IHSNumber: got %q, want %q", patient.IHSNumber, patientSearchIHS)
	}
	if patient.BirthDate != patientSearchBirthdate {
		t.Errorf("BirthDate: got %q, want %q", patient.BirthDate, patientSearchBirthdate)
	}
	if patient.Gender != models.GenderFemale {
		t.Errorf("Gender: got %q, want %q", patient.Gender, models.GenderFemale)
	}
	if patient.Marital != models.MaritalMarried {
		t.Errorf("Marital: got %q, want %q", patient.Marital, models.MaritalMarried)
	}
	if patient.Citizenship != models.CitizenshipWNI {
		t.Errorf("Citizenship: got %q, want %q", patient.Citizenship, models.CitizenshipWNI)
	}
	if patient.BirthPlace != "Bandung" {
		t.Errorf("BirthPlace: got %q, want Bandung", patient.BirthPlace)
	}
	if patient.Contact.MobilePhone != "08123456789" {
		t.Errorf("Contact.MobilePhone: got %q, want 08123456789", patient.Contact.MobilePhone)
	}
	if patient.Contact.HomePhone != "+622123456789" {
		t.Errorf("Contact.HomePhone: got %q, want +622123456789", patient.Contact.HomePhone)
	}
	if patient.Contact.Email != "john.smith@xyz.com" {
		t.Errorf("Contact.Email: got %q, want john.smith@xyz.com", patient.Contact.Email)
	}
	if patient.PreferredLanguage != "Indonesian" {
		t.Errorf("PreferredLanguage: got %q, want Indonesian", patient.PreferredLanguage)
	}
	if len(patient.EmergencyContacts) != 1 {
		t.Fatalf("len(EmergencyContacts): got %d, want 1", len(patient.EmergencyContacts))
	}
	if patient.EmergencyContacts[0].Name != "Jane Smith" {
		t.Errorf("EmergencyContacts[0].Name: got %q, want Jane Smith", patient.EmergencyContacts[0].Name)
	}
	if len(patient.EmergencyContacts[0].Phones) != 1 || patient.EmergencyContacts[0].Phones[0] != "0690383372" {
		t.Errorf("EmergencyContacts[0].Phones: got %v, want [0690383372]", patient.EmergencyContacts[0].Phones)
	}
	if !patient.Active {
		t.Error("Active: got false, want true")
	}
	if patient.Deceased {
		t.Error("Deceased: got true, want false")
	}
	if patient.Meta == nil || patient.Meta.VersionID == "" {
		t.Error("Meta: expected non-empty VersionID")
	}
}

func assertPatientSearchPartialProfile(t *testing.T, patient *models.Patient) {
	t.Helper()

	if patient.FullName != "Sa** An** Ri**" {
		t.Errorf("FullName: got %q, want Sa** An** Ri**", patient.FullName)
	}
	if patient.NIK != patientSearchMaskedNIK {
		t.Errorf("NIK: got %q, want %q", patient.NIK, patientSearchMaskedNIK)
	}
	if patient.IHSNumber != patientSearchIHS {
		t.Errorf("IHSNumber: got %q, want %q", patient.IHSNumber, patientSearchIHS)
	}
	if !patient.Active {
		t.Error("Active: got false, want true")
	}
	if patient.BirthDate != "" {
		t.Errorf("BirthDate: got %q, want empty", patient.BirthDate)
	}
	if patient.Gender != "" {
		t.Errorf("Gender: got %q, want empty", patient.Gender)
	}
	if patient.Contact.MobilePhone != "" || patient.Contact.Email != "" {
		t.Errorf("Contact: got %+v, want empty", patient.Contact)
	}
	if len(patient.EmergencyContacts) != 0 {
		t.Errorf("EmergencyContacts: got %d, want 0", len(patient.EmergencyContacts))
	}
	if patient.Meta == nil || patient.Meta.VersionID == "" {
		t.Error("Meta: expected non-empty VersionID")
	}
}

func TestPatientSearch_FullProfile_NameBirthdateNIK(t *testing.T) {
	t.Parallel()

	name := patientSearchName
	birthdate := patientSearchBirthdate
	nik := patientSearchNIK

	client := newPatientTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/fhir+json")
		_, _ = w.Write(patientSearchNameBirthdateNIKBundle)
	})

	patient, err := client.Patients.Search(context.Background(), PatientSearchParams{
		Name:      &name,
		Birthdate: &birthdate,
		NIK:       &nik,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	assertPatientSearchFullProfile(t, patient, patientSearchNIK)
}

func TestPatientSearch_FullProfile_NameBirthdateGender(t *testing.T) {
	t.Parallel()

	name := patientSearchName
	birthdate := patientSearchBirthdate
	gender := patientSearchGender

	client := newPatientTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/fhir+json")
		_, _ = w.Write(patientSearchNameBirthdateGenderBundle)
	})

	patient, err := client.Patients.Search(context.Background(), PatientSearchParams{
		Name:      &name,
		Birthdate: &birthdate,
		Gender:    &gender,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	assertPatientSearchFullProfile(t, patient, patientSearchMaskedNIK)
}

func TestPatientSearch_PartialProfile_NameNIK(t *testing.T) {
	t.Parallel()

	name := patientSearchName
	nik := patientSearchNIK

	client := newPatientTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/fhir+json")
		_, _ = w.Write(patientSearchNameNIKBundle)
	})

	patient, err := client.Patients.Search(context.Background(), PatientSearchParams{
		Name: &name,
		NIK:  &nik,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	assertPatientSearchPartialProfile(t, patient)
}

func TestPatientSearch_PartialProfile_NIKOnly(t *testing.T) {
	t.Parallel()

	nik := patientSearchNIK

	client := newPatientTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/fhir+json")
		_, _ = w.Write(patientSearchNIKOnlyBundle)
	})

	patient, err := client.Patients.Search(context.Background(), PatientSearchParams{
		NIK: &nik,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	assertPatientSearchPartialProfile(t, patient)
}
