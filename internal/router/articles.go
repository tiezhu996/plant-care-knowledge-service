package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/handler"
	"github.com/gbplantwiki/gbplantwiki/internal/middleware"
)

func registerArticleRoutes(v1 *gin.RouterGroup, auth gin.HandlerFunc, h *handler.CareArticleHandler, limiter *middleware.RateLimiter) {
	articles := v1.Group("/articles")
	articles.GET("", h.List)
	articles.GET("/:id", h.Get)
	authGrp := articles.Group("", auth)
	authGrp.POST("", limiter.Limit(), h.Create)
	authGrp.PUT("/:id", h.Update)
	authGrp.DELETE("/:id", h.Delete)
}
