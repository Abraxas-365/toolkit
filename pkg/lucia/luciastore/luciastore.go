package luciastore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Abraxas-365/toolkit/pkg/errors"
	"github.com/Abraxas-365/toolkit/pkg/lucia"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sqlx.DB
}

// NewStoreFromConnection creates a new PostgresStore from an existing sqlx.DB connection
func NewStoreFromConnection(db *sqlx.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// NewStoreFromConnectionString creates a new PostgresStore from a connection string
func NewStoreFromConnectionString(connectionString string) (*PostgresStore, error) {
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return nil, errors.ErrDatabase(fmt.Sprintf("failed to connect to database: %v", err))
	}
	return &PostgresStore{db: db}, nil
}

// NewStoreFromConnectionStringAndDB creates a new PostgresStore from a connection string and database name
func NewStoreFromConnectionStringAndDB(connectionString, dbName string) (*PostgresStore, error) {
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return nil, errors.ErrDatabase(fmt.Sprintf("failed to connect to database: %v", err))
	}

	// Switch to the specified database
	_, err = db.Exec(fmt.Sprintf("USE %s", dbName))
	if err != nil {
		db.Close()
		return nil, errors.ErrDatabase(fmt.Sprintf("failed to switch to database %s: %v", dbName, err))
	}

	return &PostgresStore{db: db}, nil
}

// SessionStore implementation

func (s *PostgresStore) CreateSession(ctx context.Context, session *lucia.Session) error {
	query := `INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)`
	_, err := s.db.ExecContext(ctx, query, session.ID, session.UserID, time.Unix(session.ExpiresAt, 0))
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return errors.ErrConflict("Session already exists")
			case "foreign_key_violation":
				return errors.ErrBadRequest("Invalid user ID")
			}
		}
		return errors.ErrDatabase(fmt.Sprintf("Failed to create session: %v", err))
	}
	return nil
}

func (s *PostgresStore) GetSession(ctx context.Context, sessionID string) (*lucia.Session, error) {
	query := `SELECT id, user_id, EXTRACT(EPOCH FROM expires_at) as expires_at FROM sessions WHERE id = $1`
	var session lucia.Session
	err := s.db.GetContext(ctx, &session, query, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound("Session not found")
		}
		return nil, errors.ErrDatabase(fmt.Sprintf("Failed to get session: %v", err))
	}

	if time.Unix(session.ExpiresAt, 0).Before(time.Now()) {
		return nil, errors.ErrUnauthorized("Session expired")
	}

	return &session, nil
}

func (s *PostgresStore) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return errors.ErrDatabase(fmt.Sprintf("Failed to delete session: %v", err))
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.ErrDatabase(fmt.Sprintf("Failed to get rows affected: %v", err))
	}
	if rowsAffected == 0 {
		return errors.ErrNotFound("Session not found")
	}
	return nil
}

// Close closes the database connection
func (s *PostgresStore) Close() error {
	return s.db.Close()
}
