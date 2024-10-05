package lucia

import "context"

type Repository interface {
	GetSession(ctx context.Context, sessionID string) (*UserSession, error)
}
