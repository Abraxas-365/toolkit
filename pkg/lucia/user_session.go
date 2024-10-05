package lucia

import (
	"time"

	"github.com/google/uuid"
)

type UserSession struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    string    `json:"user_id"`
}

func (us *UserSession) IsExpired() bool {
	return us.ExpiresAt.Before(time.Now())
}

func NewUserSession(userID string, expiresIn time.Duration) *UserSession {
	return &UserSession{
		ID:        uuid.New().String(),
		ExpiresAt: time.Now().Add(expiresIn),
		UserID:    userID,
	}
}
