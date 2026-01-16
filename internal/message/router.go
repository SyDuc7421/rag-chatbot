package message

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup) {
	message := rg.Group("")
	{
		message.POST("/:conversation_id/message", HandleResponseMessage)
		message.GET("/:conversation_id/message", HandleGetAllMessages)
		message.GET("/:conversation_id/message/:id", HandleGetMessageByID)
		message.DELETE("/:conversation_id/message/:id", HandleDeleteMessage)
	}
}
