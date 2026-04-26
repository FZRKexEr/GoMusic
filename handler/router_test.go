package handler

import (
	"context"
	"testing"

	. "github.com/bytedance/mockey"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFrontendRoutes(t *testing.T) {
	PatchConvey("embedded frontend routes", t, func() {
		PatchConvey("serves root and static assets without disk files", func() {
			router := NewRouter()

			status, body, cacheControl := performGET(router, "/")
			So(status, ShouldEqual, consts.StatusOK)
			So(body, ShouldContainSubstring, "<title>GoMusic</title>")
			So(cacheControl, ShouldEqual, "no-store")

			status, body, _ = performGET(router, "/app.js?v=test")
			So(status, ShouldEqual, consts.StatusOK)
			So(body, ShouldContainSubstring, "requestSongList")
		})

		PatchConvey("returns not found for missing embedded assets", func() {
			router := server.Default()
			router.GET("/missing", serveStaticAsset("missing.txt"))

			status, body, _ := performGET(router, "/missing")

			So(status, ShouldEqual, consts.StatusNotFound)
			So(body, ShouldEqual, "not found")
		})
	})
}

func performGET(router *server.Hertz, uri string) (int, string, string) {
	ctx := router.NewContext()
	ctx.Request.SetRequestURI(uri)
	ctx.Request.Header.SetMethod(consts.MethodGet)

	router.ServeHTTP(context.Background(), ctx)

	return ctx.Response.StatusCode(), string(ctx.Response.Body()), string(ctx.Response.Header.Peek("Cache-Control"))
}
