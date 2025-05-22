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
	"time"
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

// GetUserByLogin Fetches user from database and stores all non-sensitive data in User struct. Returns User and an error.
func (s *Storage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	err := s.conn.QueryRow(ctx, `SELECT id, login, created_at, passwordHash FROM users WHERE login = $1`, login).
		Scan(&user.ID, &user.Login, &user.CreatedAt, &user.PasswordHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.logger.Error("Failed to find user",
				zap.String("method", "GetUserByLogin"),
				zap.String("login", login),
				zap.Error(pgErr))

			return nil, errs.ErrInternalServerError
		}
	}

	return &user, nil
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

func (s *Storage) GetOrdersByUser(ctx context.Context, userID string) ([]models.Order, error) {
	var result []models.Order
	rows, err := s.conn.Query(ctx, `SELECT id, user_id, order_id, status, accrual, created_at FROM orders WHERE user_id = $1`, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.logger.Error("Failed to get orders",
				zap.String("method", "GetOrdersByUser"),
				zap.String("userID", userID),
				zap.Error(pgErr))

			return nil, errs.ErrInternalServerError
		}
	}
	defer rows.Close()

	var ID, UserID, Status string
	var OrderID int64
	var Accrual float64
	var CreatedAt time.Time

	for rows.Next() {
		err = rows.Scan(&ID, &UserID, &OrderID, &Status, &Accrual, &CreatedAt)
		if err != nil {
			s.logger.Error("Error in row",
				zap.String("method", "GetOrdersByUser"),
				zap.String("userID", userID),
				zap.Error(err))
			continue
		}

		MStatus, err := models.NewStatus(Status)
		if err != nil {
			s.logger.Error("Failed to convert to models.Status", zap.String("got_status", Status))
			continue
		}

		data := models.Order{
			ID:        ID,
			UserID:    UserID,
			OrderID:   OrderID,
			Status:    MStatus,
			Accrual:   Accrual,
			CreatedAt: CreatedAt,
		}

		result = append(result, data)
	}

	if rows.Err() != nil {
		s.logger.Error("Error in rows",
			zap.String("method", "GetOrdersByUser"),
			zap.String("userID", userID),
			zap.Error(err))
		return nil, errs.ErrInternalServerError
	}

	if len(result) == 0 {
		return nil, errs.ErrOrderNotFound
	}

	return result, nil
}

// GetBalance Returns balance and withdrawn amount.
func (s *Storage) GetBalance(ctx context.Context, userID string) (float64, float64, error) {
	var balance, withdrawn float64
	err := s.conn.QueryRow(ctx, `SELECT SUM(amount) FROM withdrawals WHERE user_id = $1`, userID).Scan(&withdrawn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, errs.ErrUserNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.logger.Error("Failed to get withdrawn",
				zap.String("method", "GetBalance"),
				zap.String("userID", userID),
				zap.Error(pgErr))

			return 0, 0, errs.ErrInternalServerError
		}
	}

	err = s.conn.QueryRow(ctx, `SELECT SUM(accrual) FROM orders WHERE user_id = $1`, userID).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		balance = 0
		err = nil
	}

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.logger.Error("Failed to get balance",
				zap.String("method", "GetBalance"),
				zap.String("userID", userID),
				zap.Error(pgErr))

			return 0, 0, errs.ErrInternalServerError
		}
	}

	return balance - withdrawn, withdrawn, nil
}

func (s *Storage) CreateNewOrderWithWithdrawal(ctx context.Context, userID string, orderID int64, sum float64) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		s.logger.Error("Failed to begin transaction",
			zap.String("method", "CreateNewOrderWithWithdrawal"),
			zap.String("userID", userID),
			zap.Int64("orderID", orderID),
			zap.Float64("sum", sum),
			zap.Error(err))

		return errs.ErrInternalServerError
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `INSERT INTO orders (user_id, order_id) VALUES ($1, $2)`, userID, orderID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.logger.Error("Failed to create order",
				zap.String("method", "CreateNewOrderWithWithdrawal"),
				zap.String("userID", userID),
				zap.Int64("orderID", orderID),
				zap.Float64("sum", sum),
				zap.Error(pgErr))

			return errs.ErrInternalServerError
		}
	}

	_, err = tx.Exec(ctx, `INSERT INTO withdrawals (user_id, order_id, amount) VALUES ($1, $2, $3)`, userID, orderID, sum)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.logger.Error("Failed to insert into withdrawals",
				zap.String("method", "CreateNewOrderWithWithdrawal"),
				zap.String("userID", userID),
				zap.Int64("orderID", orderID),
				zap.Float64("sum", sum),
				zap.Error(pgErr))

			return errs.ErrInternalServerError
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.logger.Error("Failed to commit transaction",
			zap.String("method", "CreateNewOrderWithWithdrawal"),
			zap.String("userID", userID),
			zap.Int64("orderID", orderID),
			zap.Float64("sum", sum),
			zap.Error(err))

		return errs.ErrInternalServerError
	}

	return nil
}

func (s *Storage) GetWithdrawalsByUser(ctx context.Context, userID string) ([]models.Withdrawal, error) {
	var result []models.Withdrawal
	rows, err := s.conn.Query(ctx, `SELECT id, user_id, order_id, amount, created_at FROM withdrawals WHERE user_id = $1`, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.logger.Error("Failed to get withdrawals",
				zap.String("method", "GetWithdrawalsByUser"),
				zap.String("userID", userID),
				zap.Error(pgErr))

			return nil, errs.ErrInternalServerError
		}
	}
	defer rows.Close()

	var ID, UserID string
	var OrderID int64
	var Sum float64
	var CreatedAt time.Time

	for rows.Next() {
		err = rows.Scan(&ID, &UserID, &OrderID, &Sum, &CreatedAt)
		if err != nil {
			s.logger.Error("Error in row",
				zap.String("method", "GetWithdrawalsByUser"),
				zap.String("userID", userID),
				zap.Error(err))
			continue
		}

		data := models.Withdrawal{
			ID:        ID,
			UserID:    UserID,
			OrderID:   OrderID,
			Sum:       Sum,
			CreatedAt: CreatedAt,
		}

		result = append(result, data)
	}

	if rows.Err() != nil {
		s.logger.Error("Error in rows",
			zap.String("method", "GetWithdrawalsByUser"),
			zap.String("userID", userID),
			zap.Error(err))

		return nil, errs.ErrInternalServerError
	}

	if len(result) == 0 {
		return nil, errs.ErrWithdrawalNotFound
	}

	return result, nil
}
