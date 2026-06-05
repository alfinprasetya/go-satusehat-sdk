package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

const (
	patientSearchNIK       = "9104025209000006"
	patientSearchMaskedNIK = "################"
	patientSearchIHS       = "P02280547535"
	patientSearchName      = "Salsabilla Anjani Rizki"
	patientSearchBirthdate = "2001-04-16"
	newbornMotherNIK       = "9104025209000006"
	newbornLouisaBirthdate = "2024-12-09"
	newbornLouisaIHS       = "P20394967125"
	newbornLouisaName      = "LOUISA MINGAME"
)

func loadBundle(t *testing.T, filename string) *fhir.Bundle {
	t.Helper()

	path := filepath.Join("..", "testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	bundle := &fhir.Bundle{}
	if err := json.Unmarshal(data, bundle); err != nil {
		t.Fatalf("decode bundle %s: %v", filename, err)
	}
	return bundle
}

func decodePatientResource(t *testing.T, raw json.RawMessage) *fhir.Patient {
	t.Helper()

	patient := &fhir.Patient{}
	if err := json.Unmarshal(raw, patient); err != nil {
		t.Fatalf("decode patient: %v", err)
	}
	return patient
}

func loadFirstPatientFromBundle(t *testing.T, filename string) *fhir.Patient {
	t.Helper()

	bundle := loadBundle(t, filename)
	if len(bundle.Entry) == 0 {
		t.Fatalf("bundle %s has no entries", filename)
	}
	return decodePatientResource(t, bundle.Entry[0].Resource)
}

func loadPatientFromBundleByIHS(t *testing.T, filename, ihs string) *fhir.Patient {
	t.Helper()

	const ihsSystem = "https://fhir.kemkes.go.id/id/ihs-number"

	bundle := loadBundle(t, filename)
	for _, entry := range bundle.Entry {
		patient := decodePatientResource(t, entry.Resource)
		for _, identifier := range patient.Identifier {
			if identifier.System != nil && *identifier.System == ihsSystem &&
				identifier.Value != nil && *identifier.Value == ihs {
				return patient
			}
		}
	}
	t.Fatalf("patient with IHS %q not found in %s", ihs, filename)
	return nil
}

func mapPatient(t *testing.T, src *fhir.Patient) *Patient {
	t.Helper()

	out := &Patient{}
	if err := MapFHIRPatient(src, out); err != nil {
		t.Fatalf("MapFHIRPatient: %v", err)
	}
	return out
}

func assertMapPatient(t *testing.T, filename string, check func(t *testing.T, patient *Patient)) {
	t.Helper()

	src := loadFirstPatientFromBundle(t, filename)
	check(t, mapPatient(t, src))
}

func minimalFHIRPatient() *fhir.Patient {
	active := true
	versionID := "MTczMTU2ODUwMjgyNDQyMzAwMA"
	lastUpdated := "2024-11-14T07:15:02.824423+00:00"
	official := fhir.NameUseOfficial
	name := "Test Patient"

	return &fhir.Patient{
		Active: &active,
		Meta: &fhir.Meta{
			VersionId:   &versionID,
			LastUpdated: &lastUpdated,
		},
		Name: []fhir.HumanName{{
			Use:  &official,
			Text: &name,
		}},
	}
}

func TestMapFHIRPatient_nil(t *testing.T) {
	t.Parallel()

	out := &Patient{FullName: "unchanged"}
	if err := MapFHIRPatient(nil, out); err != nil {
		t.Fatalf("MapFHIRPatient: %v", err)
	}
	if out.FullName != "unchanged" {
		t.Errorf("FullName: got %q, want unchanged", out.FullName)
	}
}

func TestMapFHIRPatient_fullProfile(t *testing.T) {
	t.Parallel()

	assertMapPatient(t, "patient_search_name_birthdate_nik.json", func(t *testing.T, patient *Patient) {
		if patient.FullName != patientSearchName {
			t.Errorf("FullName: got %q, want %q", patient.FullName, patientSearchName)
		}
		if patient.NIK != patientSearchNIK {
			t.Errorf("NIK: got %q, want %q", patient.NIK, patientSearchNIK)
		}
		if patient.IHSNumber != patientSearchIHS {
			t.Errorf("IHSNumber: got %q, want %q", patient.IHSNumber, patientSearchIHS)
		}
		if patient.ID != patientSearchIHS {
			t.Errorf("ID: got %q, want %q", patient.ID, patientSearchIHS)
		}
		if patient.BirthDate != patientSearchBirthdate {
			t.Errorf("BirthDate: got %q, want %q", patient.BirthDate, patientSearchBirthdate)
		}
		if patient.Gender != GenderFemale {
			t.Errorf("Gender: got %q, want %q", patient.Gender, GenderFemale)
		}
		if patient.Marital != MaritalMarried {
			t.Errorf("Marital: got %q, want %q", patient.Marital, MaritalMarried)
		}
		if patient.Citizenship != CitizenshipWNI {
			t.Errorf("Citizenship: got %q, want %q", patient.Citizenship, CitizenshipWNI)
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
		if patient.Meta != nil && patient.Meta.LastUpdate.IsZero() {
			t.Error("Meta.LastUpdate: expected parsed time")
		}
	})
}

func TestMapFHIRPatient_partialProfile(t *testing.T) {
	t.Parallel()

	assertMapPatient(t, "patient_search_name_nik.json", func(t *testing.T, patient *Patient) {
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
	})
}

func TestMapFHIRPatient_newborn(t *testing.T) {
	t.Parallel()

	src := loadPatientFromBundleByIHS(t, "patient_bayi_search_nik_ibu.json", newbornLouisaIHS)
	patient := mapPatient(t, src)

	if patient.MotherNIK != newbornMotherNIK {
		t.Errorf("MotherNIK: got %q, want %q", patient.MotherNIK, newbornMotherNIK)
	}
	if patient.IHSNumber != newbornLouisaIHS {
		t.Errorf("IHSNumber: got %q, want %q", patient.IHSNumber, newbornLouisaIHS)
	}
	if patient.FullName != newbornLouisaName {
		t.Errorf("FullName: got %q, want %q", patient.FullName, newbornLouisaName)
	}
	if patient.BirthDate != newbornLouisaBirthdate {
		t.Errorf("BirthDate: got %q, want %q", patient.BirthDate, newbornLouisaBirthdate)
	}
	if !patient.Active {
		t.Error("Active: got false, want true")
	}
	if patient.Meta == nil || patient.Meta.VersionID == "" {
		t.Error("Meta: expected non-empty VersionID")
	}
}

func TestMapFHIRPatient_edgeCases(t *testing.T) {
	t.Parallel()

	phone := fhir.ContactPointSystemPhone
	mobile := fhir.ContactPointUseMobile
	official := fhir.NameUseOfficial

	tests := []struct {
		name  string
		build func() *fhir.Patient
		check func(t *testing.T, patient *Patient)
	}{
		{
			name: "citizenship WNA",
			build: func() *fhir.Patient {
				p := minimalFHIRPatient()
				wna := "WNA"
				p.Extension = []fhir.Extension{{
					Url:       "https://fhir.kemkes.go.id/r4/StructureDefinition/citizenshipStatus",
					ValueCode: &wna,
				}}
				return p
			},
			check: func(t *testing.T, patient *Patient) {
				if patient.Citizenship != CitizenshipWNA {
					t.Errorf("Citizenship: got %q, want %q", patient.Citizenship, CitizenshipWNA)
				}
			},
		},
		{
			name: "marital Single",
			build: func() *fhir.Patient {
				p := minimalFHIRPatient()
				text := "Single"
				p.MaritalStatus = &fhir.CodeableConcept{Text: &text}
				return p
			},
			check: func(t *testing.T, patient *Patient) {
				if patient.Marital != MaritalSingle {
					t.Errorf("Marital: got %q, want %q", patient.Marital, MaritalSingle)
				}
			},
		},
		{
			name: "marital Unmarried",
			build: func() *fhir.Patient {
				p := minimalFHIRPatient()
				text := "Unmarried"
				p.MaritalStatus = &fhir.CodeableConcept{Text: &text}
				return p
			},
			check: func(t *testing.T, patient *Patient) {
				if patient.Marital != MaritalUnmarried {
					t.Errorf("Marital: got %q, want %q", patient.Marital, MaritalUnmarried)
				}
			},
		},
		{
			name: "marital Divorced",
			build: func() *fhir.Patient {
				p := minimalFHIRPatient()
				text := "Divorced"
				p.MaritalStatus = &fhir.CodeableConcept{Text: &text}
				return p
			},
			check: func(t *testing.T, patient *Patient) {
				if patient.Marital != MaritalDivorced {
					t.Errorf("Marital: got %q, want %q", patient.Marital, MaritalDivorced)
				}
			},
		},
		{
			name: "deceased true",
			build: func() *fhir.Patient {
				p := minimalFHIRPatient()
				deceased := true
				p.DeceasedBoolean = &deceased
				return p
			},
			check: func(t *testing.T, patient *Patient) {
				if !patient.Deceased {
					t.Error("Deceased: got false, want true")
				}
			},
		},
		{
			name: "preferred communication overrides first",
			build: func() *fhir.Patient {
				p := minimalFHIRPatient()
				english := "English"
				indonesian := "Indonesian"
				preferred := true
				p.Communication = []fhir.PatientCommunication{
					{Language: fhir.CodeableConcept{Text: &english}},
					{Language: fhir.CodeableConcept{Text: &indonesian}, Preferred: &preferred},
				}
				return p
			},
			check: func(t *testing.T, patient *Patient) {
				if patient.PreferredLanguage != "Indonesian" {
					t.Errorf("PreferredLanguage: got %q, want Indonesian", patient.PreferredLanguage)
				}
			},
		},
		{
			name: "multiple emergency contacts",
			build: func() *fhir.Patient {
				p := minimalFHIRPatient()
				jane := "Jane Smith"
				john := "John Doe"
				phoneJane := "0690383372"
				phoneJohn := "0811111111"
				p.Contact = []fhir.PatientContact{
					{
						Name: &fhir.HumanName{Text: &jane, Use: &official},
						Telecom: []fhir.ContactPoint{{
							System: &phone,
							Value:  &phoneJane,
						}},
					},
					{
						Name: &fhir.HumanName{Text: &john, Use: &official},
						Telecom: []fhir.ContactPoint{{
							System: &phone,
							Use:    &mobile,
							Value:  &phoneJohn,
						}},
					},
				}
				return p
			},
			check: func(t *testing.T, patient *Patient) {
				if len(patient.EmergencyContacts) != 2 {
					t.Fatalf("len(EmergencyContacts): got %d, want 2", len(patient.EmergencyContacts))
				}
				if patient.EmergencyContacts[0].Name != "Jane Smith" {
					t.Errorf("EmergencyContacts[0].Name: got %q, want Jane Smith", patient.EmergencyContacts[0].Name)
				}
				if len(patient.EmergencyContacts[0].Phones) != 1 || patient.EmergencyContacts[0].Phones[0] != "0690383372" {
					t.Errorf("EmergencyContacts[0].Phones: got %v, want [0690383372]", patient.EmergencyContacts[0].Phones)
				}
				if patient.EmergencyContacts[1].Name != "John Doe" {
					t.Errorf("EmergencyContacts[1].Name: got %q, want John Doe", patient.EmergencyContacts[1].Name)
				}
				if len(patient.EmergencyContacts[1].Phones) != 1 || patient.EmergencyContacts[1].Phones[0] != "0811111111" {
					t.Errorf("EmergencyContacts[1].Phones: got %v, want [0811111111]", patient.EmergencyContacts[1].Phones)
				}
			},
		},
		{
			name: "home address mapped",
			build: func() *fhir.Patient {
				p := minimalFHIRPatient()
				home := fhir.AddressUseHome
				line := "Jl. Example 1"
				city := "Jakarta"
				postal := "12345"
				country := "ID"
				p.Address = []fhir.Address{{
					Use:        &home,
					Line:       []string{line},
					City:       &city,
					PostalCode: &postal,
					Country:    &country,
				}}
				return p
			},
			check: func(t *testing.T, patient *Patient) {
				if patient.Address.Line != "Jl. Example 1" {
					t.Errorf("Address.Line: got %q, want Jl. Example 1", patient.Address.Line)
				}
				if patient.Address.City != "Jakarta" {
					t.Errorf("Address.City: got %q, want Jakarta", patient.Address.City)
				}
				if patient.Address.PostalCode != "12345" {
					t.Errorf("Address.PostalCode: got %q, want 12345", patient.Address.PostalCode)
				}
				if patient.Address.Country != "ID" {
					t.Errorf("Address.Country: got %q, want ID", patient.Address.Country)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.check(t, mapPatient(t, tt.build()))
		})
	}
}

func TestMapFHIRPatient_metaLastUpdateParsed(t *testing.T) {
	t.Parallel()

	src := loadFirstPatientFromBundle(t, "patient_search_name_birthdate_nik.json")
	patient := mapPatient(t, src)

	want, err := time.Parse(time.RFC3339, "2024-11-14T07:15:02.824423+00:00")
	if err != nil {
		t.Fatalf("parse want time: %v", err)
	}
	if patient.Meta == nil {
		t.Fatal("Meta is nil")
	}
	if !patient.Meta.LastUpdate.Equal(want) {
		t.Errorf("Meta.LastUpdate: got %v, want %v", patient.Meta.LastUpdate, want)
	}
}
