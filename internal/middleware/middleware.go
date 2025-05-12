package middleware

import (
	"errors"
	"github.com/MukizuL/diploma-1/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
)

type MiddlewareService struct {
	service *services.Services
	logger  *zap.Logger
}

func NewMiddlewareService(service *services.Services, logger *zap.Logger) *MiddlewareService {
	return &MiddlewareService{
		service: service,
		logger:  logger,
	}
}

func Provide() fx.Option {
	return fx.Provide(NewMiddlewareService)
}

func (s *MiddlewareService) Authorization() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accessToken, err := ctx.Cookie("Access-token")
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			ctx.JSON(http.StatusInternalServerError, &gin.H{
				"Error": http.StatusText(http.StatusInternalServerError),
			})
			return
		}

		if errors.Is(err, http.ErrNoCookie) {
			ctx.JSON(http.StatusUnauthorized, &gin.H{
				"Error": http.StatusText(http.StatusUnauthorized),
			})
			return
		}

		userID, err := s.service.ValidateToken(accessToken)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, &gin.H{
				"Error":   http.StatusText(http.StatusInternalServerError),
				"Message": "Access token is invalid",
			})
			return
		}

		ctx.Set("userID", userID)

		ctx.Next()
	}
}
