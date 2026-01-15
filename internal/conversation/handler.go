package conversation

import (
	"RAG/models"
	"RAG/pgk/database"
	"RAG/pgk/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateConversationInput struct {
	Title string `json:"title" binding:"required"`
}

type UpdateConversationInput struct {
	Title string `json:"title" binding:"required"`
}

func CreateNewConversation(c *gin.Context) {

	var input CreateConversationInput

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	UserID, ok := c.MustGet("UserID").(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	conversation := models.Conversation{
		Title:  input.Title,
		UserID: UserID,
	}
	if err := database.DB.Create(&conversation).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Success", gin.H{
		"conversation": conversation,
	})
}

func GetAllConversations(c *gin.Context) {
	var conversations []models.Conversation

	UserID, ok := c.MustGet("UserID").(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if err := database.DB.Where("user_id = ?", UserID).Find(&conversations).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{
		"conversations": conversations,
	})
}

func GetConversationByID(c *gin.Context) {
	id := c.Param("id")
	ConversationID, err := uuid.Parse(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Conversation ID")
		return
	}

	UserID, ok := c.MustGet("UserID").(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var conversation models.Conversation
	if err := database.DB.Where("id = ? AND user_id = ?", ConversationID, UserID).First(&conversation).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{
		"conversation": conversation,
	})
}

func UpdateConversation(c *gin.Context) {
	id := c.Param("id")
	ConversationID, err := uuid.Parse(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid Conversation ID")
		return
	}

	UserID, ok := c.MustGet("UserID").(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var input UpdateConversationInput
	if err := c.BindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	result := database.DB.Model(models.Conversation{}).Where("id = ? AND user_id = ?", ConversationID, UserID).Updates(map[string]interface{}{
		"title": input.Title,
	})
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Update conversation failed")
		return
	}
	if result.RowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "Not Found")
		return
	}

	var updatedConversation models.Conversation
	if err := database.DB.First(&updatedConversation, ConversationID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Can not find conversation")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{
		"conversation": updatedConversation,
	})

}

func DeleteConversation(c *gin.Context) {
	id := c.Param("id")
	ConversationID, err := uuid.Parse(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid conversation id")
		return
	}

	UserID, ok := c.MustGet("UserID").(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	result := database.DB.Model(models.Conversation{}).Where("id = ? AND user_id = ?", ConversationID, UserID).Delete(&models.Conversation{})
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Can't delete conversation")
		return
	}
	if result.RowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "Not Found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", nil)
}
