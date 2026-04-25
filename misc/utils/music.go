package utils

import (
	"regexp"
	"strings"
)

const (
	bracketsPattern = `（|）`
	miscPattern     = `\s?【[^】]*】`
)

var (
	bracketsRegex = regexp.MustCompile(bracketsPattern)
	miscRegex     = regexp.MustCompile(miscPattern)
)

// StandardSongName 获取标准化歌名
func StandardSongName(songName string) string {
	return strings.TrimSpace(miscRegex.ReplaceAllString(replaceCNBrackets(songName), ""))
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
