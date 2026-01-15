package ai

import (
	"RAG/config"
	"context"

	"github.com/sashabaranov/go-openai"
)

func GetResponse(message string) (string, error) {
	cfg := config.GetConfig()
	client := openai.NewClient(cfg.OPENAIAPIKey)

	var apiMessages []openai.ChatCompletionMessage

	apiMessages = append(apiMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: "Bạn là 1 chatbox rag thông tin. Hảy trả lời ngắn gọn, chích xác và hữu ích",
	})

	//	Loop in input is a list message
	apiMessages = append(apiMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    cfg.OPENAIMODEL,
			Messages: apiMessages,
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
