package config

import (
	"fmt"
	"log"
	"sync"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type MySQL struct {
	Host         string `env:"DB_HOST" required:"true"`
	Port         string `env:"DB_PORT" envDefault:"3306"`
	User         string `env:"DB_USER" required:"true"`
	Password     string `env:"DB_PASSWORD" required:"true"`
	DBName       string `env:"DB_NAME" required:"true"`
	MaxIdleConns int    `env:"DB_MAX_IDLE_CONNS" envDefault:"10"`
	MaxOpenConns int    `env:"DB_MAX_OPEN_CONNS" envDefault:"100"`
}

type Redis struct {
	Host     string `env:"REDIS_HOST" required:"true"`
	Port     string `env:"REDIS_PORT" envDefault:"6379"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

type Minio struct {
	Endpoint   string `env:"MINIO_ENDPOINT" required:"true"`
	AccessKey  string `env:"MINIO_ACCESS_KEY" required:"true"`
	SecretKey  string `env:"MINIO_SECRET_KEY" required:"true"`
	UseSSL     bool   `env:"MINIO_USE_SSL" envDefault:"false"`
	BucketName string `env:"MINIO_BUCKET_NAME" required:"true"`
}

type Qdrant struct {
	Host           string `env:"QDRANT_HOST" required:"true"`
	GRPCPort       int    `env:"QDRANT_GRPC_PORT" required:"true"`
	HTTPPort       int    `env:"QDRANT_HTTP_PORT" required:"true"`
	CollectionName string `env:"QDRANT_COLLECTION_NAME" required:"true"`
}

type EmbeddingModel struct {
	OpenAIEmbeddingModel string `env:"OPENAI_EMBEDDING_MODEL" required:"true"`
	Dimension            int    `env:"EMBEDDING_DIMENSION" required:"true"`
}

type Config struct {
	AppEnv           string `env:"APP_ENV" envDefault:"development"`
	Port             string `env:"APP_PORT" envDefault:"8080"`
	JWTAccessSecret  string `env:"JWT_ACCESS_SECRET" required:"true"`
	JWTRefreshSecret string `env:"JWT_REFRESH_SECRET" required:"true"`
	OPENAIAPIKey     string `env:"OPENAI_API_KEY" required:"true"`
	OPENAIModel      string `env:"OPENAI_MODEL" required:"true"`
	MySQL            MySQL
	Redis            Redis
	Minio            Minio
	Qdrant           Qdrant
	EmbeddingModel   EmbeddingModel
}

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		_ = godotenv.Load()
		instance = &Config{}
		if err := env.Parse(instance); err != nil {
			log.Fatal("Error parsing env variables")
		}
		log.Printf("Successfully loaded config")
	})
	return instance
}

func (c *Config) GetMySQLDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s&readTimeout=5s&writeTimeout=5s",
		c.MySQL.User,
		c.MySQL.Password,
		c.MySQL.Host,
		c.MySQL.Port,
		c.MySQL.DBName,
	)
}
