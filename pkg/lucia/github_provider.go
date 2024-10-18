package lucia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

	return &token, nil
}

func (p *GitHubProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*UserInfo, error) {
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
	}

	if githubUser.AvatarURL != "" {
		userInfo.ProfilePicture = &githubUser.AvatarURL
	}

	return userInfo, nil
}
