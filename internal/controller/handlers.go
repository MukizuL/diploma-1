package controller

import (
	"errors"
	"github.com/MukizuL/diploma-1/internal/dto"
	"github.com/MukizuL/diploma-1/internal/errs"
	"github.com/MukizuL/diploma-1/internal/helpers"
	"github.com/gin-gonic/gin"
	"github.com/greatcloak/decimal"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const MaxOrderIDLength = 18

func (c *Controller) Register(ctx *gin.Context) {
	var data dto.AuthFormRequest
	err := ctx.BindJSON(&data)
	if err != nil {
		helpers.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	token, err := c.service.CreateUser(ctx.Request.Context(), data.Login, data.Password)
	if err != nil {
		if errors.Is(err, errs.ErrConflictLogin) {
			helpers.RespondError(ctx, http.StatusConflict, err.Error())
			return
		}

		c.logger.Error("Error in handler", zap.String("handler", "Register"), zap.Error(err))

		helpers.RespondInternalServerError(ctx, err.Error())
		return
	}

	ctx.SetCookie("Access-token", token, 3600, "/", c.domain, false, true)

	ctx.JSON(http.StatusOK, &gin.H{
		"Result": http.StatusText(http.StatusOK),
	})
}

func (c *Controller) Login(ctx *gin.Context) {
	var data dto.AuthFormRequest
	err := ctx.BindJSON(&data)
	if err != nil {
		helpers.RespondError(ctx, http.StatusBadRequest, err.Error())

		return
	}

	token, err := c.service.LoginUser(ctx.Request.Context(), data.Login, data.Password)
	if err != nil {
		if errors.Is(err, errs.ErrNotAuthorized) {
			helpers.RespondError(ctx, http.StatusUnauthorized, "Access token is invalid")
			return
		}

		if errors.Is(err, errs.ErrUserNotFound) {
			helpers.RespondError(ctx, http.StatusUnauthorized, "Login or password is incorrect")
			return
		}

		c.logger.Error("Error in handler", zap.String("handler", "Login"), zap.Error(err))

		helpers.RespondInternalServerError(ctx, err.Error())
		return
	}

	ctx.SetCookie("Access-token", token, 3600, "/", c.domain, false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"Result": http.StatusText(http.StatusOK),
	})
}

func (c *Controller) PostOrders(ctx *gin.Context) {
	if !strings.Contains(ctx.GetHeader("Content-type"), "text/plain") {
		helpers.RespondError(ctx, http.StatusBadRequest, "Only accepts text/plain")
		return
	}

	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		c.logger.Error("Error in handler", zap.String("handler", "PostOrders"), zap.Error(err))

		helpers.RespondInternalServerError(ctx, err.Error())
		return
	}

	if len(data) > MaxOrderIDLength {
		helpers.RespondError(ctx, http.StatusUnprocessableEntity, "Invalid order ID")
		return
	}

	orderID, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		helpers.RespondError(ctx, http.StatusUnprocessableEntity, "OrderResponse ID must be a number")
		return
	}

	userIDValue, ok := ctx.Get("userID")
	if !ok {
		helpers.RespondError(ctx, http.StatusUnauthorized, "User ID not found in context")
	}

	userID, ok := userIDValue.(string)
	if !ok {
		helpers.RespondInternalServerError(ctx, "User ID is not a string")
	}

	err = c.service.CreateOrder(ctx.Request.Context(), userID, orderID)
	if err != nil {
		if errors.Is(err, errs.ErrWrongOrderFormat) {
			helpers.RespondError(ctx, http.StatusUnprocessableEntity, "Invalid order ID")
			return
		}

		if errors.Is(err, errs.ErrConflictOrder) {
			helpers.RespondError(ctx, http.StatusConflict, err.Error())
			return
		}

		if errors.Is(err, errs.ErrDuplicateOrder) {
			ctx.JSON(http.StatusOK, gin.H{
				"Result": "This order is already uploaded",
			})
			return
		}
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"Result": http.StatusText(http.StatusAccepted),
	})
}

func (c *Controller) GetOrders(ctx *gin.Context) {
	userIDValue, ok := ctx.Get("userID")
	if !ok {
		helpers.RespondError(ctx, http.StatusUnauthorized, "User ID not found in context")
	}

	userID, ok := userIDValue.(string)
	if !ok {
		helpers.RespondInternalServerError(ctx, "User ID is not a string")
	}

	orders, err := c.service.GetOrders(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			ctx.Status(http.StatusNoContent)
			return
		}

		helpers.RespondInternalServerError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, orders)
}

func (c *Controller) GetBalance(ctx *gin.Context) {
	userIDValue, ok := ctx.Get("userID")
	if !ok {
		helpers.RespondError(ctx, http.StatusUnauthorized, "User ID not found in context")
	}

	userID, ok := userIDValue.(string)
	if !ok {
		helpers.RespondInternalServerError(ctx, "User ID is not a string")
	}

	balance, err := c.service.GetBalance(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			helpers.RespondError(ctx, http.StatusUnauthorized, err.Error())
			return
		}

		helpers.RespondInternalServerError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, balance)
}

func (c *Controller) Withdraw(ctx *gin.Context) {
	var data dto.OrderIn
	err := ctx.BindJSON(&data)
	if err != nil {
		helpers.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	orderID, err := strconv.ParseInt(data.OrderID, 10, 64)
	if err != nil {
		helpers.RespondError(ctx, http.StatusUnprocessableEntity, "Order ID must be a number")
		return
	}

	userIDValue, ok := ctx.Get("userID")
	if !ok {
		helpers.RespondError(ctx, http.StatusUnauthorized, "User ID not found in context")
	}

	userID, ok := userIDValue.(string)
	if !ok {
		helpers.RespondInternalServerError(ctx, "User ID is not a string")
	}

	sum := decimal.NewFromFloatWithExponent(data.Sum, -2)

	err = c.service.CreateOrderWithWithdrawal(ctx, userID, orderID, sum)
	if err != nil {
		if errors.Is(err, errs.ErrWrongOrderFormat) {
			helpers.RespondError(ctx, http.StatusUnprocessableEntity, "Invalid order ID")
			return
		}

		if errors.Is(err, errs.ErrConflictOrder) {
			helpers.RespondError(ctx, http.StatusConflict, err.Error())
			return
		}

		if errors.Is(err, errs.ErrInsufficientBalance) {
			helpers.RespondError(ctx, http.StatusPaymentRequired, "Insufficient balance")
			return
		}

		if errors.Is(err, errs.ErrDuplicateOrder) {
			ctx.JSON(http.StatusOK, gin.H{
				"Result": "This order is already uploaded",
			})
			return
		}

		c.logger.Error("Error in handler", zap.String("handler", "Withdraw"), zap.Error(err))

		helpers.RespondInternalServerError(ctx, err.Error())
		return
	}
}

func (c *Controller) GetWithdrawals(ctx *gin.Context) {
	userIDValue, ok := ctx.Get("userID")
	if !ok {
		helpers.RespondError(ctx, http.StatusUnauthorized, "User ID not found in context")
	}

	userID, ok := userIDValue.(string)
	if !ok {
		helpers.RespondInternalServerError(ctx, "User ID is not a string")
	}

	orders, err := c.service.GetWithdrawals(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, errs.ErrWithdrawalNotFound) {
			ctx.Status(http.StatusNoContent)
			return
		}

		helpers.RespondInternalServerError(ctx, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, orders)
}
