package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"testing"

	"GoMusic/misc/models"

	. "github.com/bytedance/mockey"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMusicHandler(t *testing.T) {
	PatchConvey("music handler", t, func() {
		PatchConvey("returns a parsed Qishui playlist", func() {
			MockValue(&discoverQiShuiMusic).To(func(link string, detailed bool) (*models.SongList, error) {
				So(link, ShouldEqual, "https://qishui.douyin.com/s/testToken/")
				So(detailed, ShouldBeTrue)
				return &models.SongList{
					Name:       "Test Playlist",
					Songs:      []string{"Song A - Singer A", "Song B - Singer B"},
					SongsCount: 2,
				}, nil
			})

			statusCode, result := performSongListRequest("/songlist?detailed=true&format=singer-song&order=reverse", url.Values{
				"url": {"https://qishui.douyin.com/s/testToken/"},
			}.Encode())

			So(statusCode, ShouldEqual, consts.StatusOK)
			So(result.Code, ShouldEqual, models.ResultCodeOK)
			So(result.Msg, ShouldEqual, "success")
			So(result.Data.Name, ShouldEqual, "Test Playlist")
			So(result.Data.Songs, ShouldResemble, []string{"Singer B - Song B", "Singer A - Song A"})
			So(result.Data.SongsCount, ShouldEqual, 2)
		})

		PatchConvey("rejects unsupported links before discovery", func() {
			called := false
			MockValue(&discoverQiShuiMusic).To(func(link string, detailed bool) (*models.SongList, error) {
				called = true
				return nil, nil
			})

			statusCode, result := performSongListRequest("/songlist", url.Values{
				"url": {"https://example.com/playlist"},
			}.Encode())

			So(called, ShouldBeFalse)
			So(statusCode, ShouldEqual, consts.StatusBadRequest)
			So(result.Code, ShouldEqual, models.ResultCodeBadRequest)
			So(result.Msg, ShouldEqual, unsupportedLinkMessage)
		})

		PatchConvey("returns discovery errors as bad requests", func() {
			MockValue(&discoverQiShuiMusic).To(func(link string, detailed bool) (*models.SongList, error) {
				return nil, errors.New("discover failed")
			})

			statusCode, result := performSongListRequest("/songlist", url.Values{
				"url": {"https://qishui.douyin.com/s/testToken/"},
			}.Encode())

			So(statusCode, ShouldEqual, consts.StatusBadRequest)
			So(result.Code, ShouldEqual, models.ResultCodeBadRequest)
			So(result.Msg, ShouldEqual, "discover failed")
		})
	})
}

func TestSongListFormatting(t *testing.T) {
	PatchConvey("song list formatting", t, func() {
		PatchConvey("reverses song order in place", func() {
			songList := &models.SongList{Songs: []string{"A - Singer A", "B - Singer B"}}

			processSongOrder(songList, orderReverse)

			So(songList.Songs, ShouldResemble, []string{"B - Singer B", "A - Singer A"})
		})

		PatchConvey("formats singer-song using the last separator", func() {
			songList := &models.SongList{Songs: []string{"everyday - live in a van - Ryce, The Guest Room"}}

			formatSongList(songList, formatSingerSong)

			So(songList.Songs, ShouldResemble, []string{"Ryce, The Guest Room - everyday - live in a van"})
		})

		PatchConvey("formats song-only using the last separator", func() {
			songList := &models.SongList{Songs: []string{"everyday - live in a van - Ryce, The Guest Room"}}

			formatSongList(songList, formatSongOnly)

			So(songList.Songs, ShouldResemble, []string{"everyday - live in a van"})
		})

		PatchConvey("keeps unknown formats unchanged", func() {
			songList := &models.SongList{Songs: []string{"Song A - Singer A"}}

			formatSongList(songList, "unknown")

			So(songList.Songs, ShouldResemble, []string{"Song A - Singer A"})
		})
	})
}

type songListHTTPResponse struct {
	Code models.ResultCode `json:"code"`
	Msg  string            `json:"msg"`
	Data models.SongList   `json:"data"`
}

func performSongListRequest(uri, body string) (int, songListHTTPResponse) {
	router := NewRouter()
	ctx := router.NewContext()
	ctx.Request.SetRequestURI(uri)
	ctx.Request.Header.SetMethod(consts.MethodPost)
	ctx.Request.Header.SetContentTypeBytes([]byte("application/x-www-form-urlencoded"))
	ctx.Request.SetBodyString(body)

	router.ServeHTTP(context.Background(), ctx)

	var result songListHTTPResponse
	_ = json.Unmarshal(ctx.Response.Body(), &result)
	return ctx.Response.StatusCode(), result
}
