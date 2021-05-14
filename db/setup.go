package db

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("POSTGRES_DSN")
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database!")
	}

	database.AutoMigrate(&User{})
	database.AutoMigrate(&Slot{})
	database.AutoMigrate(&UserSlot{})
	database.AutoMigrate(&District{})
	if err := database.SetupJoinTable(&User{}, "Slots", &UserSlot{}); err != nil {
		panic("Can't setup join table")
	}

	DB = database
}
