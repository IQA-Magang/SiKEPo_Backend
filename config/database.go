package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"backend/models"
)

var DB *gorm.DB

func ConnectDatabase() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default environment variables")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
	)

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	// Cek apakah tabel users sudah ada
	hasTable := database.Migrator().HasTable(&models.User{})

	if !hasTable {
		log.Println("Table 'users' belum ada. Membuat table...")

		err := database.AutoMigrate(&models.User{})
		if err != nil {
			panic(fmt.Sprintf("Failed to migrate users table: %v", err))
		}

		log.Println("Table 'users' berhasil dibuat!")
	} else {
		log.Println("Table 'users' sudah ada. AutoMigrate dilewati.")
	}

	DB = database

	log.Println("Database connected successfully!")
}