package handler

import (
	"fmt"

	"GoMusic/misc/models"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
)

func NewRouter() *server.Hertz {
	router := server.Default(server.WithHostPorts(fmt.Sprintf(models.PortFormat, models.Port)))
	router.Use(cors.Default())     // allow all origins
	router.Static("/", "./static") // load static files
	router.POST("/songlist", MusicHandler)
	return router
}
