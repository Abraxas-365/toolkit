package lucia

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/Abraxas-365/toolkit/pkg/errors"
)

// AuthUser is an interface that any user type must implement
type AuthUser interface {
	GetID() string
}

type AuthService[U AuthUser] struct {
	providers    map[string]OAuthProvider
	userStore    AuthUserStore[U]
	sessionStore SessionStore
}

func NewAuthService[U AuthUser](userStore AuthUserStore[U], sessionStore SessionStore) *AuthService[U] {
	return &AuthService[U]{
		providers:    make(map[string]OAuthProvider),
		userStore:    userStore,
		sessionStore: sessionStore,
	}
}

func (s *AuthService[U]) RegisterProvider(name string, provider OAuthProvider) {
	s.providers[name] = provider
}

func (s *AuthService[U]) GetAuthURL(provider string) (string, string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return "", "", errors.NewLuciaError("UnknownProvider", "Unknown OAuth provider")
	}
	state := generateState()
	url := p.GetAuthURL(state)
	return url, state, nil
}

func (s *AuthService[U]) HandleCallback(ctx context.Context, provider, code string) (*Session, error) {
	p, ok := s.providers[provider]
	if !ok {
		return nil, errors.NewLuciaError("UnknownProvider", "Unknown OAuth provider")
	}

	token, err := p.ExchangeCode(ctx, code)
	if err != nil {
		return nil, errors.NewLuciaError("TokenExchangeError", "Failed to exchange code for token")
	}

	userInfo, err := p.GetUserInfo(ctx, token)
	if err != nil {
		return nil, errors.NewLuciaError("UserInfoError", "Failed to get user info")
	}
	userInfo.Token = token

	user, err := s.userStore.GetUserByProviderID(ctx, provider, userInfo.ID)
	if err != nil {
		if errors.IsNotFound(err) {
			// If user doesn't exist, create a new one
			user, err = s.userStore.CreateUser(ctx, userInfo)
			if err != nil {
				return nil, errors.NewLuciaError("UserCreationFailed", "Failed to create user")
			}
		} else {
			return nil, errors.NewLuciaError("DatabaseError", "Failed to fetch user")
		}
	}

	session := &Session{
		ID:        GenerateID(),
		UserID:    user.GetID(),
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}
	if err := s.sessionStore.CreateSession(ctx, session); err != nil {
		return nil, errors.NewLuciaError("SessionCreationFailed", "Failed to create session")
	}

	return session, nil
}

func (s *AuthService[U]) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	session, err := s.sessionStore.GetSession(ctx, sessionID)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, errors.NewLuciaError("UserSessionNotFound", "Session not found")
		}
		return nil, errors.NewLuciaError("DatabaseError", "Failed to fetch session")
	}
	return session, nil
}

func (s *AuthService[U]) Logout(ctx context.Context, sessionID string) error {
	err := s.sessionStore.DeleteSession(ctx, sessionID)
	if err != nil {
		return errors.NewLuciaError("SessionDeletionFailed", "Failed to delete session")
	}
	return nil
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (s *AuthService[U]) CreateSession(ctx context.Context, user U) (*Session, error) {
	session := &Session{
		ID:        GenerateID(),
		UserID:    user.GetID(),
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}
	if err := s.sessionStore.CreateSession(ctx, session); err != nil {
		return nil, errors.NewLuciaError("SessionCreationFailed", "Failed to create session")
	}
	return session, nil
}

func (s *AuthService[U]) DeleteSession(ctx context.Context, sessionID string) error {
	err := s.sessionStore.DeleteSession(ctx, sessionID)
	if err != nil {
		return errors.NewLuciaError("SessionDeletionFailed", "Failed to delete session")
	}
	return nil
}
