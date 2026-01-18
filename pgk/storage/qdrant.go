package storage

import (
	"RAG/config"
	"context"
	"log"

	"github.com/qdrant/go-client/qdrant"
)

var QdrantClient *qdrant.Client

func ConnectQdrant() {
	cfg := config.GetConfig()

	client, err := qdrant.NewClient(&qdrant.Config{
		Host: cfg.Qdrant.Host,
		Port: cfg.Qdrant.GRPCPort,
	})

	if err != nil {
		log.Fatalf("Can not connect to Qdrant: %s", err)
	}
	QdrantClient = client
	CreateCollectionIfNotExists(cfg.Qdrant.CollectionName, cfg.EmbeddingModel.Dimension)
	log.Println("Connected to Qdrant")
}

func CreateCollectionIfNotExists(name string, dim int) {
	ctx := context.Background()

	exists, err := QdrantClient.CollectionExists(ctx, name)
	if err != nil || !exists {
		err = QdrantClient.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: name,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     uint64(dim),
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			log.Fatalf("Can not create Qdrant collection: %s", err)
		}
	}
}
