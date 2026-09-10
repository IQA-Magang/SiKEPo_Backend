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
		panic(fmt.Sprintf(
			"Failed to connect to database: %v",
			err,
		))
	}

	// Cek tabel users
	hasTable := database.Migrator().HasTable(&models.User{})

	if !hasTable {

		log.Println("Table 'users' belum ada. Membuat table...")

		err := database.AutoMigrate(&models.User{})

		if err != nil {
			panic(fmt.Sprintf(
				"Failed to migrate users table: %v",
				err,
			))
		}

		log.Println("Table 'users' berhasil dibuat!")

	} else {

		log.Println("Table 'users' sudah ada.")

		// Tambahkan kolom baru jika belum ada
		hasPIC := database.Migrator().HasColumn(
			&models.User{},
			"pic",
		)

		if !hasPIC {

			log.Println("Column 'pic' belum ada. Menambahkan column...")

			err := database.Migrator().AddColumn(
				&models.User{},
				"PIC",
			)

			if err != nil {
				panic(fmt.Sprintf(
					"Failed to add pic column: %v",
					err,
				))
			}

			log.Println("Column 'pic' berhasil ditambahkan!")

		} else {

			log.Println("Column 'pic' sudah ada.")

		}
	}

	DB = database

	log.Println("Database connected successfully!")
}