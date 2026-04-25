package models

const (
	Port       = 8081
	PortFormat = ":%d"
)

const (
	SuccessCode = 1
	FailureCode = -1
)

// SongList represents a playlist response.
type SongList struct {
	Name       string   `json:"name"`
	Songs      []string `json:"songs"`
	SongsCount int      `json:"songs_count"`
}
