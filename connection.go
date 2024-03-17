package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
)

func GetDB() *gorm.DB {
	dsn := fmt.Sprintf(
		"%s:%s@tcp4(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_DATABASE"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(&Flat{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&Subscription{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&SearchConfig{})
	if err != nil {
		panic(err)
	}

	return db
}
