package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type Provider struct {
	Config   *oauth2.Config
	UserInfo string // URL to fetch user info
}

// generates a secure random state string for OAuth2 CSRF protection
func GenerateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// returns the login URL to redirect the user to
func (p *Provider) GetLoginURL(state string) string {
	return p.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// Exchanges the callback code for an OAuth2 Token
func (p *Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.Config.Exchange(ctx, code)
}

// fetch JSON user profile from provider
func (p *Provider) FetchUserInfo(ctx context.Context, token *oauth2.Token) (map[string]interface{}, error) {
	client := p.Config.Client(ctx, token)
	resp, err := client.Get(p.UserInfo)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user info: status %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

// creates an OAuth2 provider for GitHub
func NewGitHubProvider(clientID, clientSecret, redirectURL string, scopes []string) *Provider {
	if len(scopes) == 0 {
		scopes = []string{"read:user", "user:email"}
	}
	return &Provider{
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     github.Endpoint,
		},
		UserInfo: "https://api.github.com/user",
	}
}

// creates an OAuth2 provider for Google
func NewGoogleProvider(clientID, clientSecret, redirectURL string, scopes []string) *Provider {
	if len(scopes) == 0 {
		scopes = []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"}
	}
	return &Provider{
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		},
		UserInfo: "https://www.googleapis.com/oauth2/v2/userinfo",
	}
}

// Creates a provider for any arbitrary OAuth2 service
func NewCustomProvider(config *oauth2.Config, userInfoURL string) *Provider {
	return &Provider{
		Config:   config,
		UserInfo: userInfoURL,
	}
}
