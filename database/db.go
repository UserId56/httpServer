package database

import (
	"fmt"
	"httpServer/models"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//var DB *gorm.DB

func Connect() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable timezone=Europe/Moscow",
		"localhost",                     // ваш_хост
		os.Getenv("SERVER_BD_USER"),     // имя_пользователя
		os.Getenv("SERVER_DB_PASSWORD"), // ваш_пароль
		os.Getenv("SERVER_DB_NAME"),     // имя_базы_данных
		os.Getenv("SERVER_DB_PORT"),     // порт_pg (по умолчанию 5432)
	)
	fmt.Println(dsn)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	fmt.Println("Успешно подключено к базе данных")

	err = db.AutoMigrate(models.Models()...)
	if err != nil {
		return nil, err
	}
	fmt.Println("Миграция базы данных выполнена успешно")
	return db, nil
}
