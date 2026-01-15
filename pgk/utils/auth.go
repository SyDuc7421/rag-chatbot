package utils

import (
	"RAG/config"
	"RAG/models"
	"RAG/pgk/database"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateToken(user models.User) (string, string, error) {
	cfg := config.GetConfig()
	accessTokenClaims := jwt.MapClaims{
		"id":   user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 15).Unix(),
		"iat":  time.Now().Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	at, err := accessToken.SignedString([]byte(cfg.JWTAccessSecret))
	key := fmt.Sprintf("rt:%s", user.ID)

	var rt string
	existingRefreshToken, err := database.RedisClient.Get(context.Background(), key).Result()

	if err == nil && existingRefreshToken != "" {
		rt = existingRefreshToken
	} else {
		refreshTokenClaims := jwt.MapClaims{
			"id":  user.ID,
			"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
			"iat": time.Now().Unix(),
		}
		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
		rt, err = refreshToken.SignedString([]byte(cfg.JWTRefreshSecret))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = database.RedisClient.Set(ctx, "rt:"+user.ID.String(), rt, time.Hour*24*7).Err()
	}
	return at, rt, err
}

func RefreshToken(RefreshToken string) (string, string, error) {
	cfg := config.GetConfig()
	token, err := jwt.Parse(RefreshToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTRefreshSecret), nil
	})

	if err != nil {
		return "", "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !token.Valid || !ok {
		return "", "", errors.New("invalid refresh token")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	userID := fmt.Sprintf("%v", claims["id"])
	rt, err := database.RedisClient.Get(ctx, "rt:"+userID).Result()
	if err != nil {
		return "", "", err
	}

	var user models.User
	err = database.DB.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return "", "", errors.New("user not found")
	}

	accessTokenClaims := jwt.MapClaims{
		"id":   userID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 15).Unix(),
		"iat":  time.Now().Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	at, err := accessToken.SignedString([]byte(cfg.JWTAccessSecret))

	return at, rt, err
}
