package satusehat

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubTokenProvider struct {
	token string
	err   error
	calls int
}

func (p *stubTokenProvider) GetToken(ctx context.Context) (string, error) {
	p.calls++
	return p.token, p.err
}

func TestNewClientUsesEnvironmentConfig(t *testing.T) {
	orgID := envOrDefault("ORG_ID", "test-org")
	baseURL := envOrDefault("FHIR_URL", "https://fhir.example.test")
	auth := &stubTokenProvider{token: "env-token"}

	client := NewClient(orgID, baseURL, auth)

	if client.OrgID != orgID {
		t.Fatalf("OrgID = %s, want %s", client.OrgID, orgID)
	}
	if client.BaseURL != baseURL {
		t.Fatalf("BaseURL = %s, want %s", client.BaseURL, baseURL)
	}
	if client.Auth != auth {
		t.Fatal("Auth provider was not stored")
	}
	if client.Patients == nil {
		t.Fatal("Patients service is nil")
	}
	if client.Patients.client != client {
		t.Fatal("Patients service does not reference its parent client")
	}
}

func TestClientDoAddsAuthAndFHIRHeaders(t *testing.T) {
	token := envOrDefault("SATUSEHAT_TEST_BEARER_TOKEN", "mock-bearer-token")
	auth := &stubTokenProvider{token: token}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer "+token {
			t.Fatalf("Authorization = %s, want Bearer %s", got, token)
		}
		if got := r.Header.Get("Accept"); got != "application/fhir+json" {
			t.Fatalf("Accept = %s, want application/fhir+json", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/fhir+json" {
			t.Fatalf("Content-Type = %s, want application/fhir+json", got)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(envOrDefault("ORG_ID", "test-org"), server.URL, auth)
	client.httpClient = server.Client()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/metadata", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext returned error: %v", err)
	}

	resp, err := client.do(req)
	if err != nil {
		t.Fatalf("do returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	if auth.calls != 1 {
		t.Fatalf("token provider calls = %d, want 1", auth.calls)
	}
}

func TestClientDoReturnsAuthError(t *testing.T) {
	wantErr := errors.New("token failed")
	client := NewClient("test-org", "https://fhir.example.test", &stubTokenProvider{err: wantErr})

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://fhir.example.test/Patient", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext returned error: %v", err)
	}

	_, err = client.do(req)
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
}
