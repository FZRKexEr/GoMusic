package logic

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"GoMusic/misc/httputil"

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
		PatchConvey("uses the default HTTP client through the exported entrypoint", func() {
			finalURL := "https://music.douyin.com/qishui/share/playlist?playlist_id=playlist_test"
			Mock(httputil.GetRedirectLocation).To(func(link string) (string, error) {
				So(link, ShouldEqual, "https://qishui.douyin.com/s/testToken/")
				return finalURL, nil
			}).Build()
			Mock(httputil.Get).To(func(link string) (*http.Response, error) {
				So(link, ShouldEqual, finalURL)
				return htmlResponse(http.StatusOK, routerDataHTML()), nil
			}).Build()

			songList, err := QiShuiMusicDiscover("https://qishui.douyin.com/s/testToken/", false)

			So(err, ShouldBeNil)
			So(songList.SongsCount, ShouldEqual, 2)
		})

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

		PatchConvey("propagates playlist fetch errors", func() {
			client := fakeQishuiClient{
				redirectFn: func(link string) (string, error) { return "", nil },
				getFn: func(link string) (*http.Response, error) {
					return nil, errors.New("fetch failed")
				},
			}

			songList, err := discoverQiShuiMusic("https://music.douyin.com/qishui/share/playlist?playlist_id=playlist_test", true, client)

			So(songList, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "fetch failed")
		})

		PatchConvey("propagates parse errors", func() {
			client := fakeQishuiClient{
				redirectFn: func(link string) (string, error) { return "", nil },
				getFn: func(link string) (*http.Response, error) {
					return htmlResponse(http.StatusOK, "<html></html>"), nil
				},
			}

			songList, err := discoverQiShuiMusic("https://music.douyin.com/qishui/share/playlist?playlist_id=playlist_test", true, client)

			So(songList, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "未解析到汽水音乐歌曲")
		})
	})
}

