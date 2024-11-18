package lucia

import (
	"context"
	"time"

	"github.com/Abraxas-365/toolkit/pkg/errors"
)

type OAuthProvider interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*OAuthToken, error)
	GetUserInfo(ctx context.Context, token *OAuthToken) (*UserInfo, error)
	RefreshToken(ctx context.Context, refreshToken string) (*OAuthToken, error)
}

type OAuthToken struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

func (t *OAuthToken) NeedsRefresh() bool {
	if t.ExpiresIn == 0 {
		return false
	}

	bufferTime := int64(300)
	return time.Now().Unix()+bufferTime >= t.ExpiresIn
}

func (t *OAuthToken) RefreshIfNeeded(ctx context.Context, provider OAuthProvider) error {
	if t.NeedsRefresh() && t.RefreshToken != "" {
		newToken, err := provider.RefreshToken(ctx, t.RefreshToken)
		if err != nil {
			return err
		}
		// Update the current token with new values
		t.AccessToken = newToken.AccessToken
		t.ExpiresIn = newToken.ExpiresIn
		// Only update refresh token if a new one was provided
		if newToken.RefreshToken != "" {
			t.RefreshToken = newToken.RefreshToken
		}
	}
	return nil
}

type UserInfo struct {
	ID             string
	Email          string
	Name           string
	Provider       string
	ProfilePicture *string
	Token          *OAuthToken
}

type AuthUserStore[U AuthUser] interface {
	GetUserByProviderID(ctx context.Context, provider, providerID string) (U, error)
	CreateUser(ctx context.Context, userInfo *UserInfo) (U, error)
}

type SessionStore interface {
	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

type Session struct {
	ID        string
	UserID    interface{}
	ExpiresAt int64
}

func (s *Session) IsExpired() bool {
	return s.ExpiresAt < time.Now().Unix()
}

func (s *Session) UserIDToString() (string, error) {
	id, ok := s.UserID.(string)
	if !ok {
		return "", errors.ErrParse("UserID is not a string")
	}
	return id, nil
}

func (s *Session) UserIDToInt() (int, error) {
	id, ok := s.UserID.(int)
	if !ok {
		return 0, errors.ErrParse("UserID is not an int")
	}
	return id, nil
}

//TODO: Add more like uuid,mongo primitive object id , etc
