package document

import (
	"RAG/config"
	"RAG/models"
	"RAG/pgk/database"
	"RAG/pgk/storage"
	"RAG/pgk/utils"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

const MaxUploadSize = 10 << 20 // 10 MB
var allowExtension = map[string]bool{
	".pdf":  true,
	".docx": true,
}

func UploadDocument(c *gin.Context) {
	ConversationIDParam, _ := c.Params.Get("conversation_id")
	ConversationID, err := uuid.Parse(ConversationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID not valid")
		return
	}

	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		MaxUploadSize,
	)
	if err := c.Request.ParseMultipartForm(MaxUploadSize); err != nil {
		if err.Error() == "http: request body too large" {
			utils.ErrorResponse(c, http.StatusRequestEntityTooLarge, "Upload too large, maximum 10MB")
			return
		}
		utils.ErrorResponse(c, http.StatusBadRequest, "Upload file not valid")
		return
	}

	file, err := c.FormFile("document")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File not found")
		return
	}

	extFile := strings.ToLower(filepath.Ext(file.Filename))
	if !allowExtension[extFile] {
		utils.ErrorResponse(c, http.StatusBadRequest, "File extension not allowed")
		return
	}
	src, err := file.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Can not open file")
		return
	}
	defer func(src multipart.File) {
		_ = src.Close()
	}(src)

	cfg := config.GetConfig()
	objectName := ConversationID.String() + "_" + file.Filename
	contentType := file.Header.Get("Content-Type")

	info, err := storage.MinioClient.PutObject(c.Request.Context(),
		cfg.Minio.BucketName,
		objectName,
		src,
		file.Size,
		minio.PutObjectOptions{ContentType: contentType})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error when upload to minio"+err.Error())
		return
	}
	document := models.Document{
		Name:           file.Filename,
		ConversationID: ConversationID,
		SourceType:     contentType,
		SourceUri:      info.Key,
	}
	if err := database.DB.Create(&document).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error when upload to database")
		return
	}

	log.Println("Successfully uploaded file to " + objectName)
	utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{
		"document": document,
	})

}

func GetPresignedURL(ctx context.Context, key string) (string, error) {
	cfg := config.GetConfig()
	expiry := time.Minute * 15

	presignedURL, err := storage.MinioClient.PresignedGetObject(ctx,
		cfg.Minio.BucketName,
		key,
		expiry,
		nil)
	if err != nil {
		log.Println("MinIO Presigned GetObject Error", err.Error())
		return "", errors.New("error when get presigned url: " + err.Error())
	}
	return presignedURL.String(), nil
}

func GetObjectBuffer(ctx context.Context, key string) ([]byte, int64, error) {
	cfg := config.GetConfig()

	info, err := storage.MinioClient.StatObject(
		ctx,
		cfg.Minio.BucketName,
		key,
		minio.StatObjectOptions{},
	)
	if err != nil {
		return nil, 0, errors.New("error when stat object: " + err.Error())
	}

	object, err := storage.MinioClient.GetObject(ctx, cfg.Minio.BucketName, key, minio.GetObjectOptions{})

	if err != nil {
		return nil, 0, errors.New("error when get object: " + err.Error())
	}
	defer func(object *minio.Object) {
		err := object.Close()
		if err != nil {

		}
	}(object)

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, 0, errors.New("error when read object: " + err.Error())
	}
	return data, info.Size, nil
}

func GetAllDocuments(c *gin.Context) {
	ConversationIDParam, _ := c.Params.Get("conversation_id")
	ConversationID, err := uuid.Parse(ConversationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID not valid")
		return
	}
	// Need to check author?
	var documents []models.Document
	if err := database.DB.Where("conversation_id = ?", ConversationID).Find(&documents).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error when get all documents")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{
		"documents": documents,
	})
}

