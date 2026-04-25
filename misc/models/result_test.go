package models

import (
	"testing"

	. "github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
)

func TestResultConstructors(t *testing.T) {
	PatchConvey("result constructors", t, func() {
		PatchConvey("OK builds a success response", func() {
			data := map[string]string{"name": "test"}

			result := OK(data)

			So(result.Code, ShouldEqual, ResultCodeOK)
			So(result.Msg, ShouldEqual, "success")
			So(result.Data, ShouldResemble, data)
		})

		PatchConvey("BadRequest builds a bad request response", func() {
			result := BadRequest("bad input")

			So(result.Code, ShouldEqual, ResultCodeBadRequest)
			So(result.Msg, ShouldEqual, "bad input")
			So(result.Data, ShouldBeNil)
		})
	})
}
