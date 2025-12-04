package plugins

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PluginTest struct {
	Name string
	Path string
}

func (p *PluginTest) testHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Plugin Test Works!",
	})
}

func (p *PluginTest) PluginInit(r *gin.Engine, db *gorm.DB, pluginsData *map[string]interface{}) error {
	r.GET(p.Path, p.testHandler)
	return nil
}
