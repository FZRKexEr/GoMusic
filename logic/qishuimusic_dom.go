package logic

import (
	"errors"
	"strings"

	"GoMusic/misc/models"
	"GoMusic/misc/utils"

	"github.com/PuerkitoBio/goquery"
)

func parseDOMSongList(docDetail *goquery.Document, detailed bool) (*models.SongList, error) {
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

			artist = strings.TrimSpace(strings.Split(artist, "•")[0])
			if !detailed {
				title = utils.StandardSongName(title)
			}

			formattedSong := formatSong(title, artist)
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
