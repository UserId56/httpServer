package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/UserId56/httpServer/core/database"
	"github.com/UserId56/httpServer/core/logger"
	"github.com/UserId56/httpServer/core/models"
	"github.com/UserId56/httpServer/core/plugins"
	"github.com/UserId56/httpServer/core/routes"
	"github.com/UserId56/httpServer/core/services"
	"github.com/gin-gonic/gin"
)

func ServerInit(plugins []plugins.Plugin, wg *sync.WaitGroup, conf models.Config) {
	services.RegisterValidators()

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
	conf.Init()
	fmt.Printf("Сервер запущен c параметрами: %+v\n", conf)
	sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
	sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
	// 1. Ограничиваем общую жизнь соединения.
	// Даже если всё идеально, раз в 30 минут полезно создать чистое соединение.
	sqlDB.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Minute)

	// 2. Ограничиваем время простоя.
	// Если соединение просто лежит в пуле 5 минут без дела — закрываем его.
	// Это освободит память (RAM) на стороне Postgres-процессов.
	sqlDB.SetConnMaxIdleTime(time.Duration(conf.ConnMaxIdleTime) * time.Minute)
	r := gin.Default()
	routes.SetupRouter(r, DB)
	PluginAPI := make(map[string]interface{})
	pluginCtx, pluginCancel := context.WithCancel(context.Background())
	for _, plugin := range plugins {
		err := plugin.PluginInit(r, DB, &PluginAPI, wg, pluginCtx)
		if err != nil {
			logger.Log(err, "Ошибка при инициализации плагина", logger.Error)
		}
	}
	addr := fmt.Sprintf("%s:%d", conf.Address, conf.Port)
	srv := &http.Server{
		Addr:    addr,
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
	pluginCancel()
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
