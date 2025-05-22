package services

import (
	"context"
	"errors"
	"github.com/MukizuL/diploma-1/internal/dto"
	"github.com/MukizuL/diploma-1/internal/errs"
	"github.com/MukizuL/diploma-1/internal/helpers"
	"go.uber.org/zap"
	"strconv"
)

func (s *Services) PostOrder(ctx context.Context, userID string, orderID int64) error {
	if !helpers.ValidLuhn(orderID) {
		return errs.ErrWrongOrderFormat
	}

	user, err := s.storage.GetOrderByID(ctx, orderID)
	if err != nil && !errors.Is(err, errs.ErrOrderNotFound) {
		return err
	}

	if user != "" {
		if userID != user {
			return errs.ErrConflictOrder
		} else {
			return errs.ErrDuplicateOrder
		}
	}

	err = s.storage.CreateNewOrder(ctx, userID, orderID)
	if err != nil {
		return err
	}

	err = s.worker.Push(orderID)
	if err != nil {
		return errs.ErrInternalServerError
	}

	return nil
}

func (s *Services) PostOrderWithWithdrawal(ctx context.Context, userID string, orderID int64, sum float64) error {
	if !helpers.ValidLuhn(orderID) {
		return errs.ErrWrongOrderFormat
	}

	user, err := s.storage.GetOrderByID(ctx, orderID)
	if err != nil && !errors.Is(err, errs.ErrOrderNotFound) {
		return err
	}

	balance, _, err := s.storage.GetBalance(ctx, userID)
	if err != nil {
		return err
	}

	if balance < sum {
		return errs.ErrInsufficientBalance
	}

	if user != "" {
		if userID != user {
			return errs.ErrConflictOrder
		} else {
			return errs.ErrDuplicateOrder
		}
	}

	err = s.storage.CreateNewOrderWithWithdrawal(ctx, userID, orderID, sum)
	if err != nil {
		return err
	}

	err = s.worker.Push(orderID)
	if err != nil {
		return errs.ErrInternalServerError
	}

	return nil
}

func (s *Services) GetOrders(ctx context.Context, userID string) ([]dto.OrderOut, error) {
	orders, err := s.storage.GetOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []dto.OrderOut

	for _, v := range orders {
		order := dto.OrderOut{
			OrderID:   strconv.FormatInt(v.OrderID, 10),
			Status:    v.Status.String(),
			Accrual:   v.Accrual,
			CreatedAt: v.CreatedAt,
		}

		result = append(result, order)
	}

	return result, nil
}

func (s *Services) GetWithdrawals(ctx context.Context, userID string) ([]dto.WithdrawalOut, error) {
	orders, err := s.storage.GetWithdrawalsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []dto.WithdrawalOut

	for _, v := range orders {
		order := dto.WithdrawalOut{
			OrderID:   strconv.FormatInt(v.OrderID, 10),
			Sum:       v.Sum,
			CreatedAt: v.CreatedAt,
		}

		result = append(result, order)
	}

	return result, nil
}

func (s *Services) GetBalance(ctx context.Context, userID string) (*dto.BalanceOut, error) {
	balance, withdrawn, err := s.storage.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("Balance in service", zap.Float64("balance", balance), zap.Float64("withdrawn", withdrawn))

	return &dto.BalanceOut{
		Balance:   balance,
		Withdrawn: withdrawn,
	}, nil
}
