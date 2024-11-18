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

func NewGoogleProvider(clientID, clientSecret, redirectURI string, scopes []string) *GoogleProvider {
	return &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURI,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
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
	// First check if token needs refresh
	if token.NeedsRefresh() {
		newToken, err := p.RefreshToken(ctx, token.RefreshToken)
		if err != nil {
			return nil, err
		}
		// Update token with new values
		token.AccessToken = newToken.AccessToken
		token.ExpiresIn = newToken.ExpiresIn
		if newToken.RefreshToken != "" {
			token.RefreshToken = newToken.RefreshToken
		}
	}

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
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, errors.ErrUnexpected(fmt.Sprintf("Failed to decode user info: %v", err))
	}

	userInfo := &UserInfo{
		ID:       googleUser.ID,
		Email:    googleUser.Email,
		Name:     googleUser.Name,
		Provider: "google",
		Token:    token, // Include the potentially refreshed token
	}

	if googleUser.Picture != "" {
		userInfo.ProfilePicture = &googleUser.Picture
	}

	return userInfo, nil
}

func (p *GoogleProvider) RefreshToken(ctx context.Context, refreshToken string) (*OAuthToken, error) {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := p.config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, errors.ErrUnauthorized(fmt.Sprintf("Failed to refresh token: %v", err))
	}

	return &OAuthToken{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		ExpiresIn:    newToken.Expiry.Unix(),
	}, nil
}
