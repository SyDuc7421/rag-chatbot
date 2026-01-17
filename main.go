package main

import (
	"RAG/config"
	"RAG/internal/routers"
	"RAG/pgk/database"
	"RAG/pgk/storage"
	"fmt"
	"log"
)

func main() {

	cfg := config.GetConfig()

	database.ConnectMySQL(cfg)
	database.ConnectRedis(cfg)
	database.Migrate(database.DB)
	storage.ConnectMinio()

	r := routers.SetupRouter()

	fmt.Printf("Server is running on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server")
	}

}
