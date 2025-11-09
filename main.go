package main

import (
	"httpServer/database"
	"httpServer/logger"
	"httpServer/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	DB, err := database.Connect()
	if err != nil {
		logger.LogError(err, "Ошибка при подключении к базе данных", logger.Error)
		return
	}
	r := gin.Default()
	routes.SetupRouter(r, DB)

	err = r.Run(":3000")
	if err != nil {
		logger.LogError(err, "Ошибка при запуске сервера", logger.Error)
	}
	logger.LogError(nil, "Сервер успешно запущен на порту 8080", logger.Info)
}
