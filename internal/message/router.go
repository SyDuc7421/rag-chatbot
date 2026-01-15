package message

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup) {
	message := rg.Group("/chat")
	{
		message.POST("", HandleResponseMessage)
	}
}
