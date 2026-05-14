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

type PatientSearchParams struct {
	Name      *string
	Birthdate *string
	Gender    *models.Gender
	NIK       *string
}

func (s *PatientService) Search(ctx context.Context, params PatientSearchParams) (patient *models.Patient, err error) {
	reqUrl := fmt.Sprintf(
		"%s/Patient",
		s.client.BaseURL,
	)

	switch {
	case params.Name != nil &&
		params.Birthdate != nil &&
		params.NIK != nil:

		reqUrl += fmt.Sprintf(
			"?name=%s&birthdate=%s&identifier=https://fhir.kemkes.go.id/id/nik|%s",
			url.QueryEscape(*params.Name),
			url.QueryEscape(*params.Birthdate),
			url.QueryEscape(*params.NIK),
		)

	case params.Name != nil &&
		params.Birthdate != nil &&
		params.Gender != nil:

		reqUrl += fmt.Sprintf(
			"?name=%s&birthdate=%s&gender=%s",
			url.QueryEscape(*params.Name),
			url.QueryEscape(*params.Birthdate),
			url.QueryEscape(string(*params.Gender)),
		)

	case params.Name != nil &&
		params.NIK != nil:

		reqUrl += fmt.Sprintf(
			"?name=%s&identifier=https://fhir.kemkes.go.id/id/nik|%s",
			url.QueryEscape(*params.Name),
			url.QueryEscape(*params.NIK),
		)

	case params.NIK != nil:

		reqUrl += fmt.Sprintf(
			"?identifier=https://fhir.kemkes.go.id/id/nik|%s",
			url.QueryEscape(*params.NIK),
		)

	default:
		return nil, fmt.Errorf("invalid search params")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
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
