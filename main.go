package main

import (
	"httpServer/plugins"
	"sync"

	"github.com/UserId56/httpServer/core"
	"github.com/UserId56/httpServer/core/models"
	plugins2 "github.com/UserId56/httpServer/core/plugins"
)

func main() {
	testPlugin := plugins.PluginTest{
		Name: "Test Plugin",
		Path: "/plugin/test",
	}
	wg := &sync.WaitGroup{}
	InitPlugins := []plugins2.Plugin{&testPlugin}
	config := models.Config{
		Address:         "0.0.0.0",
		Port:            3000,
		MaxIdleConns:    100,  // держать готовые соединения
		MaxOpenConns:    240,  // максимальные параллельные подключения к БД
		ConnMaxLifetime: 1800, // в секундах (30 минут) — перезапускать соединения периодически
		ConnMaxIdleTime: 300,  // в секундах (5 минут)
	}
	core.ServerInit(InitPlugins, wg, config)
}
