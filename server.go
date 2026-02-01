package httpserver

import (
	"sync"

	"github.com/UserId56/httpServer/core"
	"github.com/UserId56/httpServer/core/models"
	coreplugins "github.com/UserId56/httpServer/core/plugins"
)

// Реэкспорт типов для удобства импорта из внешних проектов.
type Plugin = coreplugins.Plugin
type Config = models.Config

// ServerInit - публичная точка входа, просто вызывает core.ServerInit.
// Позволяет импортировать github.com/UserId56/httpServer и вызывать
// httpserver.ServerInit(...) из других проектов.
func ServerInit(plugins []Plugin, wg *sync.WaitGroup, cfg Config) {
	core.ServerInit(plugins, wg, cfg)
}
