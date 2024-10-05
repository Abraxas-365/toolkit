package infrastructure

import (
	"context"
	"database/sql"

	"github.com/Abraxas-365/toolkit/pkg/errors"
	"github.com/Abraxas-365/toolkit/pkg/lucia"
	"github.com/jmoiron/sqlx"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetSession(ctx context.Context, sessionID string) (*lucia.UserSession, error) {
	query := `SELECT id, expires_at, user_id FROM user_session WHERE id = $1`
	var session lucia.UserSession

	err := r.db.GetContext(ctx, &session, query, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewLuciaError("UserSessionNotFound", "User session not found")
		}
		return nil, errors.NewLuciaError("DatabaseQueryError", err.Error())
	}

	if session.IsExpired() {
		return nil, errors.NewLuciaError("SessionExpired", "User session has expired")
	}

	return &session, nil
}
