//go:build integration

package satusehat

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/alfinprasetya/go-satusehat-sdk/models"
)

func requireEnv(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Skipf("skipping integration test: %s not set", key)
	}
	return v
}

func newIntegrationClient(t *testing.T) *Client {
	t.Helper()

	auth := NewOAuth2Provider(
		requireEnv(t, "AUTH_URL"),
		requireEnv(t, "CLIENT_ID"),
		requireEnv(t, "CLIENT_SECRET"),
	)

	return NewClient(
		requireEnv(t, "ORG_ID"),
		requireEnv(t, "FHIR_URL"),
		auth,
	)
}

func TestPatientSearch_Integration(t *testing.T) {
	client := newIntegrationClient(t)

	name := patientSearchName
	birthdate := patientSearchBirthdate
	nik := patientSearchNIK
	gender := patientSearchGender

	tests := []struct {
		name   string
		params PatientSearchParams
	}{
		{
			name: "name_birthdate_nik",
			params: PatientSearchParams{
				Name:      &name,
				Birthdate: &birthdate,
				NIK:       &nik,
			},
		},
		{
			name: "name_birthdate_gender",
			params: PatientSearchParams{
				Name:      &name,
				Birthdate: &birthdate,
				Gender:    &gender,
			},
		},
		{
			name: "name_nik",
			params: PatientSearchParams{
				Name: &name,
				NIK:  &nik,
			},
		},
		{
			name: "nik_only",
			params: PatientSearchParams{
				NIK: &nik,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patient, err := client.Patients.Search(context.Background(), tt.params)
			if err != nil {
				t.Fatalf("Search: %v", err)
			}
			if patient.IHSNumber == "" {
				t.Error("IHSNumber: expected non-empty")
			}
			if !strings.Contains(patient.FullName, "Salsabilla") &&
				!strings.Contains(patient.FullName, "Sa**") {
				t.Errorf("FullName: got %q, expected fixture patient name", patient.FullName)
			}
			if !patient.Active {
				t.Error("Active: got false, want true")
			}
			if patient.Meta == nil {
				t.Error("Meta: expected non-nil")
			}
		})
	}
}

func TestPatientSearch_Integration_NameBirthdateNIK_RealNIK(t *testing.T) {
	client := newIntegrationClient(t)

	name := patientSearchName
	birthdate := patientSearchBirthdate
	nik := patientSearchNIK

	patient, err := client.Patients.Search(context.Background(), PatientSearchParams{
		Name:      &name,
		Birthdate: &birthdate,
		NIK:       &nik,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if patient.NIK != patientSearchNIK {
		t.Errorf("NIK: got %q, want %q", patient.NIK, patientSearchNIK)
	}
	if patient.Gender != models.GenderFemale {
		t.Errorf("Gender: got %q, want %q", patient.Gender, models.GenderFemale)
	}
}

func assertNewbornIntegrationPatients(t *testing.T, patients []*models.Patient) {
	t.Helper()

	if len(patients) < 1 {
		t.Fatalf("len(patients): got %d, want >= 1", len(patients))
	}

	for i, patient := range patients {
		if patient.MotherNIK != newbornMotherNIK {
			t.Errorf("patients[%d].MotherNIK: got %q, want %q", i, patient.MotherNIK, newbornMotherNIK)
		}
		if patient.IHSNumber == "" {
			t.Errorf("patients[%d].IHSNumber: expected non-empty", i)
		}
		if !patient.Active {
			t.Errorf("patients[%d].Active: got false, want true", i)
		}
	}
}

func TestSearchNewbornsByMotherNIK_Integration(t *testing.T) {
	client := newIntegrationClient(t)
	birthdate := newbornLouisaBirthdate

	tests := []struct {
		name      string
		birthdate *string
	}{
		{name: "mother_nik_with_birthdate", birthdate: &birthdate},
		{name: "mother_nik_only", birthdate: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patients, err := client.Patients.SearchNewbornsByMotherNIK(
				context.Background(),
				newbornMotherNIK,
				tt.birthdate,
			)
			if err != nil {
				t.Fatalf("SearchNewbornsByMotherNIK: %v", err)
			}
			assertNewbornIntegrationPatients(t, patients)
		})
	}
}

func TestSearchNewbornsByMotherNIK_Integration_Louisa(t *testing.T) {
	client := newIntegrationClient(t)
	birthdate := newbornLouisaBirthdate

	patients, err := client.Patients.SearchNewbornsByMotherNIK(
		context.Background(),
		newbornMotherNIK,
		&birthdate,
	)
	if err != nil {
		t.Fatalf("SearchNewbornsByMotherNIK: %v", err)
	}

	for _, patient := range patients {
		if patient.IHSNumber == newbornLouisaIHS ||
			strings.Contains(patient.FullName, "LOUISA") {
			return
		}
	}

	t.Skip("Louisa fixture newborn not present in staging")
}
