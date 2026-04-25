package utils

import (
	"testing"

	. "github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStandardSongName(t *testing.T) {
	PatchConvey("standard song name", t, func() {
		PatchConvey("normalizes Chinese brackets", func() {
			So(StandardSongName("歌曲（Live）"), ShouldEqual, "歌曲 (Live)")
		})

		PatchConvey("removes bracketed metadata without crossing groups", func() {
			So(StandardSongName("歌曲【伴奏】 remix【现场】"), ShouldEqual, "歌曲 remix")
		})

		PatchConvey("trims surrounding whitespace after cleanup", func() {
			So(StandardSongName(" 歌曲【伴奏】 "), ShouldEqual, "歌曲")
		})
	})
}
