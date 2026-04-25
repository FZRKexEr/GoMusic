package logic

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"GoMusic/misc/httputil"
	"GoMusic/misc/models"

	"github.com/PuerkitoBio/goquery"
)

// 歌曲链接正则
const (
	qishuiMusicURL       = `https?://[^\s"'<>，。；;、)]+`
	playlistIDKey        = "playlist_id"
	routerDataAssignment = "_ROUTER_DATA = "
	qishuiShareHost      = "qishui.douyin.com"
	qishuiMusicHost      = "music.douyin.com"
	qishuiMusicPath      = "/qishui/"
)

var (
	qishuiMusicURLRegx = regexp.MustCompile(qishuiMusicURL)
)

type qishuiHTTPClient interface {
	Get(link string) (*http.Response, error)
	GetRedirectLocation(link string) (string, error)
}

type defaultQishuiHTTPClient struct{}

// QiShuiMusicDiscover 解析歌单
// link: 歌单链接
// detailed: 是否使用详细歌曲名（原始歌曲名，不去除括号等内容）
func QiShuiMusicDiscover(link string, detailed bool) (*models.SongList, error) {
	return discoverQiShuiMusic(link, detailed, defaultQishuiHTTPClient{})
}

func (defaultQishuiHTTPClient) Get(link string) (*http.Response, error) {
	return httputil.Get(link)
}

func (defaultQishuiHTTPClient) GetRedirectLocation(link string) (string, error) {
	return httputil.GetRedirectLocation(link)
}

func discoverQiShuiMusic(link string, detailed bool, client qishuiHTTPClient) (*models.SongList, error) {
	playlistURL, err := resolveQiShuiURL(link, client)
	if err != nil {
		return nil, err
	}

	resp, err := client.Get(playlistURL)
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

func IsQiShuiMusicLink(link string) bool {
	return isQiShuiURL(extractURL(link))
}

func resolveQiShuiURL(link string, client qishuiHTTPClient) (string, error) {
	extractedLink := extractURL(link)
	if extractedLink == "" {
		return "", errors.New("无效的汽水音乐链接")
	}
	if !isQiShuiURL(extractedLink) {
		return "", errors.New("不支持的汽水音乐链接")
	}

	if hasPlaylistID(extractedLink) {
		return extractedLink, nil
	}

	redirectedLink, err := client.GetRedirectLocation(extractedLink)
	if err != nil {
		return "", err
	}
	if redirectedLink == "" {
		return extractedLink, nil
	}
	resolvedURL, err := resolveRedirectURL(extractedLink, redirectedLink)
	if err != nil {
		return "", err
	}
	if !isQiShuiURL(resolvedURL) {
		return "", errors.New("不支持的汽水音乐链接")
	}
	return resolvedURL, nil
}

func isQiShuiURL(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	host := strings.ToLower(parsedURL.Hostname())
	switch host {
	case qishuiShareHost:
		return true
	case qishuiMusicHost:
		return strings.HasPrefix(parsedURL.EscapedPath(), qishuiMusicPath)
	default:
		return false
	}
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
	html, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	if songList, ok := parseRouterDataSongList(html, detailed); ok {
		return songList, nil
	}

	docDetail, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, err
	}

	return parseDOMSongList(docDetail, detailed)
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
