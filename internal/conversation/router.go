package conversation

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup) {
	conversations := rg.Group("/conversations")
	{
		conversations.GET("", GetAllConversations)
		conversations.GET("/:id", GetConversationByID)
		conversations.POST("", CreateNewConversation)
		conversations.PUT("/:id", UpdateConversation)
		conversations.DELETE("/:id", DeleteConversation)
	}
}
