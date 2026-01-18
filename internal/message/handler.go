package message

import (
	"RAG/internal/ai"
	"RAG/models"
	"RAG/pgk/database"
	"RAG/pgk/utils"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateMessageInput struct {
	Message string `json:"message" binding:"required"`
}

func HandleResponseMessage(c *gin.Context) {
	var input CreateMessageInput
	if err := c.BindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	ConversationIDParam, _ := c.Params.Get("conversation_id")
	ConversationID, err := uuid.Parse(ConversationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID is invalid")
		return
	}

	UserID, ok := c.MustGet("UserID").(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var conv models.Conversation
	if err := database.DB.Where("id = ? AND user_id = ?", ConversationID, UserID).First(&conv).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Conversation not found")
		return
	}

	docs, err := utils.SearchInQdrant(c, input.Message, conv.ID.String())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to search in qdrant")
		return
	}

	var messages []models.Message
	if err := database.DB.Where("conversation_id = ?", conv.ID).Order("created_at DESC").Limit(10).Find(&messages).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Conversation message not found")
		return
	}
	// Reverse messages because get in DESC
	slices.Reverse(messages)
	response, totalToken, err := ai.GetResponse(messages, docs, input.Message)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error while getting response")
		return
	}

	// Save new Massage
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		tx.Create(&models.Message{
			Sender:         utils.BoolPtr(true),
			Content:        input.Message,
			ConversationID: conv.ID,
			TokenCount:     0,
		})
		tx.Create(&models.Message{
			Sender:         utils.BoolPtr(false),
			Content:        response,
			ConversationID: conv.ID,
			TokenCount:     totalToken,
		})
		return nil
	}); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error while creating message")
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{
		"response": response,
	})
}

func HandleGetAllMessages(c *gin.Context) {
	ConversationIDParam, _ := c.Params.Get("conversation_id")
	ConversationID, err := uuid.Parse(ConversationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID not found")
		return
	}

	var messages []models.Message
	if err := database.DB.Where("conversation_id = ?", ConversationID).Order("created_at ASC").Find(&messages).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Messages not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{
		"messages": messages,
	})

}

func HandleGetMessageByID(c *gin.Context) {
	ConversationIDParam, _ := c.Params.Get("conversation_id")
	ConversationID, err := uuid.Parse(ConversationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID not found")
		return
	}

	ID, _ := c.Params.Get("id")
	MessageID, err := uuid.Parse(ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Message ID not found")
		return
	}

	var message models.Message

	if err := database.DB.Where("id = ? AND conversation_id = ?", MessageID, ConversationID).First(&message).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Message not found")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{"message": message})
}

func HandleDeleteMessage(c *gin.Context) {
	ConversationIDParam, _ := c.Params.Get("conversation_id")
	ConversationID, err := uuid.Parse(ConversationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID not found")
		return
	}

	ID, _ := c.Params.Get("id")
	MessageID, err := uuid.Parse(ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Message ID not found")
		return
	}

	UserID, ok := c.MustGet("UserID").(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	result := database.DB.Where("id = ? AND conversation_id = ?", MessageID, ConversationID).
		Where("conversation_id IN (?)",
			database.DB.Model(&models.Conversation{}).Select("id").Where("user_id = ?", UserID),
		).
		Delete(&models.Message{})
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Cannot delete message")
		return
	}
	if result.RowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "Message not found")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", nil)
}
