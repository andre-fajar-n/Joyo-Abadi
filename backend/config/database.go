package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/andre-fajar-n/Joyo-Abadi/backend/model/basemodel"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB() *gorm.DB {
	log.Println("Connecting DB...")

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second,   // Slow SQL threshold
			LogLevel:      logger.Silent, // Log level
			Colorful:      true,          // Enable color
		},
	)

	envPath := ".env"
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading .env file")
	}

	username := os.Getenv("USER_DATABASE")
	password := os.Getenv("PASS_DATABASE")
	host := os.Getenv("HOST_DATABASE")
	port := os.Getenv("PORT_DATABASE")
	name := os.Getenv("NAME_DATABASE")

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		username,
		password,
		host,
		port,
		name,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		panic("Failed to connect to database!")
	}
	log.Println("Database connection success!")

	return db.Debug()
}

func MigrateDB(db *gorm.DB) {
	log.Println("Start migrate")
	db.AutoMigrate(
		&basemodel.AccountRole{},
		&basemodel.Account{},
		&basemodel.Customer{},
	)
	log.Println("Finish migrate")
}
