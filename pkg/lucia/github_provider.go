package lucia

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
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
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var token OAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

func (p *GitHubProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var githubUser struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:       string(githubUser.ID),
		Email:    githubUser.Email,
		Name:     githubUser.Name,
		Provider: "github",
	}, nil
}
