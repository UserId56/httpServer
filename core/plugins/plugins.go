package plugins

import (
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Plugin interface {
	PluginInit(engine *gin.Engine, db *gorm.DB, pluginsData *map[string]interface{}, group *sync.WaitGroup) error
}
