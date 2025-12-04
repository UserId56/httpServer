package core

import (
	"github.com/UserId56/httpServer/core/database"
	"github.com/UserId56/httpServer/core/logger"
	"github.com/UserId56/httpServer/core/plugins"
	"github.com/UserId56/httpServer/core/routes"
	"github.com/gin-gonic/gin"
)

func ServerInit(plugins []plugins.Plugin) {
	DB, err := database.Connect()
	if err != nil {
		logger.Log(err, "Ошибка при подключении к базе данных", logger.Error)
		return
	}
	sqlDB, err := DB.DB()
	if err != nil {
		logger.Log(err, "Ошибка при получении экземпляра базы данных", logger.Error)
		return
	}
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetMaxOpenConns(250)
	r := gin.Default()
	routes.SetupRouter(r, DB)
	PluginAPI := make(map[string]interface{})
	for _, plugin := range plugins {
		err := plugin.PluginInit(r, DB, &PluginAPI)
		if err != nil {
			logger.Log(err, "Ошибка при инициализации плагина", logger.Error)
		}
	}
	err = r.Run(":3000")
	if err != nil {
		logger.Log(err, "Ошибка при запуске сервера", logger.Error)
	}
}
