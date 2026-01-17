package document

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup) {
	document := rg.Group("")
	{
		document.POST("/:conversation_id/document/upload", UploadDocument)
		document.GET("/:conversation_id/document/", GetAllDocuments)
		document.GET("/:conversation_id/document/:id", GetDocumentByID)
		document.GET("/:conversation_id/document/:id/url", GetDocumentPresignedURL)
		document.DELETE("/:conversation_id/document/:id", DeleteDocument)
	}
}
