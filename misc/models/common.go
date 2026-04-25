package models

// SongList represents a playlist response.
type SongList struct {
	Name       string   `json:"name"`
	Songs      []string `json:"songs"`
	SongsCount int      `json:"songs_count"`
}
