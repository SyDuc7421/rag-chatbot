package message

import (
	"RAG/internal/ai"
	"RAG/pgk/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateMessageInput struct {
	Message string `json:"message" binding:"required"`
}

func HandleResponseMessage(c *gin.Context) {
	var input CreateMessageInput
	if err := c.BindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
	response, err := ai.GetResponse(input.Message)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error while getting response")
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{"response": response})
}
