package httputil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHTTPUtil(t *testing.T) {
	PatchConvey("HTTP utilities", t, func() {
		PatchConvey("Get sends the default user agent and returns the response", func() {
			var method, userAgent string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				method = r.Method
				userAgent = r.UserAgent()
				_, _ = w.Write([]byte("ok"))
			}))
			defer server.Close()

			resp, err := Get(server.URL)
			So(err, ShouldBeNil)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(method, ShouldEqual, http.MethodGet)
			So(userAgent, ShouldEqual, defaultUserAgent)
			So(string(body), ShouldEqual, "ok")
		})

		PatchConvey("GetRedirectLocation returns Location without following redirects", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/next", http.StatusFound)
			}))
			defer server.Close()

			location, err := GetRedirectLocation(server.URL)

			So(err, ShouldBeNil)
			So(location, ShouldEqual, "/next")
		})

		PatchConvey("invalid URLs return an error", func() {
			resp, err := Get("://bad-url")

			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})
	})
}
