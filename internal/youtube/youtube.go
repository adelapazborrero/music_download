package youtube

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SearchResult represents a YouTube video from search results
type SearchResult struct {
	Title string
	ID    string
}

// VideoMetadata represents detailed video information
type VideoMetadata struct {
	Title     string `json:"title"`
	Channel   string `json:"channel"`
	Duration  int    `json:"duration"`
	ViewCount int64  `json:"view_count"`
	ID        string `json:"id"`
}

// Messages (exported so they can be used in app package)
type SearchCompleteMsg struct {
	Results []SearchResult
	Err     error
}

type MetadataFetchedMsg struct {
	Metadata *VideoMetadata
	Err      error
}

type DownloadCompleteMsg struct {
	Err error
}

// SearchYouTube performs a YouTube search with the given query and limit
func SearchYouTube(query string, limit int) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("yt-dlp",
			fmt.Sprintf("ytsearch%d:%s", limit, query),
			"--flat-playlist",
			"--print", "%(title)s|||%(id)s",
		)

		output, err := cmd.Output()
		if err != nil {
			return SearchCompleteMsg{Err: fmt.Errorf("search failed: %w", err)}
		}

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		results := make([]SearchResult, 0, len(lines))

		for _, line := range lines {
			parts := strings.Split(line, "|||")
			if len(parts) == 2 {
				results = append(results, SearchResult{
					Title: parts[0],
					ID:    parts[1],
				})
			}
		}

		if len(results) == 0 {
			return SearchCompleteMsg{Err: fmt.Errorf("no results found")}
		}

		return SearchCompleteMsg{Results: results}
	}
}

// FetchMetadata retrieves detailed metadata for a video
func FetchMetadata(videoID string) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
		cmd := exec.Command("yt-dlp", "-j", url)

		output, err := cmd.Output()
		if err != nil {
			return MetadataFetchedMsg{Err: fmt.Errorf("failed to fetch metadata: %w", err)}
		}

		var metadata VideoMetadata
		if err := json.Unmarshal(output, &metadata); err != nil {
			return MetadataFetchedMsg{Err: fmt.Errorf("failed to parse metadata: %w", err)}
		}

		return MetadataFetchedMsg{Metadata: &metadata}
	}
}

// DownloadVideo downloads a video as MP3
func DownloadVideo(videoID, title string) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
		cmd := exec.Command("yt-dlp",
			"-f", "bestaudio",
			"--extract-audio",
			"--audio-format", "mp3",
			"--audio-quality", "0",
			"--embed-thumbnail",
			"--add-metadata",
			"--quiet", // Suppress yt-dlp output to avoid UI interference
			"--progress",
			"-o", "%(title)s.%(ext)s",
			url,
		)

		// Don't redirect output to avoid breaking the TUI
		// cmd.Stdout and cmd.Stderr are nil by default, which discards output

		if err := cmd.Run(); err != nil {
			return DownloadCompleteMsg{Err: fmt.Errorf("download failed: %w", err)}
		}

		return DownloadCompleteMsg{}
	}
}

// ExtractVideoID extracts the video ID from various YouTube URL formats
func ExtractVideoID(input string) string {
	input = strings.TrimSpace(input)

	// If it's already just an ID (11 characters)
	if len(input) == 11 && !strings.Contains(input, "/") && !strings.Contains(input, ".") {
		return input
	}

	// Handle various YouTube URL formats
	// https://www.youtube.com/watch?v=VIDEO_ID
	// https://youtu.be/VIDEO_ID
	// https://www.youtube.com/embed/VIDEO_ID
	// https://m.youtube.com/watch?v=VIDEO_ID

	if strings.Contains(input, "youtube.com/watch?v=") {
		parts := strings.Split(input, "v=")
		if len(parts) >= 2 {
			videoID := parts[1]
			// Remove any additional parameters
			if idx := strings.Index(videoID, "&"); idx != -1 {
				videoID = videoID[:idx]
			}
			return videoID
		}
	}

	if strings.Contains(input, "youtu.be/") {
		parts := strings.Split(input, "youtu.be/")
		if len(parts) >= 2 {
			videoID := parts[1]
			// Remove any additional parameters
			if idx := strings.Index(videoID, "?"); idx != -1 {
				videoID = videoID[:idx]
			}
			return videoID
		}
	}

	if strings.Contains(input, "youtube.com/embed/") {
		parts := strings.Split(input, "embed/")
		if len(parts) >= 2 {
			videoID := parts[1]
			// Remove any additional parameters
			if idx := strings.Index(videoID, "?"); idx != -1 {
				videoID = videoID[:idx]
			}
			return videoID
		}
	}

	return ""
}

// CheckDependencies verifies that required tools are installed
func CheckDependencies() error {
	required := []string{"yt-dlp", "mpv", "ffmpeg"}
	for _, cmd := range required {
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("Required tool '%s' is not installed. Please install it first", cmd)
		}
	}
	return nil
}
