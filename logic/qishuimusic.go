package logic

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"GoMusic/misc/httputil"
	"GoMusic/misc/models"
	"GoMusic/misc/utils"

	"github.com/PuerkitoBio/goquery"
)

// 歌曲链接正则
const (
	qishuiMusicURL = `https?://[^\s"'<>，。；;、)]+`
	playlistIDKey  = "playlist_id"
)

var (
	qishuiMusicURLRegx = regexp.MustCompile(qishuiMusicURL)
)

// 歌曲信息列表#root > div > div > div > div > div:nth-child(2) > div > div > div > div > div 下的子元素nth-child
// 歌曲名称 div:nth-child(2) > div:nth-child(1) > p
// 歌曲作者 div:nth-child(2) > div:nth-child(2) > p

// QiShuiMusicDiscover 解析歌单
// link: 歌单链接
// detailed: 是否使用详细歌曲名（原始歌曲名，不去除括号等内容）
func QiShuiMusicDiscover(link string, detailed bool) (*models.SongList, error) {
	playlistURL, err := resolveQiShuiURL(link)
	if err != nil {
		return nil, err
	}

	resp, err := httputil.Get(playlistURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求汽水音乐页面失败: %s", resp.Status)
	}

	songList, err := parseQsSongList(resp.Body, detailed)
	if err != nil {
		return nil, err
	}
	return songList, nil
}

func resolveQiShuiURL(link string) (string, error) {
	extractedLink := extractURL(link)
	if extractedLink == "" {
		return "", errors.New("无效的汽水音乐链接")
	}

	if hasPlaylistID(extractedLink) {
		return extractedLink, nil
	}

	redirectedLink, err := httputil.GetRedirectLocation(extractedLink)
	if err != nil {
		return "", err
	}
	if redirectedLink == "" {
		return extractedLink, nil
	}
	return resolveRedirectURL(extractedLink, redirectedLink)
}

func extractURL(text string) string {
	rawURL := strings.TrimSpace(qishuiMusicURLRegx.FindString(text))
	rawURL = strings.Trim(rawURL, `"'<>`)
	return strings.TrimRight(rawURL, "，。；;、)")
}

func resolveRedirectURL(baseURL, location string) (string, error) {
	redirectedURL, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	if redirectedURL.IsAbs() {
		return redirectedURL.String(), nil
	}

	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	return parsedBaseURL.ResolveReference(redirectedURL).String(), nil
}

func hasPlaylistID(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return parsedURL.Query().Get(playlistIDKey) != ""
}

// parseQsSongList 解析网页
// detailed: 是否使用详细歌曲名（原始歌曲名，不去除括号等内容）
func parseQsSongList(body io.Reader, detailed bool) (*models.SongList, error) {
	docDetail, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}
	songListName := strings.TrimSpace(docDetail.Find("#root > div > div > div > div > div:nth-child(1) > div:nth-child(3) > h1 > p").Text())
	songListAuthor := strings.TrimSpace(docDetail.Find("#root > div > div > div > div > div:nth-child(1) > div:nth-child(3) > div > div > div:nth-child(2) > p").Text())
	songList := models.SongList{
		Name:       formatPlaylistName(songListName, songListAuthor),
		SongsCount: 0,
	}
	docDetail.Find("#root > div > div > div > div > div:nth-child(2) > div > div > div > div > div").Each(
		func(i int, s *goquery.Selection) {
			title := strings.TrimSpace(s.Find("div:nth-child(2) > div:nth-child(1) > p").Text())
			artist := strings.TrimSpace(s.Find("div:nth-child(2) > div:nth-child(2) > p").Text())
			if title == "" {
				return
			}

			// artist 需要格式化，去除 • 后面的字符，例如：G.E.M. 邓紫棋 • T.I.M.E. -> G.E.M. 邓紫棋
			artist = strings.TrimSpace(strings.Split(artist, "•")[0])

			// 根据detailed参数决定是否使用原始歌曲名
			var songName string
			if detailed {
				songName = title // 使用原始歌曲名
			} else {
				songName = utils.StandardSongName(title) // 使用标准化的歌曲名
			}

			formattedSong := formatSong(songName, artist)
			if formattedSong == "" {
				return
			}
			songList.Songs = append(songList.Songs, formattedSong)
			songList.SongsCount++
		},
	)
	if songList.SongsCount == 0 {
		return nil, errors.New("未解析到汽水音乐歌曲")
	}
	return &songList, nil
}

func formatPlaylistName(name, author string) string {
	switch {
	case name == "":
		return author
	case author == "":
		return name
	default:
		return fmt.Sprintf("%s-%s", name, author)
	}
}

func formatSong(name, artist string) string {
	name = strings.TrimSpace(name)
	artist = strings.TrimSpace(artist)
	if name == "" {
		return ""
	}
	if artist == "" {
		return name
	}
	return fmt.Sprintf("%s - %s", name, artist)
}
