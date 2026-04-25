package logic

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	. "github.com/bytedance/mockey"
	. "github.com/smartystreets/goconvey/convey"
)

type fakeQishuiClient struct {
	getFn      func(link string) (*http.Response, error)
	redirectFn func(link string) (string, error)
}

func (client fakeQishuiClient) Get(link string) (*http.Response, error) {
	return client.getFn(link)
}

func (client fakeQishuiClient) GetRedirectLocation(link string) (string, error) {
	return client.redirectFn(link)
}

func TestDiscoverQiShuiMusic(t *testing.T) {
	PatchConvey("discovering a Qishui playlist", t, func() {
		PatchConvey("resolves a share URL and parses SSR data without real network", func() {
			finalURL := "https://music.douyin.com/qishui/share/playlist?playlist_id=playlist_test"
			client := fakeQishuiClient{
				redirectFn: func(link string) (string, error) {
					So(link, ShouldEqual, "https://qishui.douyin.com/s/testToken/")
					return finalURL, nil
				},
				getFn: func(link string) (*http.Response, error) {
					So(link, ShouldEqual, finalURL)
					return htmlResponse(http.StatusOK, routerDataHTML()), nil
				},
			}

			songList, err := discoverQiShuiMusic("歌单 https://qishui.douyin.com/s/testToken/ @汽水音乐", false, client)

			So(err, ShouldBeNil)
			So(songList.Name, ShouldEqual, "Test Playlist-Test Owner")
			So(songList.SongsCount, ShouldEqual, 2)
			So(songList.Songs, ShouldResemble, []string{
				"Song A (Live) - Singer A, Singer B",
				"Video A - Creator A",
			})
		})

		PatchConvey("returns an error for non-OK playlist responses", func() {
			client := fakeQishuiClient{
				redirectFn: func(link string) (string, error) { return "", nil },
				getFn: func(link string) (*http.Response, error) {
					return htmlResponse(http.StatusForbidden, ""), nil
				},
			}

			songList, err := discoverQiShuiMusic("https://music.douyin.com/qishui/share/playlist?playlist_id=playlist_test", true, client)

			So(songList, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "403")
		})

		PatchConvey("propagates redirect errors", func() {
			client := fakeQishuiClient{
				redirectFn: func(link string) (string, error) {
					return "", errors.New("redirect failed")
				},
				getFn: func(link string) (*http.Response, error) {
					return nil, errors.New("should not fetch")
				},
			}

			songList, err := discoverQiShuiMusic("https://qishui.douyin.com/s/testToken/", true, client)

			So(songList, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "redirect failed")
		})
	})
}

func TestQiShuiURLHelpers(t *testing.T) {
	PatchConvey("Qishui URL helpers", t, func() {
		PatchConvey("extracts share URLs from mixed text without real tokens", func() {
			input := `歌单｜测试歌单 https://qishui.douyin.com/s/testToken/ @汽水音乐`

			So(extractURL(input), ShouldEqual, "https://qishui.douyin.com/s/testToken/")
		})

		PatchConvey("recognizes only supported Qishui share pages", func() {
			So(IsQiShuiMusicLink("https://qishui.douyin.com/s/testToken/"), ShouldBeTrue)
			So(IsQiShuiMusicLink("https://music.douyin.com/qishui/share/playlist?playlist_id=playlist_test"), ShouldBeTrue)
			So(IsQiShuiMusicLink("https://www.douyin.com/video/test_video"), ShouldBeFalse)
			So(IsQiShuiMusicLink("not a url"), ShouldBeFalse)
		})

		PatchConvey("resolves relative redirect locations", func() {
			resolvedURL, err := resolveRedirectURL("https://qishui.douyin.com/s/testToken/", "/qishui/share/playlist?playlist_id=playlist_test")

			So(err, ShouldBeNil)
			So(resolvedURL, ShouldEqual, "https://qishui.douyin.com/qishui/share/playlist?playlist_id=playlist_test")
		})
	})
}

func TestParseQsSongList(t *testing.T) {
	PatchConvey("parsing Qishui song lists", t, func() {
		PatchConvey("prefers router data over DOM selectors", func() {
			songList, err := parseQsSongList(strings.NewReader(routerDataHTML()), true)

			So(err, ShouldBeNil)
			So(songList.Name, ShouldEqual, "Test Playlist-Test Owner")
			So(songList.Songs, ShouldResemble, []string{
				"Song A（Live）【Official】 - Singer A, Singer B",
				"Video A - Creator A",
			})
		})

		PatchConvey("falls back to DOM parsing when router data is absent", func() {
			songList, err := parseQsSongList(strings.NewReader(domOnlyHTML()), false)

			So(err, ShouldBeNil)
			So(songList.Name, ShouldEqual, "DOM Playlist-DOM Owner")
			So(songList.Songs, ShouldResemble, []string{"DOM Song (Live) - DOM Singer"})
		})

		PatchConvey("reports an error when no songs can be parsed", func() {
			songList, err := parseQsSongList(strings.NewReader("<html></html>"), true)

			So(songList, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "未解析到汽水音乐歌曲")
		})
	})
}

func htmlResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func routerDataHTML() string {
	return `<html><body><script>_ROUTER_DATA = {"loaderData":{"playlist_page":{"medias":[{"type":"track","entity":{"track":{"name":"Song A（Live）【Official】","artists":[{"name":"Singer A"},{"simple_display_name":"Singer B"}]}}},{"type":"video","entity":{"video":{"title":"Video A","artists":[{"user_info":{"nickname":"Creator A"}}]}}}],"playlistInfo":{"title":"Test Playlist","owner":{"nickname":"Test Owner"}}}}};
function runWindowFn(){}</script></body></html>`
}

func domOnlyHTML() string {
	return `<html><body><div id="root"><div><div><div><div>
		<div>
			<div></div><div></div>
			<div>
				<h1><p>DOM Playlist</p></h1>
				<div><div><div></div><div><p>DOM Owner</p></div></div></div>
			</div>
		</div>
		<div><div><div><div><div>
			<div>
				<div>1</div>
				<div>
					<div><p>DOM Song（Live）</p></div>
					<div><p>DOM Singer • DOM Album</p></div>
				</div>
			</div>
		</div></div></div></div></div>
	</div></div></div></div></div></body></html>`
}
