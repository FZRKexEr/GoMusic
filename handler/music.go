package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"sync/atomic"

	"GoMusic/logic"
	"GoMusic/misc/models"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

const (
	unsupportedLinkMessage = "不支持的音乐链接格式"

	formatSongSinger = "song-singer"
	formatSingerSong = "singer-song"
	formatSongOnly   = "song"
	orderReverse     = "reverse"
	songSeparator    = " - "
)

var (
	counter atomic.Int64

	discoverQiShuiMusic = logic.QiShuiMusicDiscover
)

type songListRequest struct {
	Link   string
	Clean  bool
	Format string
	Order  string
}

type songListJSONRequest struct {
	URL      string `json:"url"`
	Format   string `json:"format"`
	Clean    *bool  `json:"clean"`
	Detailed *bool  `json:"detailed"`
	Reverse  bool   `json:"reverse"`
	Order    string `json:"order"`
}

// MusicHandler 处理音乐请求的入口函数
func MusicHandler(_ context.Context, c *app.RequestContext) {
	request, err := parseSongListRequest(c)
	if err != nil {
		c.JSON(consts.StatusBadRequest, models.BadRequest(err.Error()))
		return
	}

	detailed := !request.Clean
	currentCount := counter.Add(1)

	slog.Info("歌单请求", "count", currentCount, "link", request.Link, "clean", request.Clean, "format", request.Format, "order", request.Order)

	// 路由到不同的音乐服务处理函数
	switch {
	case logic.IsQiShuiMusicLink(request.Link):
		handleQiShuiMusic(c, request.Link, detailed, request.Format, request.Order)
	default:
		slog.Warn("不支持的音乐链接格式", "link", request.Link)
		c.JSON(consts.StatusBadRequest, models.BadRequest(unsupportedLinkMessage))
	}
}

func parseSongListRequest(c *app.RequestContext) (songListRequest, error) {
	if strings.Contains(string(c.Request.Header.ContentType()), "application/json") {
		return parseSongListJSONRequest(c.Request.Body())
	}

	return songListRequest{
		Link:   strings.TrimSpace(c.PostForm("url")),
		Clean:  c.Query("detailed") != "true",
		Format: c.Query("format"),
		Order:  c.Query("order"),
	}, nil
}

func parseSongListJSONRequest(body []byte) (songListRequest, error) {
	var payload songListJSONRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		return songListRequest{}, errors.New("请求 JSON 格式错误")
	}

	clean := true
	if payload.Detailed != nil {
		clean = !*payload.Detailed
	}
	if payload.Clean != nil {
		clean = *payload.Clean
	}

	order := payload.Order
	if payload.Reverse {
		order = orderReverse
	}

	return songListRequest{
		Link:   strings.TrimSpace(payload.URL),
		Clean:  clean,
		Format: payload.Format,
		Order:  order,
	}, nil
}

// handleQiShuiMusic 处理汽水音乐歌单
func handleQiShuiMusic(c *app.RequestContext, link string, detailed bool, format, order string) {
	songList, err := discoverQiShuiMusic(link, detailed)
	if err != nil {
		slog.Error("获取汽水音乐歌单失败", "err", err)
		c.JSON(consts.StatusBadRequest, models.BadRequest(err.Error()))
		return
	}

	// 根据格式选项处理歌曲列表
	formatSongList(songList, format)

	// 根据顺序选项处理歌曲列表
	processSongOrder(songList, order)

	c.JSON(consts.StatusOK, models.OK(songList))
}

// processSongOrder 根据指定的顺序处理歌曲列表
func processSongOrder(songList *models.SongList, order string) {
	if songList == nil || len(songList.Songs) == 0 {
		return
	}

	if order == orderReverse {
		slices.Reverse(songList.Songs)
	}
}

// formatSongList 根据指定的格式处理歌曲列表
func formatSongList(songList *models.SongList, format string) {
	if songList == nil || len(songList.Songs) == 0 {
		return
	}

	if format == "" || format == formatSongSinger {
		return
	}

	formattedSongs := make([]string, 0, len(songList.Songs))

	for _, song := range songList.Songs {
		switch format {
		case formatSingerSong:
			formattedSongs = append(formattedSongs, formatSingerSongValue(song))
		case formatSongOnly:
			formattedSongs = append(formattedSongs, formatSongOnlyValue(song))
		default:
			formattedSongs = append(formattedSongs, song)
		}
	}

	songList.Songs = formattedSongs
}

func formatSingerSongValue(song string) string {
	title, artist, ok := splitFormattedSong(song)
	if !ok || artist == "" {
		return song
	}
	return artist + songSeparator + title
}

func formatSongOnlyValue(song string) string {
	title, _, ok := splitFormattedSong(song)
	if !ok {
		return song
	}
	return title
}

func splitFormattedSong(song string) (title, artist string, ok bool) {
	idx := strings.LastIndex(song, songSeparator)
	if idx < 0 {
		return song, "", false
	}
	title = strings.TrimSpace(song[:idx])
	artist = strings.TrimSpace(song[idx+len(songSeparator):])
	return title, artist, title != ""
}
