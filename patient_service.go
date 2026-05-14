package satusehat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/alfinprasetya/go-satusehat-sdk/models"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

type PatientService struct {
	client *Client
}

type GetCompletePatientByNIK struct {
	Name      string
	Birthdate string
	NIK       string
}

func (s *PatientService) GetCompleteByNIK(ctx context.Context, params GetCompletePatientByNIK) (patient *models.Patient, err error) {
	reqUrl := fmt.Sprintf(
		"%s/Patient?name=%s&birthdate=%s&identifier=https://fhir.kemkes.go.id/id/nik|%s",
		s.client.BaseURL,
		url.QueryEscape(params.Name),
		url.QueryEscape(params.Birthdate),
		url.QueryEscape(params.NIK),
	)

	return s.get(ctx, reqUrl)
}

type GetCompletePatientByGender struct {
	Name      string
	Birthdate string
	Gender    models.Gender
}

func (s *PatientService) GetCompleteByGender(ctx context.Context, params GetCompletePatientByGender) (patient *models.Patient, err error) {
	reqUrl := fmt.Sprintf(
		"%s/Patient?name=%s&birthdate=%s&gender=%s",
		s.client.BaseURL,
		url.QueryEscape(params.Name),
		url.QueryEscape(params.Birthdate),
		url.QueryEscape(string(params.Gender)),
	)

	return s.get(ctx, reqUrl)
}

func (s *PatientService) GetPartial(ctx context.Context, nik string) (patient *models.Patient, err error) {
	reqUrl := fmt.Sprintf(
		"%s/Patient?identifier=https://fhir.kemkes.go.id/id/nik|%s",
		s.client.BaseURL,
		url.QueryEscape(nik),
	)

	return s.get(ctx, reqUrl)
}

func (s *PatientService) get(ctx context.Context, url string) (patient *models.Patient, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fhir api error: status %d", resp.StatusCode)
	}

	bundle := &fhir.Bundle{}
	if err := json.NewDecoder(resp.Body).Decode(bundle); err != nil {
		return nil, fmt.Errorf("failed to decode bundle: %w", err)
	}

	if len(bundle.Entry) == 0 {
		return nil, fmt.Errorf("patient not found")
	}

	fhirPatient := &fhir.Patient{}
	if err := json.Unmarshal(bundle.Entry[0].Resource, fhirPatient); err != nil {
		return nil, fmt.Errorf("failed to decode patient resource: %w", err)
	}

	patient = &models.Patient{}
	if err := models.MapFHIRPatient(fhirPatient, patient); err != nil {
		return nil, err
	}

	return patient, nil
}
