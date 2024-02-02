package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/YehyeokBang/Simple-SNS/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func MustNewGormDB(config *config.Config) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Fatalf("failed to migrate user: %v", err)
	}

	err = db.AutoMigrate(&Post{})
	if err != nil {
		log.Fatalf("failed to migrate post: %v", err)
	}

	err = db.AutoMigrate(&Comment{})
	if err != nil {
		log.Fatalf("failed to migrate comment: %v", err)
	}

	return db
}
