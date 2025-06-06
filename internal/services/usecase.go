package services

import (
	"context"
	"errors"
	"github.com/MukizuL/diploma-1/internal/dto"
	"github.com/MukizuL/diploma-1/internal/errs"
	"github.com/MukizuL/diploma-1/internal/helpers"
	"github.com/greatcloak/decimal"
	"strconv"
)

func (s *Services) CreateOrder(ctx context.Context, userID string, orderID int64) error {
	if !helpers.ValidLuhn(orderID) {
		return errs.ErrWrongOrderFormat
	}

	user, err := s.storage.GetUserByOrderID(ctx, orderID)
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

func (s *Services) CreateOrderWithWithdrawal(ctx context.Context, userID string, orderID int64, sum decimal.Decimal) error {
	if !helpers.ValidLuhn(orderID) {
		return errs.ErrWrongOrderFormat
	}

	user, err := s.storage.GetUserByOrderID(ctx, orderID)
	if err != nil && !errors.Is(err, errs.ErrOrderNotFound) {
		return err
	}

	balance, _, err := s.storage.GetBalance(ctx, userID)
	if err != nil {
		return err
	}

	if sum.Cmp(balance) == 1 {
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

func (s *Services) GetOrders(ctx context.Context, userID string) ([]dto.OrderResponse, error) {
	orders, err := s.storage.GetOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []dto.OrderResponse

	for _, v := range orders {
		order := dto.OrderResponse{
			OrderID:   strconv.FormatInt(v.ID, 10),
			Status:    v.Status.String(),
			Accrual:   decimal.NewFromInt(v.Accrual).Div(decimal.NewFromInt(100)).InexactFloat64(),
			CreatedAt: v.CreatedAt,
		}

		result = append(result, order)
	}

	return result, nil
}

func (s *Services) GetWithdrawals(ctx context.Context, userID string) ([]dto.WithdrawalResponse, error) {
	orders, err := s.storage.GetWithdrawalsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []dto.WithdrawalResponse

	for _, v := range orders {
		order := dto.WithdrawalResponse{
			OrderID:   strconv.FormatInt(v.OrderID, 10),
			Sum:       decimal.NewFromInt(v.Sum).Div(decimal.NewFromInt(100)).InexactFloat64(),
			CreatedAt: v.CreatedAt,
		}

		result = append(result, order)
	}

	return result, nil
}

func (s *Services) GetBalance(ctx context.Context, userID string) (*dto.BalanceResponse, error) {
	balance, withdrawn, err := s.storage.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.BalanceResponse{
		Balance:   balance.InexactFloat64(),
		Withdrawn: withdrawn.InexactFloat64(),
	}, nil
}
