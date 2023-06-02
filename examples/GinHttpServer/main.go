package main

import (
	cgin "github.com/TarsCloud/TarsGo/contrib/gin"
	"github.com/TarsCloud/TarsGo/tars"
	"github.com/gin-gonic/gin"
)

func main() {
	g := cgin.New()
	g.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Get server config
	cfg := tars.GetServerConfig()
	tars.AddHttpServant(g, cfg.App+"."+cfg.Server+".HttpObj")
	tars.Run()
}
