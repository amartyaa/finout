package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TokenResponse represents the Azure AD OAuth2 token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// AuthClient handles Azure AD authentication via Service Principal
type AuthClient struct {
	TenantID     string
	ClientID     string
	ClientSecret string
	httpClient   *http.Client
}

// NewAuthClient creates a new Azure AD auth client
func NewAuthClient(tenantID, clientID, clientSecret string) *AuthClient {
	return &AuthClient{
		TenantID:     tenantID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// GetToken obtains a bearer token using client credentials flow
func (a *AuthClient) GetToken(scope string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", a.TenantID)

	data := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {a.ClientID},
		"client_secret": {a.ClientSecret},
		"scope":         {scope},
	}

	resp, err := a.httpClient.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Azure AD token request failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// ValidateCredentials tests that the Service Principal can authenticate
func (a *AuthClient) ValidateCredentials() error {
	_, err := a.GetToken("https://management.azure.com/.default")
	return err
}
