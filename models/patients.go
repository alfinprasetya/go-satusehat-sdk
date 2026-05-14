package models

import (
	"strings"
	"time"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

type Patient struct {
	ID string

	NIK       string
	IHSNumber string

	Active bool

	FullName   string
	Gender     Gender
	BirthDate  string
	BirthPlace string

	Contact Contact

	Address Address

	Citizenship CitizenshipStatus
	Marital     MaritalStatus

	EmergencyContacts []EmergencyContact

	PreferredLanguage string

	Deceased bool

	Meta *ResourceMeta
}

type Gender string

const (
	GenderMale    Gender = "male"
	GenderFemale  Gender = "female"
	GenderOther   Gender = "other"
	GenderUnknown Gender = "unknown"
)

type MaritalStatus string

const (
	MaritalSingle    MaritalStatus = "single"
	MaritalUnmarried MaritalStatus = "unmarried"
	MaritalMarried   MaritalStatus = "married"
	MaritalDivorced  MaritalStatus = "divorced"
)

type CitizenshipStatus string

const (
	CitizenshipWNI CitizenshipStatus = "WNI"
	CitizenshipWNA CitizenshipStatus = "WNA"
)

type Contact struct {
	MobilePhone string
	HomePhone   string
	Email       string
}

type Address struct {
	Line       string
	City       string
	PostalCode string
	Country    string

	// Administrative AdministrativeCode
}

// type AdministrativeCode struct {
// 	ProvinceCode string
// 	CityCode     string
// 	DistrictCode string
// 	VillageCode  string

// 	RT string
// 	RW string
// }

type EmergencyContact struct {
	Name   string
	Phones []string
}

type ContactRelationship string

type ResourceMeta struct {
	VersionID  string
	LastUpdate time.Time
}

func MapFHIRPatient(src *fhir.Patient, out *Patient) error {
	if src == nil {
		return nil
	}

	for _, identifier := range src.Identifier {
		if identifier.System != nil && *identifier.System == "https://fhir.kemkes.go.id/id/nik" {
			out.NIK = *identifier.Value
			continue
		}

		if identifier.System != nil && *identifier.System == "https://fhir.kemkes.go.id/id/ihs-number" {
			out.IHSNumber = *identifier.Value
			out.ID = *identifier.Value
			continue
		}
	}

	out.Active = *src.Active == true

	for _, name := range src.Name {
		if name.Use != nil && *name.Use == fhir.NameUseOfficial {
			if name.Text != nil {
				out.FullName = *name.Text
			}

			break
		}
	}

	if src.Gender != nil {
		out.Gender = Gender(src.Gender.String())
	}

	if src.BirthDate != nil {
		out.BirthDate = *src.BirthDate
	}

	for _, telecom := range src.Telecom {
		if telecom.System != nil && *telecom.System == fhir.ContactPointSystemPhone {
			if telecom.Value != nil {
				if telecom.Use != nil && *telecom.Use == fhir.ContactPointUseMobile {
					out.Contact.MobilePhone = *telecom.Value
				}

				if telecom.Use != nil && *telecom.Use == fhir.ContactPointUseHome {
					out.Contact.HomePhone = *telecom.Value
				}
			}

			continue
		}

		if telecom.System != nil && *telecom.System == fhir.ContactPointSystemEmail {
			if telecom.Value != nil {
				out.Contact.Email = *telecom.Value
			}

			continue
		}
	}

	for _, address := range src.Address {
		if address.Use != nil && *address.Use == fhir.AddressUseHome {
			if len(address.Line) > 0 {
				out.Address.Line = address.Line[0]
			}

			if address.City != nil {
				out.Address.City = *address.City
			}

			if address.PostalCode != nil {
				out.Address.PostalCode = *address.PostalCode
			}

			if address.Country != nil {
				out.Address.Country = *address.Country
			}

			break
		}
	}

	for _, ext := range src.Extension {
		if strings.Contains(ext.Url, "citizenshipStatus") {
			if ext.ValueCode != nil && *ext.ValueCode == "WNA" {
				out.Citizenship = CitizenshipWNA
			} else {
				out.Citizenship = CitizenshipWNI
			}

			continue
		}

		if strings.Contains(ext.Url, "birthPlace") {
			if ext.ValueAddress != nil {
				out.BirthPlace = *ext.ValueAddress.City
			}

			continue
		}
	}

	if src.MaritalStatus != nil && src.MaritalStatus.Text != nil {
		switch *src.MaritalStatus.Text {
		case "Single":
			out.Marital = MaritalSingle
		case "Unmarried":
			out.Marital = MaritalUnmarried
		case "Married":
			out.Marital = MaritalMarried
		case "Divorced":
			out.Marital = MaritalDivorced
		}
	}

	for _, communication := range src.Communication {
		if out.PreferredLanguage == "" && communication.Language.Text != nil {
			out.PreferredLanguage = *communication.Language.Text
		} else if communication.Preferred != nil && *communication.Preferred {
			out.PreferredLanguage = *communication.Language.Text
		}
	}

	var emergencyContacts []EmergencyContact
	for _, contact := range src.Contact {
		if contact.Name != nil && contact.Name.Text != nil {
			var phones []string
			for _, telecom := range contact.Telecom {
				if telecom.System != nil && *telecom.System == fhir.ContactPointSystemPhone {
					if telecom.Value != nil {
						phones = append(phones, *telecom.Value)
					}
				}
			}

			emergencyContacts = append(emergencyContacts, EmergencyContact{
				Name:   *contact.Name.Text,
				Phones: phones,
			})
		}
	}
	out.EmergencyContacts = emergencyContacts

	if src.DeceasedBoolean != nil {
		out.Deceased = *src.DeceasedBoolean
	}

	last, _ := time.Parse(time.RFC3339, *src.Meta.LastUpdated)
	out.Meta = &ResourceMeta{
		VersionID:  *src.Meta.VersionId,
		LastUpdate: last,
	}

	return nil
}
