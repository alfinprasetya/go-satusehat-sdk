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

func (s *stubTokenProvider) GetToken(context.Context) (string, error) {
	s.calls++
	return s.token, s.err
}

func TestNewClient(t *testing.T) {
	t.Parallel()

	auth := &stubTokenProvider{token: "tok"}
	client := NewClient("org-123", "https://fhir.example", auth)

	if client.BaseURL != "https://fhir.example" {
		t.Errorf("BaseURL: got %q, want %q", client.BaseURL, "https://fhir.example")
	}
	if client.OrgID != "org-123" {
		t.Errorf("OrgID: got %q, want %q", client.OrgID, "org-123")
	}
	if client.Auth != auth {
		t.Error("Auth: expected same pointer as passed in")
	}
	if client.Patients == nil {
		t.Fatal("Patients is nil")
	}
	if client.Patients.client != client {
		t.Error("Patients.client: expected back-reference to Client")
	}
}

func TestClient_do_injectsAuthAndFHIRHeaders(t *testing.T) {
	t.Parallel()

	auth := &stubTokenProvider{token: "test-token"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization: got %q, want Bearer test-token", got)
		}
		if got := r.Header.Get("Accept"); got != "application/fhir+json" {
			t.Errorf("Accept: got %q, want application/fhir+json", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/fhir+json" {
			t.Errorf("Content-Type: got %q, want application/fhir+json", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("org-id", server.URL, auth)
	client.httpClient = server.Client()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/Patient", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	resp, err := client.do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	resp.Body.Close()

	if auth.calls != 1 {
		t.Errorf("GetToken calls: got %d, want 1", auth.calls)
	}
}

func TestClient_do_nilAuth(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("Authorization: got %q, want empty", got)
		}
		if got := r.Header.Get("Accept"); got != "application/fhir+json" {
			t.Errorf("Accept: got %q, want application/fhir+json", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/fhir+json" {
			t.Errorf("Content-Type: got %q, want application/fhir+json", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("org-id", server.URL, nil)
	client.httpClient = server.Client()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/Patient", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	resp, err := client.do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	resp.Body.Close()
}

func TestClient_do_authError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("token unavailable")
	auth := &stubTokenProvider{err: wantErr}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("HTTP server should not be called when auth fails")
	}))
	defer server.Close()

	client := NewClient("org-id", server.URL, auth)
	client.httpClient = server.Client()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/Patient", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	_, err = client.do(req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("errors.Is: got %v, want %v", err, wantErr)
	}
	if auth.calls != 1 {
		t.Errorf("GetToken calls: got %d, want 1", auth.calls)
	}
}
