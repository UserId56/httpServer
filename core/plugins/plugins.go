package plugins

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Plugin interface {
	PluginInit(engine *gin.Engine, db *gorm.DB, pluginsData *map[string]interface{}) error
}
