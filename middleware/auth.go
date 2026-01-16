package middleware

import (
	"RAG/config"
	"RAG/models"
	"RAG/pgk/database"
	"RAG/pgk/utils"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		tokenString := parts[1]
		cfg := config.GetConfig()
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTAccessSecret), nil
		})
		if err != nil || !token.Valid {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token")
			c.Abort()
			return
		}

		// If user is not available or redis cached not found return invalid token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token")
			c.Abort()
			return
		}
		userID := claims["id"].(string)
		ID, _ := uuid.Parse(userID)

		var user models.User
		if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			c.Abort()
			return
		}

		redisKey := fmt.Sprintf("rt:%s", userID)
		_, err = database.RedisClient.Get(context.Background(), redisKey).Result()
		if err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "User not found")
			c.Abort()
			return
		}

		c.Set("UserID", ID)
		c.Next()
	}
}
