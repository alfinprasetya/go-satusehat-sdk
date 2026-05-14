package satusehat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TokenProvider interface {
	GetToken(ctx context.Context) (string, error)
}

type OAuth2Provider struct {
	mu     sync.RWMutex
	token  string
	expiry time.Time

	tokenURL     string
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func NewOAuth2Provider(tokenURL, clientID, clientSecret string) *OAuth2Provider {
	return &OAuth2Provider{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *OAuth2Provider) GetToken(ctx context.Context) (string, error) {
	p.mu.RLock()
	if p.token != "" && time.Now().Add(time.Minute).Before(p.expiry) {
		defer p.mu.RUnlock()
		return p.token, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check lock
	if p.token != "" && time.Now().Add(time.Minute).Before(p.expiry) {
		return p.token, nil
	}

	newToken, expiresIn, err := p.refreshToken(ctx)
	if err != nil {
		return "", fmt.Errorf("auth refresh failed: %w", err)
	}

	p.token = newToken
	p.expiry = time.Now().Add(time.Duration(expiresIn) * time.Second)

	return p.token, nil
}

func (p *OAuth2Provider) refreshToken(ctx context.Context) (string, int, error) {
	tokenURL := p.tokenURL + "/accesstoken?grant_type=client_credentials"

	data := url.Values{}
	data.Set("client_id", p.clientID)
	data.Set("client_secret", p.clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("auth server returned status %d", resp.StatusCode)
	}

	var res struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   string `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", 0, err
	}

	expiresIn := 0
	if expiresIn, err = strconv.Atoi(res.ExpiresIn); err != nil {
		return "", 0, err
	}

	return res.AccessToken, expiresIn, nil
}
