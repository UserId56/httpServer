package database

import (
	"errors"
	"fmt"
	"httpServer/logger"
	"httpServer/models"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func seedDefaultData(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, roleName := range []string{"admin", "user"} {
			var role models.Role
			if err := tx.Where("name = ?", roleName).First(&role).Error; errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&models.Role{Name: roleName}).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}
		var adminRole models.Role
		if err := tx.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
			return err
		}
		adminEmail := os.Getenv("ADMIN_EMAIL")
		if adminEmail == "" {
			adminEmail = "admin@example.com"
		}

		adminPassword := os.Getenv("ADMIN_PASSWORD")
		if adminPassword == "" {
			adminPassword = "admin123"
		}

		var adminUser models.User
		if err := tx.Where("username = ?", "admin").First(&adminUser).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			adminUser = models.User{
				Username: "admin",
				Email:    adminEmail,
				Password: string(hashedPassword),
				RoleID:   &adminRole.ID,
			}
			if err := tx.Create(&adminUser).Error; err != nil {
				return err
			}
		}
		var usersScheme models.DynamicScheme
		if err := tx.Where("name = ?", "users").First(&usersScheme).Error; errors.Is(err, gorm.ErrRecordNotFound) {

			usersScheme = models.DynamicScheme{
				Name:        "users",
				DisplayName: "Пользователи",
			}
			if err := tx.Create(&usersScheme).Error; err != nil {
				return err
			}
		}
		var rolesScheme models.DynamicScheme
		if err := tx.Where("name = ?", "roles").First(&rolesScheme).Error; errors.Is(err, gorm.ErrRecordNotFound) {

			rolesScheme = models.DynamicScheme{
				Name:        "roles",
				DisplayName: "Роли",
			}
			if err := tx.Create(&rolesScheme).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func Connect() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable timezone=Europe/Moscow",
		"localhost",                     // ваш_хост
		os.Getenv("SERVER_BD_USER"),     // имя_пользователя
		os.Getenv("SERVER_DB_PASSWORD"), // ваш_пароль
		os.Getenv("SERVER_DB_NAME"),     // имя_базы_данных
		os.Getenv("SERVER_DB_PORT"),     // порт_pg (по умолчанию 5432)
	)
	fmt.Println(dsn)
	newLogger := gormLogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // Выводим в stdout
		gormLogger.Config{
			SlowThreshold: time.Millisecond,
			LogLevel:      gormLogger.Info, // <--- Установите Level на Info
			Colorful:      true,
		},
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}
	logger.Log(nil, "Подключение к базе данных успешно установлено", logger.Info)

	err = db.AutoMigrate(models.Models()...)
	err = seedDefaultData(db)
	if err != nil {
		logger.Log(err, "Ошибка при инициализации данных", logger.Error)
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	logger.Log(nil, "База данных успешно мигрирована", logger.Info)
	return db, nil
}
