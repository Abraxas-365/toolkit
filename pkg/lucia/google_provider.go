package lucia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Abraxas-365/toolkit/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleProvider struct {
	config *oauth2.Config
}

func NewGoogleProvider(clientID, clientSecret, redirectURI string) *GoogleProvider {
	return &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURI,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func (p *GoogleProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *GoogleProvider) ExchangeCode(ctx context.Context, code string) (*OAuthToken, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, errors.ErrUnauthorized(fmt.Sprintf("Failed to exchange code: %v", err))
	}

	return &OAuthToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    token.Expiry.Unix(),
	}, nil
}

func (p *GoogleProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*UserInfo, error) {
	client := p.config.Client(ctx, &oauth2.Token{
		AccessToken: token.AccessToken,
	})

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, errors.ErrUnauthorized(fmt.Sprintf("Failed to get user info: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.ErrUnauthorized(fmt.Sprintf("Failed to get user info: status code %d", resp.StatusCode))
	}

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, errors.ErrUnexpected(fmt.Sprintf("Failed to decode user info: %v", err))
	}

	return &UserInfo{
		ID:       googleUser.ID,
		Email:    googleUser.Email,
		Name:     googleUser.Name,
		Provider: "google",
	}, nil
}
