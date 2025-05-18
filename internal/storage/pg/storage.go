package pg

import (
	"context"
	"errors"
	"github.com/MukizuL/diploma-1/internal/errs"
	"github.com/MukizuL/diploma-1/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
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
				return "", errs.ErrConflictLogin
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", errs.ErrUserNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.logger.Error("Failed to find user",
				zap.String("method", "GetUserByLogin"),
				zap.String("login", login),
				zap.Error(pgErr))

			return nil, "", errs.ErrInternalServerError
		}
	}

	return &user, passwordHash, nil
}

func (s *Storage) GetOrderByID(ctx context.Context, orderID int64) (string, error) {
	var userID string
	err := s.conn.QueryRow(ctx, `SELECT user_id FROM orders WHERE order_id = $1`, orderID).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.ErrOrderNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.logger.Error("Failed to find order",
				zap.String("method", "GetOrderByID"),
				zap.Int64("orderID", orderID),
				zap.Error(pgErr))

			return "", errs.ErrInternalServerError
		}
	}

	return userID, nil
}

func (s *Storage) CreateNewOrder(ctx context.Context, userID string, orderID int64) error {
	_, err := s.conn.Exec(ctx, `INSERT INTO orders (user_id, order_id) VALUES ($1, $2)`, userID, orderID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.logger.Error("Failed to create order",
				zap.String("method", "CreateNewOrder"),
				zap.String("userID", userID),
				zap.Int64("orderID", orderID),
				zap.Error(pgErr))

			return errs.ErrInternalServerError
		}
	}

	return nil
}
