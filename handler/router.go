package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
)

const defaultAddr = ":8081"

func NewRouter() *server.Hertz {
	router := server.Default(server.WithHostPorts(defaultAddr))
	router.Use(cors.Default()) // allow all origins
	router.GET("/", func(_ context.Context, c *app.RequestContext) {
		c.File("./static/index.html")
	})
	router.Static("/", "./static")
	router.POST("/songlist", MusicHandler)
	return router
}
