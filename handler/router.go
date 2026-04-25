package handler

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
)

const defaultAddr = ":8081"

func NewRouter() *server.Hertz {
	router := server.Default(server.WithHostPorts(defaultAddr))
	router.Use(cors.Default()) // allow all origins
	router.POST("/songlist", MusicHandler)
	return router
}
