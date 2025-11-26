package app

import (
	"fmt"
	"os/exec"

	"github.com/adelapazborrero/music_download/internal/youtube"
	tea "github.com/charmbracelet/bubbletea"
)

// Init initializes the application
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Enable bracketed paste for clipboard paste support
	cmds = append(cmds, tea.EnableBracketedPaste)

	if m.searchQuery != "" {
		cmds = append(cmds, youtube.SearchYouTube(m.searchQuery, m.searchLimit))
	}

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

// Update handles all state updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.screen {
		case ScreenMenu:
			return m.updateMenu(msg)
		case ScreenSearchInput:
			return m.updateSearchInput(msg)
		case ScreenURLInput:
			return m.updateURLInput(msg)
		case ScreenPlaylistInput:
			return m.updatePlaylistInput(msg)
		case ScreenResults:
			return m.updateResults(msg)
		case ScreenDetails:
			return m.updateDetails(msg)
		}

	case youtube.SearchCompleteMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, tea.Quit
		}
		m.results = msg.Results
		m.screen = ScreenResults
		return m, nil

	case youtube.MetadataFetchedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, tea.Quit
		}

		// Update with full metadata
		m.selected = msg.Metadata
		m.screen = ScreenDetails

		// Only start preview if not already previewing (e.g., from URL input)
		// When coming from search results, preview is already started
		if !m.previewing {
			// Auto-start preview (for URL input flow)
			url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", msg.Metadata.ID)
			cmd := exec.Command("mpv", "--no-video", "--ytdl-format=bestaudio", url)
			m.previewCmd = cmd
			go cmd.Run()
			m.previewing = true
			m.message = "Playing preview... (press 's' to stop)"
		}

		return m, nil

	case youtube.DownloadCompleteMsg:
		m.downloading = false
		if msg.Err != nil {
			m.message = "Download failed: " + msg.Err.Error()
		} else {
			m.message = "✓ Download complete!"
		}
		// Return to details screen instead of quitting
		m.screen = ScreenDetails
		return m, nil

	case youtube.PlaylistFetchedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, tea.Quit
		}
		m.playlistItems = msg.Items
		m.playlistTotal = len(msg.Items)
		m.playlistSuccess = 0
		m.playlistFailed = 0
		m.playlistFailedItems = []string{}
		m.message = fmt.Sprintf("Found %d songs in playlist. Starting download...", len(msg.Items))
		m.screen = ScreenPlaylistDownloading
		return m, youtube.DownloadPlaylist(msg.Items)

	case youtube.PlaylistDownloadProgressMsg:
		// Update progress and counts
		m.playlistProgress = msg.Current
		if msg.Success {
			m.playlistSuccess++
			m.message = fmt.Sprintf("✓ Downloaded: %s (%d/%d)", msg.Title, msg.Current, msg.Total)
		} else {
			m.playlistFailed++
			failedMsg := fmt.Sprintf("%s: %s", msg.Title, msg.Error)
			m.playlistFailedItems = append(m.playlistFailedItems, failedMsg)
			m.message = fmt.Sprintf("✗ Failed: %s (%d/%d)", msg.Title, msg.Current, msg.Total)
		}
		// Continue downloading next item with accumulated counts
		return m, youtube.DownloadNextPlaylistItem(m.playlistItems, msg.Current, m.playlistSuccess, m.playlistFailed, m.playlistFailedItems)

	case youtube.PlaylistDownloadCompleteMsg:
		if msg.Err != nil {
			m.message = fmt.Sprintf("Playlist download failed: %s", msg.Err.Error())
		} else {
			m.message = fmt.Sprintf("✓ Playlist download complete! Success: %d, Failed: %d", msg.Success, msg.Failed)
			if msg.Failed > 0 && len(msg.FailedItems) > 0 {
				m.message += "\n\nFailed downloads:\n"
				for _, item := range msg.FailedItems {
					m.message += fmt.Sprintf("  • %s\n", item)
				}
			}
		}
		m.screen = ScreenMenu
		m.playlistItems = nil
		m.playlistProgress = 0
		m.playlistTotal = 0
		return m, nil
	}

	return m, nil
}

func (m Model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case "down", "j":
		if m.menuCursor < 2 {
			m.menuCursor++
		}
	case "enter":
		if m.menuCursor == 0 {
			// Search music
			m.screen = ScreenSearchInput
			m.textInput = ""
		} else if m.menuCursor == 1 {
			// Download from URL
			m.screen = ScreenURLInput
			m.textInput = ""
		} else {
			// Download from playlist
			m.screen = ScreenPlaylistInput
			m.textInput = ""
		}
		return m, nil
	}
	return m, nil
}