func TestQiShuiURLHelpers(t *testing.T) {
	PatchConvey("Qishui URL helpers", t, func() {
		PatchConvey("extracts share URLs from mixed text without real tokens", func() {
			input := `歌单｜测试歌单 https://qishui.douyin.com/s/testToken/ @汽水音乐`

			So(extractURL(input), ShouldEqual, "https://qishui.douyin.com/s/testToken/")
			So(extractURL(`<https://qishui.douyin.com/s/testToken/>，`), ShouldEqual, "https://qishui.douyin.com/s/testToken/")
		})

		PatchConvey("recognizes only supported Qishui share pages", func() {
			So(IsQiShuiMusicLink("https://qishui.douyin.com/s/testToken/"), ShouldBeTrue)
			So(IsQiShuiMusicLink("https://music.douyin.com/qishui/share/playlist?playlist_id=playlist_test"), ShouldBeTrue)
			So(IsQiShuiMusicLink("https://music.douyin.com/not-qishui/share"), ShouldBeFalse)
			So(IsQiShuiMusicLink("https://www.douyin.com/video/test_video"), ShouldBeFalse)
			So(IsQiShuiMusicLink("http://%zz"), ShouldBeFalse)
			So(IsQiShuiMusicLink("not a url"), ShouldBeFalse)
		})

		PatchConvey("resolves relative redirect locations", func() {
			resolvedURL, err := resolveRedirectURL("https://qishui.douyin.com/s/testToken/", "/qishui/share/playlist?playlist_id=playlist_test")

			So(err, ShouldBeNil)
			So(resolvedURL, ShouldEqual, "https://qishui.douyin.com/qishui/share/playlist?playlist_id=playlist_test")
		})

		PatchConvey("reports invalid redirect URLs", func() {
			resolvedURL, err := resolveRedirectURL("https://qishui.douyin.com/s/testToken/", "http://%zz")
			So(resolvedURL, ShouldBeEmpty)
			So(err, ShouldNotBeNil)

			resolvedURL, err = resolveRedirectURL("http://%zz", "/qishui/share/playlist")
			So(resolvedURL, ShouldBeEmpty)
			So(err, ShouldNotBeNil)
		})

		PatchConvey("validates Qishui playlist URL resolution", func() {
			client := fakeQishuiClient{
				redirectFn: func(link string) (string, error) { return "", nil },
				getFn:      func(link string) (*http.Response, error) { return nil, nil },
			}

			resolvedURL, err := resolveQiShuiURL("", client)
			So(resolvedURL, ShouldBeEmpty)
			So(err, ShouldNotBeNil)

			resolvedURL, err = resolveQiShuiURL("https://example.com/list", client)
			So(resolvedURL, ShouldBeEmpty)
			So(err, ShouldNotBeNil)

			resolvedURL, err = resolveQiShuiURL("https://qishui.douyin.com/s/testToken/", client)
			So(err, ShouldBeNil)
			So(resolvedURL, ShouldEqual, "https://qishui.douyin.com/s/testToken/")
		})

		PatchConvey("rejects redirects outside Qishui", func() {
			client := fakeQishuiClient{
				redirectFn: func(link string) (string, error) { return "https://example.com/playlist", nil },
				getFn:      func(link string) (*http.Response, error) { return nil, nil },
			}

			resolvedURL, err := resolveQiShuiURL("https://qishui.douyin.com/s/testToken/", client)

			So(resolvedURL, ShouldBeEmpty)
			So(err, ShouldNotBeNil)
		})

		PatchConvey("detects playlist IDs conservatively", func() {
			So(hasPlaylistID("https://music.douyin.com/qishui/share/playlist?playlist_id=playlist_test"), ShouldBeTrue)
			So(hasPlaylistID("https://music.douyin.com/qishui/share/playlist"), ShouldBeFalse)
			So(hasPlaylistID("http://%zz"), ShouldBeFalse)
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

		PatchConvey("returns read errors", func() {
			songList, err := parseQsSongList(errorReader{}, true)

			So(songList, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "read failed")
		})

		PatchConvey("ignores invalid router data and falls back", func() {
			songList, ok := parseRouterDataSongList([]byte(`<script>_ROUTER_DATA = {bad json};</script>`), true)

			So(songList, ShouldBeNil)
			So(ok, ShouldBeFalse)
		})
	})
}

func TestQishuiFormattingHelpers(t *testing.T) {
	PatchConvey("Qishui formatting helpers", t, func() {
		PatchConvey("formats playlist names", func() {
			So(formatPlaylistName("", "Owner"), ShouldEqual, "Owner")
			So(formatPlaylistName("Playlist", ""), ShouldEqual, "Playlist")
			So(formatPlaylistName("Playlist", "Owner"), ShouldEqual, "Playlist-Owner")
		})

		PatchConvey("formats songs with optional artists", func() {
			So(formatSong("", "Artist"), ShouldBeEmpty)
			So(formatSong("Song", ""), ShouldEqual, "Song")
			So(formatSong(" Song ", " Artist "), ShouldEqual, "Song - Artist")
		})
	})
}

func TestQishuiRouterDataHelpers(t *testing.T) {
	PatchConvey("Qishui router data helpers", t, func() {
		PatchConvey("uses fallback video fields and public owner names", func() {
			data := qishuiRouterData{}
			page := &data.LoaderData.PlaylistPage
			page.PlaylistInfo.Title = "Fallback Playlist"
			page.PlaylistInfo.Owner.PublicName = "Public Owner"
			page.Medias = []qishuiMedia{
				{
					Entity: struct {
						Track *qishuiTrack `json:"track"`
						Video *qishuiVideo `json:"video"`
					}{
						Video: &qishuiVideo{
							Description: "Video Description【tag】",
							Artists: []qishuiVideoArtist{
								{Name: "Video Artist"},
								{UserInfo: struct {
									Nickname   string `json:"nickname"`
									PublicName string `json:"public_name"`
								}{PublicName: "Public Creator"}},
							},
						},
					},
				},
			}

			songList, err := data.toSongList(false)

			So(err, ShouldBeNil)
			So(songList.Name, ShouldEqual, "Fallback Playlist-Public Owner")
			So(songList.Songs, ShouldResemble, []string{"Video Description - Video Artist, Public Creator"})
		})

		PatchConvey("returns errors for empty router data", func() {
			data := qishuiRouterData{}

			songList, err := data.toSongList(true)

			So(songList, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		PatchConvey("handles media without known entities", func() {
			title, artist := qishuiMedia{}.songInfo()

			So(title, ShouldBeEmpty)
			So(artist, ShouldBeEmpty)
		})

		PatchConvey("skips blank helper values", func() {
			So(joinNonEmpty([]string{"", " A ", "B"}, ", "), ShouldEqual, "A, B")
			So(firstNonEmpty("", "  ", "Name"), ShouldEqual, "Name")
			So(firstNonEmpty("", "  "), ShouldBeEmpty)
		})
	})
}

type errorReader struct{}

func (errorReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read failed")
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
