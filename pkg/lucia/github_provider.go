package lucia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Abraxas-365/toolkit/pkg/errors"
)

type GitHubProvider struct {
	clientID     string
	clientSecret string
	redirectURI  string
}

func NewGitHubProvider(clientID, clientSecret, redirectURI string) *GitHubProvider {
	return &GitHubProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}
}

func (p *GitHubProvider) GetAuthURL(state string) string {
	return "https://github.com/login/oauth/authorize?" + url.Values{
		"client_id":    {p.clientID},
		"redirect_uri": {p.redirectURI},
		"state":        {state},
		"scope":        {"user:email"},
	}.Encode()
}

func (p *GitHubProvider) ExchangeCode(ctx context.Context, code string) (*OAuthToken, error) {
	values := url.Values{
		"client_id":     {p.clientID},
		"client_secret": {p.clientSecret},
		"code":          {code},
		"redirect_uri":  {p.redirectURI},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, errors.ErrUnexpected(fmt.Sprintf("Failed to create request: %v", err))
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.ErrUnexpected(fmt.Sprintf("Failed to exchange code: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.ErrUnauthorized(fmt.Sprintf("Failed to exchange code: status code %d", resp.StatusCode))
	}

	var token OAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, errors.ErrUnexpected(fmt.Sprintf("Failed to decode token response: %v", err))
	}

	// GitHub tokens typically expire after 8 hours
	token.ExpiresIn = time.Now().Add(8 * time.Hour).Unix()

	return &token, nil
}

func (p *GitHubProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*UserInfo, error) {
	// Check if token needs refresh
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

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, errors.ErrUnexpected(fmt.Sprintf("Failed to create request: %v", err))
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.ErrUnexpected(fmt.Sprintf("Failed to get user info: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.ErrUnauthorized(fmt.Sprintf("Failed to get user info: status code %d", resp.StatusCode))
	}

	var githubUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, errors.ErrUnexpected(fmt.Sprintf("Failed to decode user info: %v", err))
	}

	userInfo := &UserInfo{
		ID:       fmt.Sprintf("%d", githubUser.ID),
		Email:    githubUser.Email,
		Name:     githubUser.Name,
		Provider: "github",
		Token:    token, // Include the potentially refreshed token
	}

	if githubUser.AvatarURL != "" {
		userInfo.ProfilePicture = &githubUser.AvatarURL
	}

	return userInfo, nil
}

func (p *GitHubProvider) RefreshToken(ctx context.Context, refreshToken string) (*OAuthToken, error) {
	values := url.Values{
		"client_id":     {p.clientID},
		"client_secret": {p.clientSecret},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, errors.ErrUnexpected(fmt.Sprintf("Failed to create refresh request: %v", err))
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.ErrUnexpected(fmt.Sprintf("Failed to refresh token: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.ErrUnauthorized(fmt.Sprintf("Failed to refresh token: status code %d", resp.StatusCode))
	}

	var token OAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, errors.ErrUnexpected(fmt.Sprintf("Failed to decode refresh token response: %v", err))
	}

	// If GitHub doesn't provide a new refresh token, use the existing one
	if token.RefreshToken == "" {
		token.RefreshToken = refreshToken
	}

	return &token, nil
}
