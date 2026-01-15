package database

import (
	"RAG/config"
	"RAG/models"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectMySQL(cfg *config.Config) {
	var err error

	dsn := cfg.GetMySQLDSN()
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalln("Can not open mysql database")
	}
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf(err.Error())
	}

	sqlDB.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	fmt.Println("Connected mysql database")
}

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.User{},
		&models.Conversation{},
		&models.Message{},
	)
	if err != nil {
		log.Fatal("AutoMigrate failed:", err)
	}
}
