package app

import (
	"os/exec"

	"github.com/adelapazborrero/music_download/internal/youtube"
)

// Screen represents different UI states
type Screen int

const (
	ScreenMenu Screen = iota
	ScreenSearchInput
	ScreenURLInput
	ScreenPlaylistInput
	ScreenSearch
	ScreenResults
	ScreenLoading
	ScreenDetails
	ScreenDownloading
	ScreenPlaylistDownloading
)

// Model holds the application state
type Model struct {
	screen              Screen
	searchQuery         string
	results             []youtube.SearchResult
	cursor              int
	menuCursor          int
	textInput           string
	selected            *youtube.VideoMetadata
	action              string
	err                 error
	downloading         bool
	message             string
	height              int
	previewing          bool
	previewCmd          *exec.Cmd
	fromURL             bool
	searchLimit         int
	playlistItems       []youtube.SearchResult
	playlistProgress    int
	playlistTotal       int
	playlistSuccess     int
	playlistFailed      int
	playlistFailedItems []string
}

// Getters for private fields (needed by main.go)
func (m Model) Error() error {
	return m.err
}

func (m Model) Message() string {
	return m.message
}

// InitialModel creates the initial application state
func InitialModel(query string) Model {
	if query != "" {
		return Model{
			screen:      ScreenSearch,
			searchQuery: query,
			searchLimit: 20,
		}
	}
	return Model{
		screen:      ScreenMenu,
		searchLimit: 20,
	}
}
