package lucia

import "context"

type OAuthProvider interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*OAuthToken, error)
	GetUserInfo(ctx context.Context, token *OAuthToken) (*UserInfo, error)
}

type OAuthToken struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type UserInfo struct {
	ID             string
	Email          string
	Name           string
	Provider       string
	ProfilePicture *string
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
	UserID    string
	ExpiresAt int64
}
