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
		logger.Log(err, "Ошибка при подключении к базе данных", logger.Error)
		return
	}
	r := gin.Default()
	routes.SetupRouter(r, DB)

	err = r.Run(":3000")
	if err != nil {
		logger.Log(err, "Ошибка при запуске сервера", logger.Error)
	}
}
