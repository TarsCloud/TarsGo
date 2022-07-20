package main

import (
	"github.com/TarsCloud/TarsGo/tars"
	"github.com/gin-gonic/gin"
)

func main() {
	mux := &tars.TarsHttpMux{}
	r := mux.GetGinEngine()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Get server config
	cfg := tars.GetServerConfig()
	tars.AddHttpServant(mux, cfg.App+"."+cfg.Server+".HttpObj")
	tars.Run()
}
