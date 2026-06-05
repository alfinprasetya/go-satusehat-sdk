package satusehat

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewOAuth2Provider(t *testing.T) {
	t.Parallel()

	p := NewOAuth2Provider("https://auth.example", "client-id", "client-secret")
	if p == nil {
		t.Fatal("NewOAuth2Provider returned nil")
	}
	if p.tokenURL != "https://auth.example" {
		t.Errorf("tokenURL: got %q, want %q", p.tokenURL, "https://auth.example")
	}
	if p.clientID != "client-id" {
		t.Errorf("clientID: got %q, want %q", p.clientID, "client-id")
	}
	if p.clientSecret != "client-secret" {
		t.Errorf("clientSecret: got %q, want %q", p.clientSecret, "client-secret")
	}
	if p.httpClient == nil {
		t.Error("httpClient is nil")
	}
}

func TestOAuth2Provider_GetToken_success(t *testing.T) {
	t.Parallel()

	const (
		clientID     = "test-client"
		clientSecret = "test-secret"
		wantToken    = "access-token-123"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: got %s, want POST", r.Method)
		}
		if r.URL.Path != "/accesstoken" {
			t.Errorf("path: got %s, want /accesstoken", r.URL.Path)
		}
		if got := r.URL.Query().Get("grant_type"); got != "client_credentials" {
			t.Errorf("grant_type: got %q, want client_credentials", got)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type: got %q, want application/x-www-form-urlencoded", ct)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		vals, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatalf("ParseQuery: %v", err)
		}
		if vals.Get("client_id") != clientID {
			t.Errorf("client_id: got %q, want %q", vals.Get("client_id"), clientID)
		}
		if vals.Get("client_secret") != clientSecret {
			t.Errorf("client_secret: got %q, want %q", vals.Get("client_secret"), clientSecret)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"` + wantToken + `","expires_in":"3600"}`))
	}))
	defer server.Close()

	provider := NewOAuth2Provider(server.URL, clientID, clientSecret)
	provider.httpClient = server.Client()

	token, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	if token != wantToken {
		t.Errorf("token: got %q, want %q", token, wantToken)
	}
}

func TestOAuth2Provider_GetToken_caches(t *testing.T) {
	t.Parallel()

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"cached-token","expires_in":"3600"}`))
	}))
	defer server.Close()

	provider := NewOAuth2Provider(server.URL, "id", "secret")
	provider.httpClient = server.Client()

	ctx := context.Background()
	if _, err := provider.GetToken(ctx); err != nil {
		t.Fatalf("first GetToken: %v", err)
	}
	if _, err := provider.GetToken(ctx); err != nil {
		t.Fatalf("second GetToken: %v", err)
	}

	if got := requests.Load(); got != 1 {
		t.Errorf("token endpoint calls: got %d, want 1", got)
	}
}

func TestOAuth2Provider_GetToken_refreshesNearExpiry(t *testing.T) {
	t.Parallel()

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-token","expires_in":"3600"}`))
	}))
	defer server.Close()

	provider := NewOAuth2Provider(server.URL, "id", "secret")
	provider.httpClient = server.Client()

	provider.token = "stale-token"
	provider.expiry = time.Now().Add(30 * time.Second)

	token, err := provider.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	if token != "new-token" {
		t.Errorf("token: got %q, want %q", token, "new-token")
	}
	if got := requests.Load(); got != 1 {
		t.Errorf("token endpoint calls: got %d, want 1", got)
	}
}

func TestOAuth2Provider_GetToken_authError(t *testing.T) {
	t.Parallel()

	const oauthErrorBody = `{
    "ErrorCode": "invalid_client",
    "Error": "ClientId is Invalid"
}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(oauthErrorBody))
	}))
	defer server.Close()

	provider := NewOAuth2Provider(server.URL, "bad-id", "bad-secret")
	provider.httpClient = server.Client()

	_, err := provider.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "auth refresh failed") {
		t.Errorf("error: got %q, want substring auth refresh failed", err.Error())
	}

	var oauthErr *OAuthError
	if !errors.As(err, &oauthErr) {
		t.Fatalf("errors.As OAuthError: got %T", err)
	}
	if !errors.Is(err, ErrInvalidClient) {
		t.Errorf("errors.Is: got %v, want ErrInvalidClient", err)
	}
}

func TestOAuth2Provider_GetToken_invalidBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{
			name: "non-JSON",
			body: "not json",
		},
		{
			name: "invalid expires_in",
			body: `{"access_token":"tok","expires_in":"nope"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			provider := NewOAuth2Provider(server.URL, "id", "secret")
			provider.httpClient = server.Client()

			_, err := provider.GetToken(context.Background())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "auth refresh failed") {
				t.Errorf("error: got %q, want substring auth refresh failed", err.Error())
			}
		})
	}
}
