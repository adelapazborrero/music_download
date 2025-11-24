package app

import (
	"fmt"

	"github.com/adelapazborrero/music_download/internal/ui"
	"github.com/adelapazborrero/music_download/internal/utils"
)

// View renders the appropriate screen based on current state
func (m Model) View() string {
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

func menuView(m Model) string {
	s := ui.TitleStyle.Render("Music Download") + "\n\n"
	s += "  What would you like to do?\n\n"

	options := []string{"Search music", "Download from URL"}
	for i, option := range options {
		cursor := "  "
		if m.menuCursor == i {
			cursor = "> "
			s += ui.SelectedStyle.Render(fmt.Sprintf("%s%s", cursor, option)) + "\n"
		} else {
			s += fmt.Sprintf("%s%s\n", cursor, option)
		}
	}

	s += ui.HelpStyle.Render("\nup/k up • down/j down • enter select • q quit")
	return s
}

func searchInputView(m Model) string {
	s := ui.TitleStyle.Render("Search Music") + "\n\n"
	s += "  Enter search terms:\n\n"
	s += fmt.Sprintf("  > %s_\n", m.textInput)
	s += ui.HelpStyle.Render("\nenter submit • esc back • ctrl+c quit")
	return s
}

func urlInputView(m Model) string {
	s := ui.TitleStyle.Render("Download from URL") + "\n\n"
	s += "  Enter YouTube URL:\n\n"
	s += fmt.Sprintf("  > %s_\n", m.textInput)
	if m.message != "" {
		s += "\n  " + m.message + "\n"
	}
	s += ui.HelpStyle.Render("\nenter submit • esc back • ctrl+c quit")
	return s
}

func searchingView(query string) string {
	return fmt.Sprintf("\nSearching YouTube for: %s\n\n", query)
}

func loadingView() string {
	return "\nLoading video details...\n\n"
}

func resultsView(m Model) string {
	s := ui.TitleStyle.Render("Search Results") + "\n\n"

	for i, result := range m.results {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
			s += ui.SelectedStyle.Render(fmt.Sprintf("%s%s", cursor, result.Title)) + "\n"
		} else {
			s += fmt.Sprintf("%s%s\n", cursor, result.Title)
		}
	}

	// Add "Load more" option
	loadMoreText := "Load more results..."
	cursor := "  "
	if m.cursor == len(m.results) {
		cursor = "> "
		s += "\n" + ui.SelectedStyle.Render(fmt.Sprintf("%s%s", cursor, loadMoreText)) + "\n"
	} else {
		s += fmt.Sprintf("\n%s%s\n", cursor, loadMoreText)
	}

	s += ui.HelpStyle.Render("\nup/k up • down/j down • enter select • esc menu • q quit")
	return s
}

func detailsView(m Model) string {
	if m.selected == nil {
		return "Loading details..."
	}

	s := ui.TitleStyle.Render("Video Details") + "\n\n"
	s += fmt.Sprintf("  Title:    %s\n", m.selected.Title)

	// Show "Loading..." for fields not yet available
	if m.selected.Channel != "" {
		s += fmt.Sprintf("  Channel:  %s\n", m.selected.Channel)
	} else {
		s += "  Channel:  Loading...\n"
	}

	if m.selected.Duration > 0 {
		duration := utils.FormatDuration(m.selected.Duration)
		s += fmt.Sprintf("  Duration: %s\n", duration)
	} else {
		s += "  Duration: Loading...\n"
	}

	if m.selected.ViewCount > 0 {
		views := utils.FormatNumber(m.selected.ViewCount)
		s += fmt.Sprintf("  Views:    %s\n", views)
	} else {
		s += "  Views:    Loading...\n"
	}

	if m.message != "" {
		s += "\n  " + m.message + "\n"
	}

	helpText := "\nup/k up • down/j down • enter select • q quit"
	if m.previewing {
		helpText = "\ns stop preview • d download • esc back • q quit"
	} else {
		helpText = "\np preview • d download • esc back • q quit"
	}
	s += ui.HelpStyle.Render(helpText)
	return s
}

func downloadingView(m Model) string {
	s := ui.TitleStyle.Render("Downloading") + "\n\n"
	s += fmt.Sprintf("  Title:    %s\n", m.selected.Title)
	s += "\n"
	s += "  Status:   Downloading and converting to MP3...\n"
	s += "  Quality:  High-quality audio with cover art\n"
	s += "\n"
	s += "  Please wait, this may take a moment...\n"
	return s
}
