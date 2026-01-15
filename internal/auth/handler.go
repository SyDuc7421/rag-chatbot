package auth

import (
	"RAG/models"
	"RAG/pgk/database"
	"RAG/pgk/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RegisterInput struct {
	FullName string `json:"full_name"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func RegisterHandler(c *gin.Context) {
	var input RegisterInput
	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Hash password
	hashPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Create and save User
	user := models.User{
		FullName: input.FullName,
		Email:    input.Email,
		Password: hashPassword,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Can not create user")
		return
	}
	// Return
	utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{
		"user": user,
	})
}

func LoginHandler(c *gin.Context) {
	var input LoginInput
	// Validate Input
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User

	// Check email
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Email or Password is incorrect")
		return
	}

	// Check hash Password
	if !utils.CheckPasswordHash(input.Password, user.Password) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Email or Password is incorrect")
		return
	}

	accessToken, refreshToken, _ := utils.GenerateToken(user)

	utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})

}

func RefreshHandler(c *gin.Context) {
	var input RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	accessToken, refreshToken, err := utils.RefreshToken(input.RefreshToken)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
