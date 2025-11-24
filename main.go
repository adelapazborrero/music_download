package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

// Screen represents different UI states
type Screen int

const (
	ScreenMenu Screen = iota
	ScreenSearchInput
	ScreenURLInput
	ScreenSearch
	ScreenResults
	ScreenLoading
	ScreenDetails
	ScreenDownloading
)

type model struct {
	screen        Screen
	searchQuery   string
	results       []SearchResult
	cursor        int
	menuCursor    int
	textInput     string
	selected      *VideoMetadata
	action        string
	err           error
	downloading   bool
	message       string
	height        int
	previewing    bool
	previewCmd    *exec.Cmd
	fromURL       bool
}

type searchCompleteMsg struct {
	results []SearchResult
	err     error
}

type metadataFetchedMsg struct {
	metadata *VideoMetadata
	err      error
}

type downloadCompleteMsg struct {
	err error
}

func initialModel(query string) model {
	if query != "" {
		return model{
			screen:      ScreenSearch,
			searchQuery: query,
		}
	}
	return model{
		screen: ScreenMenu,
	}
}

func (m model) Init() tea.Cmd {
	if m.searchQuery != "" {
		return searchYouTube(m.searchQuery)
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case ScreenResults:
			return m.updateResults(msg)
		case ScreenDetails:
			return m.updateDetails(msg)
		}

	case searchCompleteMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.results = msg.results
		m.screen = ScreenResults
		return m, nil

	case metadataFetchedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.selected = msg.metadata
		m.screen = ScreenDetails
		return m, nil

	case downloadCompleteMsg:
		m.downloading = false
		if msg.err != nil {
			m.message = fmt.Sprintf("Download failed: %v", msg.err)
		} else {
			m.message = "Download complete!"
		}
		return m, tea.Quit
	}

	return m, nil
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case "down", "j":
		if m.menuCursor < 1 {
			m.menuCursor++
		}
	case "enter":
		if m.menuCursor == 0 {
			// Search music
			m.screen = ScreenSearchInput
			m.textInput = ""
		} else {
			// Download from URL
			m.screen = ScreenURLInput
			m.textInput = ""
		}
		return m, nil
	}
	return m, nil
}

func (m model) updateSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
			m.screen = ScreenSearch
			return m, searchYouTube(m.searchQuery)
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

func (m model) updateURLInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.screen = ScreenMenu
		m.textInput = ""
		return m, nil
	case "enter":
		if m.textInput != "" {
			videoID := extractVideoID(m.textInput)
			if videoID == "" {
				m.message = "Invalid YouTube URL"
				return m, nil
			}
			m.fromURL = true
			m.screen = ScreenLoading
			return m, fetchMetadata(videoID)
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

func (m model) updateResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.results)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.results) > 0 {
			selected := m.results[m.cursor]
			m.screen = ScreenLoading
			return m, fetchMetadata(selected.ID)
		}
	}
	return m, nil
}

func (m model) updateDetails(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		return m, downloadVideo(m.selected.ID, m.selected.Title)
	}
	return m, nil
}

