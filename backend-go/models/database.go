package models

import (
	"log"
	"parking-system-go/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	var err error
	cfg := config.GetConfig()

	if cfg.Database.Type == "sqlite" {
		DB, err = gorm.Open(sqlite.Open(cfg.Database.DBName), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")
}

func AutoMigrate() {
	err := DB.AutoMigrate(
		&ParkingLot{},
		&ParkingSpot{},
		&Reservation{},
		&Order{},
		&BillingRule{},
		&BillingDetail{},
		&Device{},
		&AccessLog{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrated successfully")
}
