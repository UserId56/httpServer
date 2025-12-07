package core

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/UserId56/httpServer/core/database"
	"github.com/UserId56/httpServer/core/logger"
	"github.com/UserId56/httpServer/core/plugins"
	"github.com/UserId56/httpServer/core/routes"
	"github.com/gin-gonic/gin"
)

func ServerInit(plugins []plugins.Plugin, wg *sync.WaitGroup) {
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
		err := plugin.PluginInit(r, DB, &PluginAPI, wg)
		if err != nil {
			logger.Log(err, "Ошибка при инициализации плагина", logger.Error)
		}
	}
	srv := &http.Server{
		Addr:    ":3000",
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log(err, "Ошибка при запуске сервера", logger.Error)
		}
	}()
	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Log(nil, "Получен сигнал завершения, выполняется graceful shutdown", logger.Info)

	// Контекст с таймаутом для завершения текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log(err, "Ошибка при остановке сервера", logger.Error)
	} else {
		logger.Log(nil, "Сервер корректно остановлен", logger.Info)
	}

	// Закрыть sql соединения
	if err := sqlDB.Close(); err != nil {
		logger.Log(err, "Ошибка при закрытии sqlDB", logger.Error)
	} else {
		logger.Log(nil, "sqlDB закрыт", logger.Info)
	}

	// Ожидание завершения воркеров/плагинов, если они используют тот же WaitGroup
	if wg != nil {
		wg.Wait()
	}
}
