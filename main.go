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
		Address:         "",
		Port:            3001,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 30,
		ConnMaxIdleTime: 10,
	}
	core.ServerInit(InitPlugins, wg, config)
}
