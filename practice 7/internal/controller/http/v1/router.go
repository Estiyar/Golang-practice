package v1

import (
	"practice-7/internal/usecase"
	"practice-7/utils"

	"github.com/gin-gonic/gin"
)

func NewRouter(handler *gin.Engine, t usecase.UserInterface) {
	handler.Use(utils.RateLimiterMiddleware())
	v1 := handler.Group("/v1")
	newUserRoutes(v1, t)
}
