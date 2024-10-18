package lucia

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"
)

type AuthService struct {
	providers    map[string]OAuthProvider
	userStore    UserStore
	sessionStore SessionStore
}

func NewAuthService(userStore UserStore, sessionStore SessionStore) *AuthService {
	return &AuthService{
		providers:    make(map[string]OAuthProvider),
		userStore:    userStore,
		sessionStore: sessionStore,
	}
}

func (s *AuthService) RegisterProvider(name string, provider OAuthProvider) {
	s.providers[name] = provider
}

func (s *AuthService) GetAuthURL(provider string) (string, string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return "", "", errors.New("unknown provider")
	}
	state := generateState()
	url := p.GetAuthURL(state)
	return url, state, nil
}

func (s *AuthService) HandleCallback(ctx context.Context, provider, code string) (*Session, error) {
	p, ok := s.providers[provider]
	if !ok {
		return nil, errors.New("unknown provider")
	}

	token, err := p.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	userInfo, err := p.GetUserInfo(ctx, token)
	if err != nil {
		return nil, err
	}

	user, err := s.userStore.GetUserByProviderID(ctx, provider, userInfo.ID)
	if err != nil {
		// If user doesn't exist, create a new one
		user = &User{
			ID:         generateID(),
			Email:      userInfo.Email,
			Name:       userInfo.Name,
			ProviderID: userInfo.ID,
			Provider:   provider,
		}
		if err := s.userStore.CreateUser(ctx, user); err != nil {
			return nil, err
		}
	}

	session := &Session{
		ID:        generateID(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}
	if err := s.sessionStore.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *AuthService) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	return s.sessionStore.GetSession(ctx, sessionID)
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	return s.sessionStore.DeleteSession(ctx, sessionID)
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (s *AuthService) CreateSession(ctx context.Context, userID string) (*Session, error) {
	session := &Session{
		ID:        generateID(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}
	if err := s.sessionStore.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *AuthService) DeleteSession(ctx context.Context, sessionID string) error {
	err := s.sessionStore.DeleteSession(ctx, sessionID)
	if err != nil {
		return err
	}
	return nil
}
