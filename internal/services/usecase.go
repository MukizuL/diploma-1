package services

import (
	"context"
	"errors"
	"github.com/MukizuL/diploma-1/internal/dto"
	"github.com/MukizuL/diploma-1/internal/errs"
	"github.com/MukizuL/diploma-1/internal/helpers"
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

	return nil
}

func (s *Services) GetOrders(ctx context.Context, userID string) ([]dto.Order, error) {
	orders, err := s.storage.GetOrdersByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []dto.Order

	for _, v := range orders {
		order := dto.Order{
			OrderID:   strconv.FormatInt(v.OrderID, 10),
			Status:    v.Status.String(),
			Accrual:   v.Accrual,
			CreatedAt: v.CreatedAt,
		}

		result = append(result, order)
	}

	return result, nil
}

func (s *Services) GetBalance(ctx context.Context, userID string) (*dto.Balance, error) {
	return nil, nil
}
