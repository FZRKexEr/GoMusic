package main

import (
	"context"
	"testing"

	"GoMusic/handler"

	. "github.com/bytedance/mockey"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStaticFrontend(t *testing.T) {
	PatchConvey("static frontend", t, func() {
		router := handler.NewRouter()
		ctx := router.NewContext()
		ctx.Request.SetRequestURI("/")
		ctx.Request.Header.SetMethod(consts.MethodGet)

		router.ServeHTTP(context.Background(), ctx)

		So(ctx.Response.StatusCode(), ShouldEqual, consts.StatusOK)
		So(string(ctx.Response.Body()), ShouldContainSubstring, "<title>GoMusic</title>")
	})
}