func (m model) View() string {
	switch m.screen {
	case ScreenMenu:
		return menuView(m)
	case ScreenSearchInput:
		return searchInputView(m)
	case ScreenURLInput:
		return urlInputView(m)
	case ScreenSearch:
		return searchingView(m.searchQuery)
	case ScreenResults:
		return resultsView(m)
	case ScreenLoading:
		return loadingView()
	case ScreenDetails:
		return detailsView(m)
	case ScreenDownloading:
		return downloadingView(m)
	}
	return ""
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Padding(1, 0)

	detailStyle = lipgloss.NewStyle().
			Padding(0, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

func menuView(m model) string {
	s := titleStyle.Render("Music Download") + "\n\n"
	s += "  What would you like to do?\n\n"

	options := []string{"Search music", "Download from URL"}
	for i, option := range options {
		cursor := "  "
		if m.menuCursor == i {
			cursor = "> "
			s += selectedStyle.Render(fmt.Sprintf("%s%s", cursor, option)) + "\n"
		} else {
			s += fmt.Sprintf("%s%s\n", cursor, option)
		}
	}

	s += helpStyle.Render("\nup/k up • down/j down • enter select • q quit")
	return s
}

func searchInputView(m model) string {
	s := titleStyle.Render("Search Music") + "\n\n"
	s += "  Enter search terms:\n\n"
	s += fmt.Sprintf("  > %s_\n", m.textInput)
	s += helpStyle.Render("\nenter submit • esc back • ctrl+c quit")
	return s
}

func urlInputView(m model) string {
	s := titleStyle.Render("Download from URL") + "\n\n"
	s += "  Enter YouTube URL:\n\n"
	s += fmt.Sprintf("  > %s_\n", m.textInput)
	if m.message != "" {
		s += "\n  " + m.message + "\n"
	}
	s += helpStyle.Render("\nenter submit • esc back • ctrl+c quit")
	return s
}

func searchingView(query string) string {
	return fmt.Sprintf("\nSearching YouTube for: %s\n\n", query)
}

func loadingView() string {
	return "\nLoading video details...\n\n"
}

func resultsView(m model) string {
	s := titleStyle.Render("Search Results") + "\n\n"

	for i, result := range m.results {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
			s += selectedStyle.Render(fmt.Sprintf("%s%s", cursor, result.Title)) + "\n"
		} else {
			s += fmt.Sprintf("%s%s\n", cursor, result.Title)
		}
	}

	s += helpStyle.Render("\nup/k up • down/j down • enter select • q quit")
	return s
}

func detailsView(m model) string {
	if m.selected == nil {
		return "Loading details..."
	}

	duration := formatDuration(m.selected.Duration)
	views := formatNumber(m.selected.ViewCount)

	s := titleStyle.Render("Video Details") + "\n\n"
	s += fmt.Sprintf("  Title:    %s\n", m.selected.Title)
	s += fmt.Sprintf("  Channel:  %s\n", m.selected.Channel)
	s += fmt.Sprintf("  Duration: %s\n", duration)
	s += fmt.Sprintf("  Views:    %s\n", views)

	if m.message != "" {
		s += "\n  " + m.message + "\n"
	}

	helpText := "\nup/k up • down/j down • enter select • q quit"
	if m.previewing {
		helpText = "\ns stop preview • d download • esc back • q quit"
	} else {
		helpText = "\np preview • d download • esc back • q quit"
	}
	s += helpStyle.Render(helpText)
	return s
}

func downloadingView(m model) string {
	s := titleStyle.Render("Downloading") + "\n\n"
	s += fmt.Sprintf("  Downloading: %s\n", m.selected.Title)
	s += "  Extracting high-quality MP3 with cover art...\n"
	return s
}

// Commands
func searchYouTube(query string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("yt-dlp",
			fmt.Sprintf("ytsearch20:%s", query),
			"--flat-playlist",
			"--print", "%(title)s|||%(id)s",
		)

		output, err := cmd.Output()
		if err != nil {
			return searchCompleteMsg{err: fmt.Errorf("search failed: %w", err)}
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
			return searchCompleteMsg{err: fmt.Errorf("no results found")}
		}

		return searchCompleteMsg{results: results}
	}
}

func fetchMetadata(videoID string) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
		cmd := exec.Command("yt-dlp", "-j", url)

		output, err := cmd.Output()
		if err != nil {
			return metadataFetchedMsg{err: fmt.Errorf("failed to fetch metadata: %w", err)}
		}

		var metadata VideoMetadata
		if err := json.Unmarshal(output, &metadata); err != nil {
			return metadataFetchedMsg{err: fmt.Errorf("failed to parse metadata: %w", err)}
		}

		return metadataFetchedMsg{metadata: &metadata}
	}
}

func downloadVideo(videoID, title string) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
		cmd := exec.Command("yt-dlp",
			"-f", "bestaudio",
			"--extract-audio",
			"--audio-format", "mp3",
			"--audio-quality", "0",
			"--embed-thumbnail",
			"--add-metadata",
			"-o", "%(title)s.%(ext)s",
			url,
		)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return downloadCompleteMsg{err: fmt.Errorf("download failed: %w", err)}
		}

		return downloadCompleteMsg{}
	}
}

// Helper functions
func formatDuration(seconds int) string {
	minutes := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

func formatNumber(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var result strings.Builder
	for i, digit := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(digit)
	}
	return result.String()
}

func extractVideoID(input string) string {
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

func checkDependencies() error {
	required := []string{"yt-dlp", "mpv", "ffmpeg"}
	for _, cmd := range required {
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("Required tool '%s' is not installed. Please install it first", cmd)
		}
	}
	return nil
}

func main() {
	if err := checkDependencies(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var query string
	if len(os.Args) >= 2 {
		// Old behavior: command line arguments
		query = strings.Join(os.Args[1:], " ")
	}
	// If no arguments, query will be empty and menu will be shown

	p := tea.NewProgram(initialModel(query))
	m, err := p.Run()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print final message if any
	if finalModel, ok := m.(model); ok {
		if finalModel.err != nil {
			fmt.Printf("Error: %v\n", finalModel.err)
			os.Exit(1)
		}
		if finalModel.message != "" {
			fmt.Println(finalModel.message)
		}
	}
}
