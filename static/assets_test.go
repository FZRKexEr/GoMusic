package static

import (
	"testing"

	. "github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRead(t *testing.T) {
	PatchConvey("embedded static assets", t, func() {
		PatchConvey("serves the default index", func() {
			asset, err := Read("")

			So(err, ShouldBeNil)
			So(asset.ContentType, ShouldContainSubstring, "text/html")
			So(string(asset.Content), ShouldContainSubstring, "<title>GoMusic</title>")
		})

		PatchConvey("serves named assets with content types", func() {
			css, err := Read("styles.css")
			So(err, ShouldBeNil)
			So(css.ContentType, ShouldContainSubstring, "text/css")
			So(string(css.Content), ShouldContainSubstring, ".song-list")

			js, err := Read("/app.js")
			So(err, ShouldBeNil)
			So(js.ContentType, ShouldNotBeEmpty)
			So(string(js.Content), ShouldContainSubstring, "requestSongList")
		})

		PatchConvey("rejects unknown assets", func() {
			asset, err := Read("../README.md")

			So(asset.Content, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}
