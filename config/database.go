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

	// Cek apakah tabel-tabel penting sudah ada
	hasUserTable := database.Migrator().HasTable(&models.User{})
	hasRuanganTable := database.Migrator().HasTable(&models.Ruangan{})
	hasLabsTable := database.Migrator().HasTable(&models.Labs{})
	hasPeralatanTable := database.Migrator().HasTable(&models.Peralatan{})

	if !hasUserTable || !hasRuanganTable || !hasLabsTable || !hasPeralatanTable {
		log.Println("Beberapa tabel belum ada. Membuat tabel...")

		err := database.AutoMigrate(
			&models.User{},
			&models.Ruangan{},
			&models.Labs{},
			&models.Peralatan{},
			&models.DetailAlatUkur{},
			&models.DetailAlatBantu{},
			&models.DetailArtefakAcuan{},
			&models.DetailKomponenPendukung{},
		)
		if err != nil {
			panic(fmt.Sprintf("Failed to migrate database tables: %v", err))
		}

		log.Println("Semua tabel berhasil dibuat!")
	} else {
		log.Println("Semua tabel sudah ada. AutoMigrate dilewati.")
	}

	DB = database

	log.Println("Database connected successfully!")
}
