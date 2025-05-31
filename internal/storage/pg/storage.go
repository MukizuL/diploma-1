package pg

import (
	"context"
	"errors"
	"github.com/MukizuL/diploma-1/internal/errs"
	"github.com/MukizuL/diploma-1/internal/models"
	"github.com/google/uuid"
	"github.com/greatcloak/decimal"
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
			default:
				return "", errs.ErrInternalServerError
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

		s.logger.Error("Failed to find user",
			zap.String("method", "GetUserByLogin"),
			zap.String("login", login),
			zap.Error(err))

		return nil, errs.ErrInternalServerError
	}

	return &user, nil
}

func (s *Storage) GetUserByOrderID(ctx context.Context, orderID int64) (string, error) {
	var userID string
	err := s.conn.QueryRow(ctx, `SELECT user_id FROM orders WHERE id = $1`, orderID).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.ErrOrderNotFound
		}

		s.logger.Error("Failed to find order",
			zap.String("method", "GetUserByOrderID"),
			zap.Int64("orderID", orderID),
			zap.Error(err))

		return "", errs.ErrInternalServerError
	}

	return userID, nil
}

func (s *Storage) CreateNewOrder(ctx context.Context, userID string, orderID int64) error {
	_, err := s.conn.Exec(ctx, `INSERT INTO orders (user_id, id) VALUES ($1, $2)`, userID, orderID)
	if err != nil {
		s.logger.Error("Failed to create order",
			zap.String("method", "CreateNewOrder"),
			zap.String("userID", userID),
			zap.Int64("orderID", orderID),
			zap.Error(err))

		return errs.ErrInternalServerError
	}

	return nil
}

func (s *Storage) GetOrdersByUser(ctx context.Context, userID string) ([]models.Order, error) {
	var result []models.Order
	rows, err := s.conn.Query(ctx, `SELECT id, user_id, status, accrual, created_at FROM orders WHERE user_id = $1`, userID)
	if err != nil {
		s.logger.Error("Failed to get orders",
			zap.String("method", "GetOrdersByUser"),
			zap.String("userID", userID),
			zap.Error(err))

		return nil, errs.ErrInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		err = rows.Scan(&order.ID, &order.UserID, &order.Status, &order.Accrual, &order.CreatedAt)
		if err != nil {
			s.logger.Error("Error in row",
				zap.String("method", "GetOrdersByUser"),
				zap.String("userID", userID),
				zap.Error(err))
			continue
		}

		result = append(result, order)
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
func (s *Storage) GetBalance(ctx context.Context, userID string) (decimal.Decimal, decimal.Decimal, error) {
	var balance, withdrawn decimal.Decimal
	err := s.conn.QueryRow(ctx, `SELECT SUM(amount) FROM withdrawals WHERE user_id = $1`, userID).Scan(&withdrawn)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		s.logger.Error("Failed to get withdrawn",
			zap.String("method", "GetBalance"),
			zap.String("userID", userID),
			zap.Error(err))

		return decimal.NewFromInt(0), decimal.NewFromInt(0), errs.ErrInternalServerError
	}

	err = s.conn.QueryRow(ctx, `SELECT SUM(accrual) FROM orders WHERE user_id = $1`, userID).Scan(&balance)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		s.logger.Error("Failed to get balance",
			zap.String("method", "GetBalance"),
			zap.String("userID", userID),
			zap.Error(err))

		return decimal.NewFromInt(0), decimal.NewFromInt(0), errs.ErrInternalServerError
	}

	return balance.Sub(withdrawn).Div(decimal.NewFromInt(100)), withdrawn.Div(decimal.NewFromInt(100)), nil
}

func (s *Storage) CreateNewOrderWithWithdrawal(ctx context.Context, userID string, orderID int64, sum decimal.Decimal) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		s.logger.Error("Failed to begin transaction",
			zap.String("method", "CreateNewOrderWithWithdrawal"),
			zap.String("userID", userID),
			zap.Int64("orderID", orderID),
			zap.String("sum", sum.String()),
			zap.Error(err))

		return errs.ErrInternalServerError
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `INSERT INTO orders (user_id, id) VALUES ($1, $2)`, userID, orderID)
	if err != nil {
		s.logger.Error("Failed to create order",
			zap.String("method", "CreateNewOrderWithWithdrawal"),
			zap.String("userID", userID),
			zap.Int64("orderID", orderID),
			zap.String("sum", sum.String()),
			zap.Error(err))

		return errs.ErrInternalServerError
	}

	_, err = tx.Exec(ctx, `INSERT INTO withdrawals (user_id, order_id, amount) VALUES ($1, $2, $3)`,
		userID,
		orderID,
		sum.Mul(decimal.NewFromInt(100)).IntPart())
	if err != nil {
		s.logger.Error("Failed to insert into withdrawals",
			zap.String("method", "CreateNewOrderWithWithdrawal"),
			zap.String("userID", userID),
			zap.Int64("orderID", orderID),
			zap.String("sum", sum.String()),
			zap.Error(err))

		return errs.ErrInternalServerError
	}

	err = tx.Commit(ctx)
	if err != nil {
		s.logger.Error("Failed to commit transaction",
			zap.String("method", "CreateNewOrderWithWithdrawal"),
			zap.String("userID", userID),
			zap.Int64("orderID", orderID),
			zap.String("sum", sum.String()),
			zap.Error(err))

		return errs.ErrInternalServerError
	}

	return nil
}

func (s *Storage) GetWithdrawalsByUser(ctx context.Context, userID string) ([]models.Withdrawal, error) {
	var result []models.Withdrawal
	rows, err := s.conn.Query(ctx, `SELECT id, user_id, order_id, amount, created_at FROM withdrawals WHERE user_id = $1`, userID)
	if err != nil {
		s.logger.Error("Failed to get withdrawals",
			zap.String("method", "GetWithdrawalsByUser"),
			zap.String("userID", userID),
			zap.Error(err))

		return nil, errs.ErrInternalServerError
	}
	defer rows.Close()

	for rows.Next() {
		var withdrawal models.Withdrawal
		err = rows.Scan(&withdrawal.ID, &withdrawal.UserID, &withdrawal.OrderID, &withdrawal.Sum, &withdrawal.CreatedAt)
		if err != nil {
			s.logger.Error("Error in row",
				zap.String("method", "GetWithdrawalsByUser"),
				zap.String("userID", userID),
				zap.Error(err))
			continue
		}

		result = append(result, withdrawal)
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

func (s *Storage) UpdateOrder(ctx context.Context, orderID int64, status string) error {
	_, err := s.conn.Exec(ctx, `UPDATE orders SET status = $1 WHERE id = $2`, status, orderID)
	if err != nil {
		s.logger.Error("Failed to update order",
			zap.String("method", "UpdateOrder"),
			zap.Int64("orderID", orderID),
			zap.String("status", status),
			zap.Error(err))

		return errs.ErrInternalServerError
	}

	return nil
}

func (s *Storage) UpdateOrderWithAccrual(ctx context.Context, orderID int64, status string, accrual decimal.Decimal) error {
	_, err := s.conn.Exec(ctx, `UPDATE orders SET status = $1, accrual = $2 WHERE id = $3`, status, accrual, orderID)
	if err != nil {
		s.logger.Error("Failed to update order",
			zap.String("method", "UpdateOrder"),
			zap.Int64("orderID", orderID),
			zap.String("status", status),
			zap.Error(err))

		return errs.ErrInternalServerError
	}

	return nil
}
