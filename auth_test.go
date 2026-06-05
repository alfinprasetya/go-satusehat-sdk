package satusehat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOAuth2ProviderGetTokenRefreshesAndCaches(t *testing.T) {
	clientID := envOrDefault("CLIENT_ID", "test-client")
	clientSecret := envOrDefault("CLIENT_SECRET", "test-secret")

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++

		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want %s", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/accesstoken" {
			t.Fatalf("path = %s, want /accesstoken", r.URL.Path)
		}
		if got := r.URL.Query().Get("grant_type"); got != "client_credentials" {
			t.Fatalf("grant_type = %s, want client_credentials", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Fatalf("Content-Type = %s, want application/x-www-form-urlencoded", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm returned error: %v", err)
		}
		if got := r.Form.Get("client_id"); got != clientID {
			t.Fatalf("client_id = %s, want %s", got, clientID)
		}
		if got := r.Form.Get("client_secret"); got != clientSecret {
			t.Fatalf("client_secret = %s, want %s", got, clientSecret)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"access_token": "mock-token",
			"expires_in":   "3600",
		}); err != nil {
			t.Fatalf("Encode returned error: %v", err)
		}
	}))
	defer server.Close()

	provider := NewOAuth2Provider(server.URL, clientID, clientSecret)
	provider.httpClient = server.Client()

	ctx := context.Background()
	token, err := provider.GetToken(ctx)
	if err != nil {
		t.Fatalf("GetToken returned error: %v", err)
	}
	if token != "mock-token" {
		t.Fatalf("token = %s, want mock-token", token)
	}

	cachedToken, err := provider.GetToken(ctx)
	if err != nil {
		t.Fatalf("cached GetToken returned error: %v", err)
	}
	if cachedToken != "mock-token" {
		t.Fatalf("cached token = %s, want mock-token", cachedToken)
	}
	if requests != 1 {
		t.Fatalf("token endpoint requests = %d, want 1", requests)
	}
	if provider.expiry.Before(time.Now().Add(59 * time.Minute)) {
		t.Fatalf("expiry = %s, want at least 59 minutes in the future", provider.expiry.Format(time.RFC3339))
	}
}

func TestOAuth2ProviderGetTokenFromEnvironment(t *testing.T) {
	tokenURL := envOrSkip(t, "AUTH_URL")
	clientID := envOrSkip(t, "CLIENT_ID")
	clientSecret := envOrSkip(t, "CLIENT_SECRET")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	token, err := NewOAuth2Provider(tokenURL, clientID, clientSecret).GetToken(ctx)
	if err != nil {
		t.Fatalf("GetToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}
}
