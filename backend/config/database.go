package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBConfig struct {
	Username string
	Password string
	Host     string
	Port     string
	Name     string
}

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

	// username := os.Getenv("USER_DATABASE")
	// password := os.Getenv("PASS_DATABASE")
	// host := os.Getenv("HOST_DATABASE")
	// port := os.Getenv("PORT_DATABASE")
	// name := os.Getenv("NAME_DATABASE")

	dsn := "host=127.0.0.1 user=username password=password dbname=joyoabadidb port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	fmt.Println(db, err)
	if err != nil {
		panic("Failed to connect to database!")
	}
	log.Println("Database connection success!")

	return db.Debug()
}
