package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adelapazborrero/music_download/internal/app"
	"github.com/adelapazborrero/music_download/internal/youtube"
)

func main() {
	// Check dependencies first
	if err := youtube.CheckDependencies(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Parse command line arguments
	var query string
	if len(os.Args) >= 2 {
		// Old behavior: command line arguments
		query = strings.Join(os.Args[1:], " ")
	}
	// If no arguments, query will be empty and menu will be shown

	// Create and run the bubbletea program
	p := tea.NewProgram(app.InitialModel(query))
	m, err := p.Run()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print final message if any
	if finalModel, ok := m.(app.Model); ok {
		if finalModel.Error() != nil {
			fmt.Printf("Error: %v\n", finalModel.Error())
			os.Exit(1)
		}
		if finalModel.Message() != "" {
			fmt.Println(finalModel.Message())
		}
	}
}
