package routers

import (
	"RAG/internal/auth"
	"RAG/internal/conversation"
	"RAG/internal/document"
	"RAG/internal/message"
	"RAG/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/api/v1")

	v1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	auth.RegisterRoutes(v1)

	protected := v1.Group("/")
	protected.Use(middleware.AuthRequired())
	{
		conversation.RegisterRoutes(protected)
		message.RegisterRoutes(protected)
		document.RegisterRoutes(protected)
	}

	return router
}
