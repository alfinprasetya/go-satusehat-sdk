package satusehat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

// Domain errors (distinct from transport/API errors).
var ErrPatientNotFound = errors.New("patient not found")

// FHIR OperationOutcome sentinel errors.
var (
	ErrConsentSuppressed   = errors.New("consent or privacy rules suppressed the response")
	ErrResourceMismatch    = errors.New("resource id mismatch between URL and payload")
	ErrUnparseableResource = errors.New("request body contains an unparseable resource")
	ErrForbidden           = errors.New("operation forbidden by business rules, consent, or access permissions")
	ErrInvalidQuery        = errors.New("invalid FHIR search or request query parameter")
)

// Gateway (Apigee) sentinel errors.
var (
	ErrNoAPIProductMatch  = errors.New("invalid API call: no API product match found")
	ErrInvalidAccessToken = errors.New("invalid access token")
	ErrAccessTokenExpired = errors.New("access token expired")
)

// OAuth sentinel errors.
var ErrInvalidClient = errors.New("invalid OAuth client credentials")

// APIError is the common base for all SATUSEHAT API error responses.
type APIError struct {
	StatusCode int
	RequestID  string
	Body       []byte
}

func (e *APIError) Error() string {
	if len(e.Body) > 0 {
		return fmt.Sprintf("satusehat api error: status %d: %s", e.StatusCode, string(e.Body))
	}
	return fmt.Sprintf("satusehat api error: status %d", e.StatusCode)
}

// FHIRError wraps a parsed FHIR OperationOutcome.
type FHIRError struct {
	APIError
	Outcome fhir.OperationOutcome
}

func (e *FHIRError) Error() string {
	return fmt.Sprintf("satusehat fhir error: status %d: %s", e.StatusCode, formatOperationOutcomeIssues(e.Outcome))
}

func (e *FHIRError) Is(target error) bool {
	switch target {
	case ErrConsentSuppressed:
		return e.StatusCode == http.StatusOK && e.hasIssueCode(fhir.IssueTypeSuppressed)
	case ErrResourceMismatch:
		return e.hasDetailText("resource_mismatch")
	case ErrUnparseableResource:
		return e.hasDetailText("unparseable_resource")
	case ErrForbidden:
		return e.hasIssueCode(fhir.IssueTypeForbidden)
	case ErrInvalidQuery:
		return e.hasDetailText("invalid_query")
	default:
		return false
	}
}

func (e *FHIRError) hasIssueCode(code fhir.IssueType) bool {
	for _, issue := range e.Outcome.Issue {
		if issue.Code == code {
			return true
		}
	}
	return false
}

func (e *FHIRError) hasDetailText(text string) bool {
	for _, issue := range e.Outcome.Issue {
		if issue.Details != nil && issue.Details.Text != nil && *issue.Details.Text == text {
			return true
		}
	}
	return false
}

// GatewayError represents an Apigee gateway fault response.
type GatewayError struct {
	APIError
	FaultString     string
	ErrorCode       string
	WWWAuthenticate string
}

func (e *GatewayError) Error() string {
	if e.ErrorCode != "" {
		return fmt.Sprintf("satusehat gateway error: status %d: %s (%s)", e.StatusCode, e.FaultString, e.ErrorCode)
	}
	return fmt.Sprintf("satusehat gateway error: status %d: %s", e.StatusCode, e.FaultString)
}

func (e *GatewayError) Is(target error) bool {
	switch target {
	case ErrNoAPIProductMatch:
		return strings.Contains(e.ErrorCode, "InvalidAPICallAsNoApiProductMatchFound")
	case ErrInvalidAccessToken:
		return strings.Contains(e.ErrorCode, "oauth.v2.InvalidAccessToken")
	case ErrAccessTokenExpired:
		return strings.Contains(e.ErrorCode, "access_token_expired")
	default:
		return false
	}
}

// OAuthError represents an OAuth token endpoint error response.
type OAuthError struct {
	APIError
	Code    string
	Message string
}

func (e *OAuthError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("satusehat oauth error: status %d: %s (%s)", e.StatusCode, e.Message, e.Code)
	}
	return fmt.Sprintf("satusehat oauth error: status %d: %s", e.StatusCode, e.Code)
}

func (e *OAuthError) Is(target error) bool {
	return target == ErrInvalidClient && e.Code == "invalid_client"
}

// HTMLError represents a non-JSON HTML error page (e.g. 403 Forbidden).
type HTMLError struct {
	APIError
	Message string
}

