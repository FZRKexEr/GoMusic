package logic

import (
	"encoding/json"
	"errors"
	"strings"

	"GoMusic/misc/models"
	"GoMusic/misc/utils"
)

func parseRouterDataSongList(html []byte, detailed bool) (*models.SongList, bool) {
	routerData, ok := extractRouterData(string(html))
	if !ok {
		return nil, false
	}

	var data qishuiRouterData
	if err := json.Unmarshal(routerData, &data); err != nil {
		return nil, false
	}

	songList, err := data.toSongList(detailed)
	if err != nil || songList.SongsCount == 0 {
		return nil, false
	}
	return songList, true
}

func extractRouterData(source string) ([]byte, bool) {
	idx := strings.Index(source, routerDataAssignment)
	if idx < 0 {
		return nil, false
	}

	routerData := source[idx+len(routerDataAssignment):]
	for _, marker := range []string{"\nfunction ", ";function ", "</script>"} {
		if idx := strings.Index(routerData, marker); idx >= 0 {
			routerData = routerData[:idx]
			break
		}
	}
	routerData = strings.TrimSpace(strings.TrimSuffix(routerData, ";"))
	return []byte(routerData), routerData != ""
}

type qishuiRouterData struct {
	LoaderData struct {
		PlaylistPage struct {
			Medias       []qishuiMedia `json:"medias"`
			PlaylistInfo struct {
				Title string `json:"title"`
				Owner struct {
					Nickname   string `json:"nickname"`
					PublicName string `json:"public_name"`
				} `json:"owner"`
			} `json:"playlistInfo"`
		} `json:"playlist_page"`
	} `json:"loaderData"`
}

type qishuiMedia struct {
	Type   string `json:"type"`
	Entity struct {
		Track *qishuiTrack `json:"track"`
		Video *qishuiVideo `json:"video"`
	} `json:"entity"`
}

type qishuiTrack struct {
	Name    string              `json:"name"`
	Artists []qishuiTrackArtist `json:"artists"`
}

type qishuiVideo struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Artists     []qishuiVideoArtist `json:"artists"`
}

type qishuiTrackArtist struct {
	Name              string `json:"name"`
	SimpleDisplayName string `json:"simple_display_name"`
}

type qishuiVideoArtist struct {
	Name     string `json:"name"`
	UserInfo struct {
		Nickname   string `json:"nickname"`
		PublicName string `json:"public_name"`
	} `json:"user_info"`
}

func (data qishuiRouterData) toSongList(detailed bool) (*models.SongList, error) {
	page := data.LoaderData.PlaylistPage
	ownerName := firstNonEmpty(page.PlaylistInfo.Owner.Nickname, page.PlaylistInfo.Owner.PublicName)
	songList := models.SongList{
		Name:       formatPlaylistName(strings.TrimSpace(page.PlaylistInfo.Title), strings.TrimSpace(ownerName)),
		SongsCount: 0,
	}

	for _, media := range page.Medias {
		title, artist := media.songInfo()
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}
		if !detailed {
			title = utils.StandardSongName(title)
		}

		formattedSong := formatSong(title, artist)
		if formattedSong == "" {
			continue
		}
		songList.Songs = append(songList.Songs, formattedSong)
	}

	songList.SongsCount = len(songList.Songs)
	if songList.SongsCount == 0 {
		return nil, errors.New("未解析到汽水音乐歌曲")
	}
	return &songList, nil
}

func (media qishuiMedia) songInfo() (string, string) {
	if media.Entity.Track != nil {
		return media.Entity.Track.Name, joinNonEmpty(trackArtistNames(media.Entity.Track.Artists), ", ")
	}
	if media.Entity.Video != nil {
		return firstNonEmpty(media.Entity.Video.Title, media.Entity.Video.Description), joinNonEmpty(videoArtistNames(media.Entity.Video.Artists), ", ")
	}
	return "", ""
}

func trackArtistNames(artists []qishuiTrackArtist) []string {
	names := make([]string, 0, len(artists))
	for _, artist := range artists {
		names = append(names, firstNonEmpty(artist.Name, artist.SimpleDisplayName))
	}
	return names
}

func videoArtistNames(artists []qishuiVideoArtist) []string {
	names := make([]string, 0, len(artists))
	for _, artist := range artists {
		names = append(names, firstNonEmpty(artist.Name, artist.UserInfo.Nickname, artist.UserInfo.PublicName))
	}
	return names
}

func joinNonEmpty(values []string, sep string) string {
	trimmed := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			trimmed = append(trimmed, value)
		}
	}
	return strings.Join(trimmed, sep)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
