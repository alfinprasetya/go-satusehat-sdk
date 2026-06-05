package satusehat

import (
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	BaseURL    string
	OrgID      string
	Auth       TokenProvider

	// Sub-services
	Patients *PatientService
}

func NewClient(orgID, baseURL string, auth TokenProvider) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		BaseURL:    baseURL,
		OrgID:      orgID,
		Auth:       auth,
	}

	// Initialize modules
	c.Patients = &PatientService{client: c}

	return c
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	// Inject Auth if provider is present
	if c.Auth != nil {
		token, err := c.Auth.GetToken(req.Context())
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Standard FHIR Headers
	req.Header.Set("Accept", "application/fhir+json")
	req.Header.Set("Content-Type", "application/fhir+json")

	return c.httpClient.Do(req)
}
