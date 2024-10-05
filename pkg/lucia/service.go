package lucia

import (
	"context"

	"github.com/Abraxas-365/toolkit/pkg/errors"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ValidateSession(ctx context.Context, sessionID string) (*UserSession, error) {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.IsExpired() {
		return nil, errors.NewLuciaError("SessionExpired", "User session has expired")
	}

	return session, nil
}
