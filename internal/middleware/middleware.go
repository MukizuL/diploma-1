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
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, &gin.H{
				"Error": http.StatusText(http.StatusInternalServerError),
			})
		}

		if errors.Is(err, http.ErrNoCookie) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, &gin.H{
				"Error": http.StatusText(http.StatusUnauthorized),
			})
		}

		userID, err := s.service.ValidateToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, &gin.H{
				"Error":   http.StatusText(http.StatusUnauthorized),
				"Message": "Access token is invalid",
			})
		}

		ctx.Set("userID", userID)

		ctx.Next()
	}
}
