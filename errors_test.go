package satusehat

import (
	"errors"
	"net/http"
	"testing"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

func TestParseAPIError_PostmanFixtures(t *testing.T) {
	t.Parallel()

	headerWithRequestID := http.Header{}
	headerWithRequestID.Set("X-Request-Id", "bfad11c6-ad78-4ca7-ba43-10ccd943677f")

	noAPIRoleHeader := http.Header{}
	noAPIRoleHeader.Set("WWW-Authenticate", `Bearer realm="null",error="invalid_token",error_description="keymanagement.service.InvalidAPICallAsNoApiProductMatchFound: Invalid API call as no apiproduct match found"`)
	noAPIRoleHeader.Set("X-Request-Id", "f854636a-ce2f-90b4-ab0e-28f5af4a8e26")

	htmlHeader := http.Header{}
	htmlHeader.Set("Content-Type", "text/html; charset=UTF-8")

	plainHeader := http.Header{}
	plainHeader.Set("Content-Type", "text/plain")

	tests := []struct {
		name       string
		statusCode int
		header     http.Header
		body       string
		wantType   string
		wantIs     error
		check      func(t *testing.T, err error)
	}{
		{
			name:       "Error 200 - Consent Error",
			statusCode: http.StatusOK,
			header:     headerWithRequestID,
			body: `{
    "resourceType": "OperationOutcome",
    "issue": [
        {
            "severity": "information",
            "code": "suppressed",
            "details": {
                "text": "The operation did not return any information due to consent or privacy rules."
            }
        }
    ]
}`,
			wantType: "FHIRError",
			wantIs:   ErrConsentSuppressed,
			check: func(t *testing.T, err error) {
				t.Helper()
				var fhirErr *FHIRError
				if !errors.As(err, &fhirErr) {
					t.Fatal("expected FHIRError")
				}
				if fhirErr.RequestID != "bfad11c6-ad78-4ca7-ba43-10ccd943677f" {
					t.Errorf("RequestID: got %q", fhirErr.RequestID)
				}
				if len(fhirErr.Outcome.Issue) != 1 {
					t.Fatalf("len(Issue): got %d, want 1", len(fhirErr.Outcome.Issue))
				}
				issue := fhirErr.Outcome.Issue[0]
				if issue.Code != fhir.IssueTypeSuppressed {
					t.Errorf("issue code: got %v, want suppressed", issue.Code)
				}
				if issue.Details == nil || issue.Details.Text == nil {
					t.Fatal("expected issue details text")
				}
				if *issue.Details.Text != "The operation did not return any information due to consent or privacy rules." {
					t.Errorf("details text: got %q", *issue.Details.Text)
				}
			},
		},
		{
			name:       "Error 400 - Path URL and Payload ID mismatch",
			statusCode: http.StatusBadRequest,
			body: `{
    "issue": [
        {
            "code": "structure",
            "details": {
                "text": "resource_mismatch"
            },
            "diagnostics": "resource id mismatch, id of specified resource: 3dedcec9-885d-435e-9ac5-58853cb216, id from payload: 3dedcec9-885d-435e-9ac5-58853cb216bb",
            "severity": "error"
        }
    ],
    "resourceType": "OperationOutcome"
}`,
			wantType: "FHIRError",
			wantIs:   ErrResourceMismatch,
			check: func(t *testing.T, err error) {
				t.Helper()
				var fhirErr *FHIRError
				if !errors.As(err, &fhirErr) {
					t.Fatal("expected FHIRError")
				}
				issue := fhirErr.Outcome.Issue[0]
				if issue.Diagnostics == nil || *issue.Diagnostics == "" {
					t.Fatal("expected diagnostics to be preserved")
				}
				if issue.Code != fhir.IssueTypeStructure {
					t.Errorf("issue code: got %v, want structure", issue.Code)
				}
			},
		},
		{
			name:       "Error 400 - Belum OPTIN",
			statusCode: http.StatusBadRequest,
			body: `{
    "resourceType": "OperationOutcome",
    "issue": [
        {
            "severity": "error",
            "code": "forbidden",
            "details": {
                "text": "Operation cannot be performed due to business rules, consent or privacy rules, or access permission constraints."
            }
        }
    ]
}`,
			wantType: "FHIRError",
			wantIs:   ErrForbidden,
		},
		{
			name:       "Error 400 - Error JSON",
			statusCode: http.StatusBadRequest,
			body: `{
    "issue": [
        {
            "code": "structure",
            "details": {
                "text": "unparseable_resource"
            },
            "diagnostics": "missing required field \"status\"",
            "expression": [
                "Bundle.entry[0].resource.serviceRequest"
            ],
            "severity": "error"
        }
    ],
    "resourceType": "OperationOutcome"
}`,
			wantType: "FHIRError",
			wantIs:   ErrUnparseableResource,
			check: func(t *testing.T, err error) {
				t.Helper()
				var fhirErr *FHIRError
				if !errors.As(err, &fhirErr) {
					t.Fatal("expected FHIRError")
				}
				issue := fhirErr.Outcome.Issue[0]
				if len(issue.Expression) != 1 || issue.Expression[0] != "Bundle.entry[0].resource.serviceRequest" {
					t.Errorf("expression: got %v", issue.Expression)
				}
			},
		},
		{
			name:       "Error 400 - invalid_query birthdate",
			statusCode: http.StatusBadRequest,
			body: `{
    "resourceType": "OperationOutcome",
    "issue": [
        {
            "severity": "error",
            "code": "value",
            "details": {
                "text": "invalid_query"
            },
            "diagnostics": "error parsing date \"2024-07-021\": invalid dateTime: 2024-07-021"
        }
    ]
}`,
			wantType: "FHIRError",
			wantIs:   ErrInvalidQuery,
		},
		{
			name:       "Error 401 - Error no API role",
			statusCode: http.StatusUnauthorized,
			header:     noAPIRoleHeader,
			body: `{
    "fault": {
        "faultstring": "Invalid API call as no apiproduct match found",
        "detail": {
            "errorcode": "keymanagement.service.InvalidAPICallAsNoApiProductMatchFound"
        }
    }
}`,
			wantType: "GatewayError",
			wantIs:   ErrNoAPIProductMatch,
			check: func(t *testing.T, err error) {
				t.Helper()
				var gatewayErr *GatewayError
				if !errors.As(err, &gatewayErr) {
					t.Fatal("expected GatewayError")
				}
				if gatewayErr.FaultString != "Invalid API call as no apiproduct match found" {
					t.Errorf("FaultString: got %q", gatewayErr.FaultString)
				}
				if gatewayErr.WWWAuthenticate == "" {
					t.Error("expected WWW-Authenticate header to be preserved")
				}
			},
		},
		{
			name:       "Error 401 - Invalid ClientID",
			statusCode: http.StatusUnauthorized,
			body: `{
    "ErrorCode": "invalid_client",
    "Error": "ClientId is Invalid"
}`,
			wantType: "OAuthError",
			wantIs:   ErrInvalidClient,
			check: func(t *testing.T, err error) {
				t.Helper()
				var oauthErr *OAuthError
				if !errors.As(err, &oauthErr) {
					t.Fatal("expected OAuthError")
				}
				if oauthErr.Message != "ClientId is Invalid" {
					t.Errorf("Message: got %q", oauthErr.Message)
				}
			},
		},
		{
			name:       "Error 401 - Invalid Token",
			statusCode: http.StatusUnauthorized,
			body: `{
    "fault": {
        "faultstring": "Invalid access token",
        "detail": {
            "errorcode": "oauth.v2.InvalidAccessToken"
        }
    }
}`,
			wantType: "GatewayError",
			wantIs:   ErrInvalidAccessToken,
		},
		{
			name:       "Error 401 - Access Token expired",
			statusCode: http.StatusUnauthorized,
			body: `{
    "fault": {
        "faultstring": "Access Token expired",
        "detail": {
            "errorcode": "keymanagement.service.access_token_expired"
        }
    }
}`,
			wantType: "GatewayError",
			wantIs:   ErrAccessTokenExpired,
		},
		{
			name:       "Error 403 - Forbidden",
			statusCode: http.StatusForbidden,
			header:     htmlHeader,
			body:       `<!doctype html><meta charset="utf-8"><meta name=viewport content="width=device-width, initial-scale=1"><title>403</title>403 Forbidden`,
			wantType:   "HTMLError",
			check: func(t *testing.T, err error) {
				t.Helper()
				var htmlErr *HTMLError
				if !errors.As(err, &htmlErr) {
					t.Fatal("expected HTMLError")
				}
				if htmlErr.Message != "403 Forbidden" {
					t.Errorf("Message: got %q, want %q", htmlErr.Message, "403 Forbidden")
				}
			},
		},
		{
			name:       "Error 404 - Wrong URL",
			statusCode: http.StatusNotFound,
			body:       "",
			wantType:   "PlainHTTPError",
		},
		{
			name:       "Error 503 - API error",
			statusCode: http.StatusServiceUnavailable,
			header:     plainHeader,
			body:       "no healthy upstream",
			wantType:   "PlainHTTPError",
			check: func(t *testing.T, err error) {
				t.Helper()
				var plainErr *PlainHTTPError
				if !errors.As(err, &plainErr) {
					t.Fatal("expected PlainHTTPError")
				}
				if plainErr.Message != "no healthy upstream" {
					t.Errorf("Message: got %q", plainErr.Message)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			header := tt.header
			if header == nil {
				header = http.Header{}
			}

			err := ParseAPIError(tt.statusCode, header, []byte(tt.body))
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if tt.wantType != "" {
				assertErrorType(t, err, tt.wantType)
			}

			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Errorf("errors.Is: got %v, want %v", err, tt.wantIs)
			}

			var apiErr *APIError
			switch e := err.(type) {
			case *FHIRError:
				apiErr = &e.APIError
			case *GatewayError:
				apiErr = &e.APIError
			case *OAuthError:
				apiErr = &e.APIError
			case *HTMLError:
				apiErr = &e.APIError
			case *PlainHTTPError:
				apiErr = &e.APIError
			default:
				t.Fatalf("unexpected error type %T", err)
			}

			if apiErr.StatusCode != tt.statusCode {
				t.Errorf("StatusCode: got %d, want %d", apiErr.StatusCode, tt.statusCode)
			}
			if string(apiErr.Body) != tt.body {
				t.Errorf("Body not preserved: got %q, want %q", string(apiErr.Body), tt.body)
			}

			if tt.check != nil {
				tt.check(t, err)
			}
		})
	}
}

func assertErrorType(t *testing.T, err error, want string) {
	t.Helper()

	switch want {
	case "FHIRError":
		var target *FHIRError
		if !errors.As(err, &target) {
			t.Fatalf("error type: got %T, want *FHIRError", err)
		}
	case "GatewayError":
		var target *GatewayError
		if !errors.As(err, &target) {
			t.Fatalf("error type: got %T, want *GatewayError", err)
		}
	case "OAuthError":
		var target *OAuthError
		if !errors.As(err, &target) {
			t.Fatalf("error type: got %T, want *OAuthError", err)
		}
	case "HTMLError":
		var target *HTMLError
		if !errors.As(err, &target) {
			t.Fatalf("error type: got %T, want *HTMLError", err)
		}
	case "PlainHTTPError":
		var target *PlainHTTPError
		if !errors.As(err, &target) {
			t.Fatalf("error type: got %T, want *PlainHTTPError", err)
		}
	default:
		t.Fatalf("unknown wantType %q", want)
	}
}
