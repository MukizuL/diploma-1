package pg

import (
	"context"
	"errors"
	"github.com/MukizuL/diploma-1/internal/errs"
	"github.com/MukizuL/diploma-1/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

// CreateNewUser Creates a new user with given login and password. Returns userID and an error.
func (s *Storage) CreateNewUser(ctx context.Context, login, passwordHash string) (string, error) {
	userID := uuid.New()

	_, err := s.conn.Exec(ctx, `INSERT INTO users (id, login, passwordHash) VALUES ($1, $2, $3)`, userID, login, passwordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return "", errs.ErrDuplicateLogin
			}
		}

		s.logger.Error("Failed to create user",
			zap.String("method", "CreateNewUser"),
			zap.String("login", login),
			zap.Error(pgErr))

		return "", errs.ErrInternalServerError
	}

	return userID.String(), nil
}

// GetUserByLogin Fetches user from database and stores all non-sensitive data in User struct. Returns User, password and an error.
func (s *Storage) GetUserByLogin(ctx context.Context, login string) (*models.User, string, error) {
	var user models.User
	var passwordHash string
	err := s.conn.QueryRow(ctx, `SELECT id, login, created_at, passwordHash FROM users WHERE login = $1`, login).
		Scan(&user.Id, &user.Login, &user.CreatedAt, &passwordHash)

	if err != nil {
		var pgErr *pgconn.PgError

		s.logger.Error("Failed to find user",
			zap.String("method", "GetUserByLogin"),
			zap.String("login", login),
			zap.Error(pgErr))

		return nil, "", errs.ErrInternalServerError
	}

	return &user, passwordHash, nil
}
