package handler

import (
	"context"
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

// MusicHandler 处理音乐请求的入口函数
func MusicHandler(_ context.Context, c *app.RequestContext) {
	link := c.PostForm("url")
	detailed := c.Query("detailed") == "true"
	format := c.Query("format")
	order := c.Query("order")
	currentCount := counter.Add(1)

	slog.Info("歌单请求", "count", currentCount, "link", link, "detailed", detailed, "format", format, "order", order)

	// 路由到不同的音乐服务处理函数
	switch {
	case logic.IsQiShuiMusicLink(link):
		handleQiShuiMusic(c, link, detailed, format, order)
	default:
		slog.Warn("不支持的音乐链接格式", "link", link)
		c.JSON(consts.StatusBadRequest, models.BadRequest(unsupportedLinkMessage))
	}
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
