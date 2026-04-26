package handler

import (
	"context"

	frontend "GoMusic/static"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/cors"
)

const defaultAddr = ":8081"

func NewRouter() *server.Hertz {
	router := server.Default(server.WithHostPorts(defaultAddr))
	router.Use(cors.Default()) // allow all origins
	registerFrontend(router)
	router.POST("/songlist", MusicHandler)
	return router
}

func registerFrontend(router *server.Hertz) {
	router.GET("/", serveStaticAsset("index.html"))
	router.GET("/index.html", serveStaticAsset("index.html"))
	router.GET("/styles.css", serveStaticAsset("styles.css"))
	router.GET("/app.js", serveStaticAsset("app.js"))
}

func serveStaticAsset(name string) app.HandlerFunc {
	return func(_ context.Context, c *app.RequestContext) {
		asset, err := frontend.Read(name)
		if err != nil {
			c.String(consts.StatusNotFound, "not found")
			return
		}

		c.Header("Cache-Control", "no-store")
		c.Data(consts.StatusOK, asset.ContentType, asset.Content)
	}
}