func GetDocumentByID(c *gin.Context) {
	ConversationIDParam, _ := c.Params.Get("conversation_id")
	ConversationID, err := uuid.Parse(ConversationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID not valid")
		return
	}

	IDParam, _ := c.Params.Get("id")
	ID, err := uuid.Parse(IDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID not valid")
		return
	}
	var document models.Document

	if err := database.DB.Where("id = ? AND conversation_id = ?", ID, ConversationID).First(&document).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error when get document by id")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{
		"document": document,
	})
}

func GetDocumentPresignedURL(c *gin.Context) {
	ConversationIDParam, _ := c.Params.Get("conversation_id")
	ConversationID, err := uuid.Parse(ConversationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID not valid")
		return
	}

	IDParam, _ := c.Params.Get("id")
	ID, err := uuid.Parse(IDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID not valid")
		return
	}
	var document models.Document

	if err := database.DB.Where("id = ? AND conversation_id = ?", ID, ConversationID).First(&document).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error when get document by id")
		return
	}
	presignedURL, err := GetPresignedURL(c, document.SourceUri)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error when get presigned url")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", gin.H{
		"document": document,
		"url":      presignedURL,
	})
}

func DeleteObject(ctx *gin.Context, key string) error {
	cfg := config.GetConfig()

	err := storage.MinioClient.RemoveObject(
		ctx,
		cfg.Minio.BucketName,
		key,
		minio.RemoveObjectOptions{},
	)

	if err != nil {
		log.Printf("MinIO Remove Object %s Error: %s", key, err.Error())
		return fmt.Errorf("MinIO Remove Object %s Error: %s", key, err.Error())
	}
	return nil
}

func DeleteDocument(c *gin.Context) {
	ConversationIDParam, _ := c.Params.Get("conversation_id")
	ConversationID, err := uuid.Parse(ConversationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID not valid")
		return
	}

	IDParam, _ := c.Params.Get("id")
	ID, err := uuid.Parse(IDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID not valid")
		return
	}

	UserID, ok := c.MustGet("UserID").(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized User")
		return
	}

	var document models.Document
	if err := database.DB.
		Where("id = ? AND conversation_id = ?", ID, ConversationID).
		Where("conversation_id IN (?)", database.DB.Model(&models.Conversation{}).Select("id").Where("user_id = ?", UserID)).
		First(&document).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Document not found")
		return
	}
	documentKey := document.SourceUri

	result := database.DB.Model(&models.Document{}).
		Where("id = ? AND conversation_id = ?", ID, ConversationID).
		Where("conversation_id IN (?)", database.DB.Model(&models.Conversation{}).Select("id").Where("user_id = ?", UserID)).
		Delete(&models.Document{})
	if result.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error when delete document in database")
		return
	}

	// If this case happened document is not found or user have no authorized? but we check document above logic
	if result.RowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "User unauthorized")
		return
	}

	err = DeleteObject(c, documentKey)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error when delete document in storge")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success", nil)

}

func IngestDocument(c *gin.Context) {
	ConversationIDParam, _ := c.Params.Get("conversation_id")
	ConversationID, err := uuid.Parse(ConversationIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID not valid")
		return
	}

	IDParam, _ := c.Params.Get("id")
	ID, err := uuid.Parse(IDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Conversation ID not valid")
		return
	}

	UserID, ok := c.MustGet("UserID").(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized User")
		return
	}

	var document models.Document
	if err := database.DB.
		Where("id = ? AND conversation_id = ?", ID, ConversationID).
		Where("conversation_id IN (?)", database.DB.Model(&models.Conversation{}).Select("id").Where("user_id = ?", UserID)).
		First(&document).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Document not found")
		return
	}
	data, size, err := GetObjectBuffer(c, document.SourceUri)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error when get object")
		return
	}

	rawText, err := utils.ExtractTextFromBuffer(data, size)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error when get extract text")
		return
	}

	docs, err := utils.PrepareDocuments(rawText, document)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error when prepare documents")
		return
	}

	if err := utils.SaveToQdrant(c, docs); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Error when save documents to qdrant store")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Success", nil)
}
