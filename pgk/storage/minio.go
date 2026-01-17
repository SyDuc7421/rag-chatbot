package storage

import (
	"RAG/config"
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	_ "github.com/minio/minio-go/v7/pkg/credentials"
)

var MinioClient *minio.Client

func ConnectMinio() {
	cfg := config.GetConfig()

	client, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKey, cfg.Minio.SecretKey, ""),
		Secure: cfg.Minio.UseSSL,
	})

	if err != nil {
		log.Fatalln("Error connecting to minio:", err)
	}
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Minio.BucketName)
	if err == nil && !exists {
		err = client.MakeBucket(ctx, cfg.Minio.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalln("Error creating bucket:", err)
		}
		log.Println("Successfully created bucket:", cfg.Minio.BucketName)
	}
	MinioClient = client
	log.Println("Successfully connected to minio")
}
