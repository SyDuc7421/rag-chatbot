package ai

import (
	"RAG/config"
	"RAG/models"
	"context"

	"github.com/sashabaranov/go-openai"
)

func GetResponse(messages []models.Message, newMessage string) (string, int, error) {
	cfg := config.GetConfig()
	client := openai.NewClient(cfg.OPENAIAPIKey)

	var apiMessages []openai.ChatCompletionMessage

	apiMessages = append(apiMessages, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
		Content: "Answer the user's question using only the provided context." +
			"If the context is insufficient, say you do not know." +
			"Do not make up information." +
			"Be concise, accurate, and professional." +
			"Do not reference the context or internal systems.",
	})

	//	Loop in input is a list message
	for _, message := range messages {
		Role := openai.ChatMessageRoleAssistant
		if *message.Sender {
			Role = openai.ChatMessageRoleUser
		}

		apiMessages = append(apiMessages, openai.ChatCompletionMessage{
			Role:    Role,
			Content: message.Content,
		})
	}

	apiMessages = append(apiMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: newMessage,
	})

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    cfg.OPENAIModel,
			Messages: apiMessages,
		},
	)
	if err != nil {
		return "", 0, err
	}
	return resp.Choices[0].Message.Content, resp.Usage.TotalTokens, nil
}
