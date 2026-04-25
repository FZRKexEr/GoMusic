package utils

import (
	"net/url"
	"regexp"

	"GoMusic/misc/log"
)

const (
	bracketsPattern = `（|）`     // 去除特殊符号
	miscPattern     = `\s?【.*】` // 去除特殊符号
)

var (
	bracketsRegex = regexp.MustCompile(bracketsPattern)
	miscRegex     = regexp.MustCompile(miscPattern)
)

func GetQQMusicParam(link string) (string, error) {
	parse, err := url.ParseRequestURI(link)
	if err != nil {
		log.Errorf("fail to parse url: %v", err)
		return "", err
	}
	query, err := url.ParseQuery(parse.RawQuery)
	if err != nil {
		log.Errorf("fail to parse query: %v", err)
		return "", err
	}
	id := query.Get("id")
	return id, nil
}

// StandardSongName 获取标准化歌名
func StandardSongName(songName string) string {
	return miscRegex.ReplaceAllString(replaceCNBrackets(songName), "")
}

// 将中文括号替换为英文括号
func replaceCNBrackets(s string) string {
	return bracketsRegex.ReplaceAllStringFunc(s, func(m string) string {
		if m == "（" {
			return " (" // 左括号前面追加空格
		}
		return ")"
	})
}
