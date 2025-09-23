package db

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDb(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Konfigurasi pool (mirip pgxpool)
	sqlDB.SetMaxOpenConns(25)                 // maksimum koneksi terbuka
	sqlDB.SetMaxIdleConns(25)                 // maksimum koneksi idle
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // umur maksimal koneksi

	// Cek koneksi
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return db, nil

}
