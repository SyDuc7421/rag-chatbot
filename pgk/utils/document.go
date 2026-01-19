package utils

import (
	"RAG/config"
	"RAG/models"
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/ledongthuc/pdf"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/qdrant"
)

func ExtractTextFromBuffer(data []byte, size int64) (string, error) {
	reader := bytes.NewReader(data)
	f, err := pdf.NewReader(reader, size)
	if err != nil {
		return "", fmt.Errorf("can not create pdf reader: %s", err)
	}

	var content bytes.Buffer
	numPages := f.NumPage()
	for pageIndex := 1; pageIndex <= numPages; pageIndex++ {
		p := f.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		text, _ := p.GetPlainText(nil)

		content.WriteString(text)
		content.WriteString("\n")
	}
	return content.String(), nil
}

func PrepareDocuments(rawText string, document models.Document) ([]schema.Document, error) {

	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(1000),
		textsplitter.WithChunkOverlap(150),
	)
	chunks, err := splitter.SplitText(rawText)
	if err != nil {
		return nil, err
	}

	docs := make([]schema.Document, 0, len(chunks))
	for _, chunk := range chunks {
		docs = append(docs, schema.Document{
			PageContent: chunk,
			Metadata: map[string]interface{}{
				"conversation_id": document.ConversationID,
				"document_id":     document.ID,
				"name":            document.Name,
			},
		})
	}
	return docs, nil
}

func SaveToQdrant(ctx context.Context, docs []schema.Document) error {
	cfg := config.GetConfig()

	llm, err := openai.New()
	if err != nil {
		return fmt.Errorf("failed to init openai llm: %w", err)
	}
	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return fmt.Errorf("failed to init embedder: %w", err)
	}

	rawAddr := fmt.Sprintf("http://%s:%d", cfg.Qdrant.Host, cfg.Qdrant.HTTPPort)

	parsedURL, err := url.Parse(rawAddr)
	if err != nil {
		return fmt.Errorf("invalid qdrant address: %w", err)
	}
	store, err := qdrant.New(
		qdrant.WithURL(*parsedURL),
		qdrant.WithCollectionName(cfg.Qdrant.CollectionName),
		qdrant.WithEmbedder(embedder),
	)
	if err != nil {
		return fmt.Errorf("failed to init langchaingo qdrant store: %w", err)
	}

	_, err = store.AddDocuments(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to add documents to qdrant: %w", err)
	}

	return nil
}

func SearchInQdrant(ctx context.Context, userQuery string, convID string) ([]schema.Document, error) {
	cfg := config.GetConfig()

	llm, err := openai.New()
	if err != nil {
		return nil, fmt.Errorf("failed to init llm: %w", err)
	}
	embedder, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, fmt.Errorf("failed to init embedder: %w", err)
	}
	rawAddr := fmt.Sprintf("http://%s:%d", cfg.Qdrant.Host, cfg.Qdrant.HTTPPort)

	parsedURL, err := url.Parse(rawAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid qdrant address: %w", err)
	}
	store, err := qdrant.New(
		qdrant.WithURL(*parsedURL),
		qdrant.WithCollectionName(cfg.Qdrant.CollectionName),
		qdrant.WithEmbedder(embedder),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to init qdrant store: %w", err)
	}

	docs, err := store.SimilaritySearch(ctx, userQuery, 5,
		vectorstores.WithFilters(map[string]any{
			"must": []map[string]any{
				{
					"key": "conversation_id",
					"match": map[string]any{
						"value": convID,
					},
				},
			},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search in qdrant: %w", err)
	}

	return docs, nil
}