func (e *HTMLError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("satusehat html error: status %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("satusehat html error: status %d", e.StatusCode)
}

// PlainHTTPError represents a plain-text or empty HTTP error body.
type PlainHTTPError struct {
	APIError
	Message string
}

func (e *PlainHTTPError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("satusehat http error: status %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("satusehat http error: status %d", e.StatusCode)
}

// ParseAPIError builds a typed error from an HTTP status, headers, and response body.
func ParseAPIError(statusCode int, header http.Header, body []byte) error {
	base := newAPIError(statusCode, header, body)

	if isOperationOutcomeBody(body) {
		outcome, err := fhir.UnmarshalOperationOutcome(body)
		if err != nil {
			return &FHIRError{
				APIError: base,
				Outcome:  fhir.OperationOutcome{},
			}
		}
		return &FHIRError{
			APIError: base,
			Outcome:  outcome,
		}
	}

	if gateway, ok := parseGatewayFault(body); ok {
		gateway.APIError = base
		gateway.WWWAuthenticate = header.Get("WWW-Authenticate")
		return gateway
	}

	if oauth, ok := parseOAuthError(body); ok {
		oauth.APIError = base
		return oauth
	}

	contentType := header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		return &HTMLError{
			APIError: base,
			Message:  stripHTML(string(body)),
		}
	}

	if len(body) > 0 {
		return &PlainHTTPError{
			APIError: base,
			Message:  strings.TrimSpace(string(body)),
		}
	}

	return &PlainHTTPError{APIError: base}
}

func newAPIError(statusCode int, header http.Header, body []byte) APIError {
	bodyCopy := make([]byte, len(body))
	copy(bodyCopy, body)

	return APIError{
		StatusCode: statusCode,
		RequestID:  header.Get("X-Request-Id"),
		Body:       bodyCopy,
	}
}

type jsonShapeProbe struct {
	ResourceType string          `json:"resourceType"`
	Issue        json.RawMessage `json:"issue"`
	Fault        json.RawMessage `json:"fault"`
	ErrorCode    string          `json:"ErrorCode"`
}

func isOperationOutcomeBody(body []byte) bool {
	body = bytes.TrimSpace(body)
	if len(body) == 0 || body[0] != '{' {
		return false
	}

	var probe jsonShapeProbe
	if err := json.Unmarshal(body, &probe); err != nil {
		return false
	}

	if probe.ResourceType == "OperationOutcome" {
		return true
	}

	return len(bytes.TrimSpace(probe.Issue)) > 2 && probe.Issue[0] == '['
}

type gatewayFaultBody struct {
	Fault struct {
		FaultString string `json:"faultstring"`
		Detail      struct {
			ErrorCode string `json:"errorcode"`
		} `json:"detail"`
	} `json:"fault"`
}

func parseGatewayFault(body []byte) (*GatewayError, bool) {
	var parsed gatewayFaultBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, false
	}
	if parsed.Fault.FaultString == "" && parsed.Fault.Detail.ErrorCode == "" {
		return nil, false
	}
	return &GatewayError{
		FaultString: parsed.Fault.FaultString,
		ErrorCode:   parsed.Fault.Detail.ErrorCode,
	}, true
}

type oauthErrorBody struct {
	ErrorCode string `json:"ErrorCode"`
	Error     string `json:"Error"`
}

func parseOAuthError(body []byte) (*OAuthError, bool) {
	var parsed oauthErrorBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, false
	}
	if parsed.ErrorCode == "" {
		return nil, false
	}
	return &OAuthError{
		Code:    parsed.ErrorCode,
		Message: parsed.Error,
	}, true
}

func formatOperationOutcomeIssues(outcome fhir.OperationOutcome) string {
	if len(outcome.Issue) == 0 {
		return "operation outcome"
	}

	parts := make([]string, 0, len(outcome.Issue))
	for _, issue := range outcome.Issue {
		var part strings.Builder
		part.WriteString(issue.Code.Code())
		if issue.Details != nil && issue.Details.Text != nil && *issue.Details.Text != "" {
			part.WriteString(": ")
			part.WriteString(*issue.Details.Text)
		}
		if issue.Diagnostics != nil && *issue.Diagnostics != "" {
			part.WriteString(" (")
			part.WriteString(*issue.Diagnostics)
			part.WriteByte(')')
		}
		if len(issue.Expression) > 0 {
			part.WriteString(" [")
			part.WriteString(strings.Join(issue.Expression, ", "))
			part.WriteByte(']')
		}
		parts = append(parts, part.String())
	}
	return strings.Join(parts, "; ")
}

func stripHTML(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.LastIndex(s, ">"); idx >= 0 && idx < len(s)-1 {
		return strings.TrimSpace(s[idx+1:])
	}
	return s
}
