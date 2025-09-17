package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/Morningstarl2504/BalkanID_repo/internal/models"
)

var DB *gorm.DB

func Connect(databaseURL string) error {
	var err error
	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	log.Println("Database connected successfully")
	return nil
}

func Migrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.FileContent{},
		&models.File{},
		&models.Folder{},
		&models.FileShare{},
		&models.AuditLog{},
	)
}