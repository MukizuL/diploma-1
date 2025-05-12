package controller

import (
	"errors"
	"github.com/MukizuL/diploma-1/internal/dto"
	"github.com/MukizuL/diploma-1/internal/errs"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func (c *Controller) Register(ctx *gin.Context) {
	var data dto.AuthForm
	err := ctx.BindJSON(&data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &gin.H{
			"Error": http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	token, err := c.service.CreateUser(ctx.Request.Context(), data.Login, data.Password)
	if err != nil {
		if errors.Is(err, errs.ErrDuplicateLogin) {
			ctx.JSON(http.StatusConflict, &gin.H{
				"Error": err.Error(),
			})
		}

		c.logger.Error("Error in handler", zap.String("handler", "Register"), zap.Error(err))

		ctx.JSON(http.StatusInternalServerError, &gin.H{
			"Error": http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	ctx.SetCookie("Access-token", token, 3600, "/", "", true, true)

	ctx.JSON(http.StatusOK, &gin.H{
		"Result": http.StatusText(http.StatusOK),
	})
}

func (c *Controller) Login(ctx *gin.Context) {
	var data dto.AuthForm
	err := ctx.BindJSON(&data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &gin.H{
			"Error": http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	token, err := c.service.LoginUser(ctx.Request.Context(), data.Login, data.Password)
	if err != nil {
		if errors.Is(err, errs.ErrNotAuthorized) {
			ctx.JSON(http.StatusUnauthorized, &gin.H{
				"Error":   http.StatusText(http.StatusUnauthorized),
				"Message": "Access token is invalid",
			})
			return
		}

		c.logger.Error("Error in handler", zap.String("handler", "Login"), zap.Error(err))

		ctx.JSON(http.StatusInternalServerError, &gin.H{
			"Error": http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	ctx.SetCookie("Access-token", token, 3600, "/", "", true, true)

	ctx.JSON(http.StatusOK, &gin.H{
		"Result": http.StatusText(http.StatusOK),
	})
}