func (m Model) updateSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle paste events
	if msg.Paste {
		m.textInput += msg.String()
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.screen = ScreenMenu
		m.textInput = ""
		return m, nil
	case "enter":
		if m.textInput != "" {
			m.searchQuery = m.textInput
			m.searchLimit = 20 // Reset to 20 for new search
			m.screen = ScreenSearch
			return m, youtube.SearchYouTube(m.searchQuery, m.searchLimit)
		}
		return m, nil
	case "backspace":
		if len(m.textInput) > 0 {
			m.textInput = m.textInput[:len(m.textInput)-1]
		}
	default:
		// Add typed character if it's a single character
		if len(msg.String()) == 1 {
			m.textInput += msg.String()
		} else if msg.String() == "space" {
			m.textInput += " "
		}
	}
	return m, nil
}

func (m Model) updateURLInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle paste events
	if msg.Paste {
		m.textInput += msg.String()
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.screen = ScreenMenu
		m.textInput = ""
		return m, nil
	case "enter":
		if m.textInput != "" {
			videoID := youtube.ExtractVideoID(m.textInput)
			if videoID == "" {
				m.message = "Invalid YouTube URL"
				return m, nil
			}
			m.fromURL = true
			m.screen = ScreenLoading
			return m, youtube.FetchMetadata(videoID)
		}
		return m, nil
	case "backspace":
		if len(m.textInput) > 0 {
			m.textInput = m.textInput[:len(m.textInput)-1]
		}
	default:
		// Add typed character if it's a single character
		if len(msg.String()) == 1 {
			m.textInput += msg.String()
		}
	}
	return m, nil
}

func (m Model) updateResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxCursor := len(m.results) // +1 for "Load more" option

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		// Go back to main menu
		m.screen = ScreenMenu
		m.results = nil
		m.cursor = 0
		m.searchQuery = ""
		m.searchLimit = 20
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < maxCursor {
			m.cursor++
		}
	case "enter":
		// Check if "Load more" option is selected
		if m.cursor == len(m.results) {
			// Load more results
			m.searchLimit += 20
			m.cursor = 0 // Reset cursor
			m.screen = ScreenSearch
			return m, youtube.SearchYouTube(m.searchQuery, m.searchLimit)
		}

		// Regular result selected
		if len(m.results) > 0 && m.cursor < len(m.results) {
			selected := m.results[m.cursor]

			// Create partial metadata from search result
			m.selected = &youtube.VideoMetadata{
				Title: selected.Title,
				ID:    selected.ID,
			}

			// Start preview immediately
			url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", selected.ID)
			cmd := exec.Command("mpv", "--no-video", "--ytdl-format=bestaudio", url)
			m.previewCmd = cmd
			go cmd.Run()
			m.previewing = true
			m.message = "Playing preview... (press 's' to stop)"

			// Go to details screen and fetch full metadata in background
			m.screen = ScreenDetails
			return m, youtube.FetchMetadata(selected.ID)
		}
	}
	return m, nil
}

func (m Model) updatePlaylistInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle paste events
	if msg.Paste {
		m.textInput += msg.String()
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.screen = ScreenMenu
		m.textInput = ""
		return m, nil
	case "enter":
		if m.textInput != "" {
			playlistID := youtube.ExtractPlaylistID(m.textInput)
			if playlistID == "" {
				m.message = "Invalid YouTube playlist URL"
				return m, nil
			}
			m.screen = ScreenLoading
			m.message = "Fetching playlist..."
			return m, youtube.FetchPlaylistItems(playlistID)
		}
		return m, nil
	case "backspace":
		if len(m.textInput) > 0 {
			m.textInput = m.textInput[:len(m.textInput)-1]
		}
	default:
		// Add typed character if it's a single character
		if len(msg.String()) == 1 {
			m.textInput += msg.String()
		}
	}
	return m, nil
}

func (m Model) updateDetails(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.previewing && m.previewCmd != nil {
			m.previewCmd.Process.Kill()
		}
		return m, tea.Quit
	case "esc":
		if m.previewing && m.previewCmd != nil {
			m.previewCmd.Process.Kill()
			m.previewing = false
			m.previewCmd = nil
		}
		if m.fromURL {
			m.screen = ScreenMenu
			m.fromURL = false
		} else {
			m.screen = ScreenResults
		}
		m.selected = nil
		m.message = ""
		return m, nil
	case "p":
		if !m.previewing {
			url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", m.selected.ID)
			cmd := exec.Command("mpv", "--no-video", "--ytdl-format=bestaudio", url)
			m.previewCmd = cmd
			go cmd.Run()
			m.previewing = true
			m.message = "Playing preview... (press 's' to stop)"
		}
		return m, nil
	case "s":
		if m.previewing && m.previewCmd != nil {
			m.previewCmd.Process.Kill()
			m.previewing = false
			m.previewCmd = nil
			m.message = "Preview stopped"
		}
		return m, nil
	case "d":
		if m.previewing && m.previewCmd != nil {
			m.previewCmd.Process.Kill()
			m.previewing = false
			m.previewCmd = nil
		}
		m.downloading = true
		m.screen = ScreenDownloading
		return m, youtube.DownloadVideo(m.selected.ID, m.selected.Title)
	}
	return m, nil
}
