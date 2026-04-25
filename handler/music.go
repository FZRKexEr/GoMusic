package handler

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
	"sync/atomic"

	"GoMusic/logic"
	"GoMusic/misc/models"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

const (
	qishuiMusic = `(qishui)|(douyin)`
	SUCCESS     = "success"
)

var (
	qishuiMusicRegx, _ = regexp.Compile(qishuiMusic)
	counter            atomic.Int64 // request counter
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
	case qishuiMusicRegx.MatchString(link):
		handleQiShuiMusic(c, link, detailed, format, order)
	default:
		slog.Warn("不支持的音乐链接格式", "link", link)
		c.JSON(consts.StatusBadRequest, &models.Result{Code: models.FailureCode, Msg: "不支持的音乐链接格式", Data: nil})
	}
}

// handleQiShuiMusic 处理汽水音乐歌单
func handleQiShuiMusic(c *app.RequestContext, link string, detailed bool, format, order string) {
	songList, err := logic.QiShuiMusicDiscover(link, detailed)
	if err != nil {
		slog.Error("获取汽水音乐歌单失败", "err", err)
		c.JSON(consts.StatusBadRequest, &models.Result{Code: models.FailureCode, Msg: err.Error(), Data: nil})
		return
	}

	// 根据格式选项处理歌曲列表
	formatSongList(songList, format)

	// 根据顺序选项处理歌曲列表
	processSongOrder(songList, order)

	c.JSON(consts.StatusOK, &models.Result{Code: models.SuccessCode, Msg: SUCCESS, Data: songList})
}

// processSongOrder 根据指定的顺序处理歌曲列表
func processSongOrder(songList *models.SongList, order string) {
	if songList == nil || len(songList.Songs) == 0 {
		return
	}

	// 如果是倒序，则反转歌曲列表
	if order == "reverse" {
		songs := songList.Songs
		for i, j := 0, len(songs)-1; i < j; i, j = i+1, j-1 {
			songs[i], songs[j] = songs[j], songs[i]
		}
	}
}

// formatSongList 根据指定的格式处理歌曲列表
func formatSongList(songList *models.SongList, format string) {
	if songList == nil || len(songList.Songs) == 0 {
		return
	}

	// 如果没有指定格式或格式为默认的"歌名-歌手"，则不做处理
	if format == "" || format == "song-singer" {
		return
	}

	formattedSongs := make([]string, 0, len(songList.Songs))

	for _, song := range songList.Songs {
		switch format {
		case "singer-song":
			// 将"歌名 - 歌手"转换为"歌手 - 歌名"
			parts := strings.Split(song, " - ")
			if len(parts) == 2 {
				formattedSongs = append(formattedSongs, parts[1]+" - "+parts[0])
			} else {
				// 如果格式不符合预期，保持原样
				formattedSongs = append(formattedSongs, song)
			}
		case "song":
			// 只保留歌名
			parts := strings.Split(song, " - ")
			if len(parts) > 0 {
				formattedSongs = append(formattedSongs, parts[0])
			} else {
				formattedSongs = append(formattedSongs, song)
			}
		default:
			// 未知格式，保持原样
			formattedSongs = append(formattedSongs, song)
		}
	}

	// 更新歌曲列表
	songList.Songs = formattedSongs
}
