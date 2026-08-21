package router

import (
	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/handler"
	"github.com/gbplantwiki/gbplantwiki/internal/middleware"
)

func registerQuestionRoutes(v1 *gin.RouterGroup, auth gin.HandlerFunc, qh *handler.QuestionHandler, ah *handler.AnswerHandler, limiter *middleware.RateLimiter) {
	questions := v1.Group("/questions")
	questions.GET("", qh.List)
	questions.GET("/:id", qh.Get)
	questions.GET("/:id/answers", ah.List)
	authGrp := questions.Group("", auth)
	authGrp.POST("", limiter.Limit(), qh.Create)
	authGrp.POST("/:id/answers", ah.Create)
	authGrp.PUT("/:id/adopt", ah.Adopt)
}
