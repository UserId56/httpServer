package main

import (
	"httpServer/plugins"
	"sync"

	"github.com/UserId56/httpServer/core"
	plugins2 "github.com/UserId56/httpServer/core/plugins"
)

func main() {
	testPlugin := plugins.PluginTest{
		Name: "Test Plugin",
		Path: "/plugin/test",
	}
	wg := &sync.WaitGroup{}
	InitPlugins := []plugins2.Plugin{&testPlugin}
	core.ServerInit(InitPlugins, wg)
}
