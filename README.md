# Music Download

A modern, interactive CLI tool for searching, previewing, and downloading music from YouTube as high-quality MP3 files with embedded metadata and cover art.

## Features

- 🔍 **Interactive Search** - Search YouTube directly from the terminal
- 🎵 **Auto-Preview** - Instant audio preview when selecting songs
- ⬇️ **High-Quality Downloads** - MP3 with embedded thumbnails and metadata
- 🌐 **URL Support** - Download directly from YouTube URLs
- 📊 **Video Details** - View title, channel, duration, and view count
- 🔄 **Load More** - Dynamically load additional search results
- 🎨 **Modern TUI** - Beautiful terminal interface with Bubbletea
- ⚡ **Parallel Loading** - Preview starts while metadata loads in background

## Prerequisites

The following tools must be installed on your system:

- **yt-dlp** - YouTube downloader
- **mpv** - Media player for audio preview
- **ffmpeg** - Audio processing and conversion
- **Go 1.25+** - For building from source

### Installing Dependencies

**macOS** (using Homebrew):
```bash
brew install yt-dlp mpv ffmpeg
```

**Ubuntu/Debian**:
```bash
sudo apt install yt-dlp mpv ffmpeg
```

**Arch Linux**:
```bash
sudo pacman -S yt-dlp mpv ffmpeg
```

## Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/adelapazborrero/music_download.git
cd music_download

# Build the application
make build

# (Optional) Install to system path
make install
```

### Option 2: Direct Go Install

```bash
go install github.com/adelapazborrero/music_download/cmd/music-download@latest
```

## Usage

### Interactive Mode (Recommended)

Start the application without arguments to access the main menu:

```bash
music-download
```

You'll be presented with two options:
1. **Search music** - Search YouTube and browse results
2. **Download from URL** - Directly download from a YouTube URL

### Command-Line Mode

Search directly from the command line:

```bash
music-download "lofi hip hop beats"
```

## Navigation

### Main Menu
- `↑/k` or `↓/j` - Navigate options
- `enter` - Select option
- `q` - Quit

### Search Input
- Type your search query
- `space` - Add space
- `backspace` - Delete character
- `enter` - Submit search
- `esc` - Back to menu
- `ctrl+c` - Quit

### Search Results
- `↑/k` or `↓/j` - Navigate results
- `enter` - Select song (starts preview immediately)
- `esc` - Back to main menu
- `q` - Quit
- **Load more results** - Select bottom option to load 20 more

### Video Details
- `p` - Start/resume preview (auto-starts on selection)
- `s` - Stop preview
- `d` - Download MP3
- `esc` - Back to results/menu
- `q` - Quit

### URL Input
- Paste YouTube URL (supports multiple formats)
- `enter` - Fetch and preview
- `esc` - Back to menu
- `ctrl+c` - Quit

## Supported URL Formats

The tool accepts various YouTube URL formats:

```
https://www.youtube.com/watch?v=VIDEO_ID
https://youtu.be/VIDEO_ID
https://www.youtube.com/embed/VIDEO_ID
https://m.youtube.com/watch?v=VIDEO_ID
VIDEO_ID (just the ID)
```

## Project Structure

```
music_download/
├── cmd/
│   └── music-download/
│       └── main.go              # Entry point
├── internal/
│   ├── app/
│   │   ├── model.go            # Application state
│   │   ├── update.go           # Event handlers
│   │   └── view.go             # UI rendering
│   ├── ui/
│   │   └── styles.go           # Lipgloss styles
│   ├── utils/
│   │   └── utils.go            # Helper functions
│   └── youtube/
│       └── youtube.go          # YouTube operations
├── Makefile                     # Build automation
├── go.mod                       # Go module definition
└── README.md                    # This file
```

## Makefile Commands

Run `make help` to see all available commands:

```bash
make help          # Show available commands
make build         # Build the application
make install       # Install to /usr/local/bin
make clean         # Remove build artifacts
make run           # Build and run the application
make fmt           # Format Go code
make vet           # Run Go vet
make test          # Run tests
make deps          # Update dependencies
```

### Common Workflows

**Development:**
```bash
# Format code and build
make fmt && make build

# Run the application
make run
```

**Production:**
```bash
# Build and install system-wide
make install

# Then use anywhere
music-download
```

**Cleanup:**
```bash
# Remove build artifacts
make clean
```

## How It Works

1. **Search Flow:**
   - Enter search terms
   - yt-dlp searches YouTube
   - Results displayed in interactive list
   - Select song → Preview starts immediately
   - Metadata loads in parallel (title, channel, duration, views)
   - Choose to download or continue browsing

2. **URL Flow:**
   - Enter YouTube URL
   - Video ID extracted
   - Preview starts immediately
   - Metadata fetched in parallel
   - Download as MP3 with metadata

3. **Download Process:**
   - Uses yt-dlp to fetch best audio quality
   - Converts to MP3
   - Embeds thumbnail as cover art
   - Adds metadata (title, artist, etc.)
   - Saves to current directory

## Output Files

Downloaded files are saved in the current directory with the format:
```
[Video Title].mp3
```

Each MP3 includes:
- High-quality audio (best available)
- Embedded album art (video thumbnail)
- ID3 metadata (title, artist, etc.)

## Performance Optimizations

- **Parallel Loading** - Preview starts while fetching metadata
- **Instant Feedback** - No loading screens, progressive UI updates
- **Efficient Search** - Results load incrementally (20 at a time)
- **Background Processing** - Downloads don't block the UI

## Development

### Building

```bash
# Install dependencies
make deps

# Format code
make fmt

# Run linter
make vet

# Build
make build
```

### Project Guidelines

- Follow Go conventions and idioms
- Keep packages focused and modular
- Use Bubbletea's Elm architecture
- Maintain clean separation of concerns:
  - `app/` - Application logic
  - `ui/` - Styling and appearance
  - `youtube/` - External API integration
  - `utils/` - Shared utilities

## Troubleshooting

### "Required tool not found"
Ensure yt-dlp, mpv, and ffmpeg are installed and in your PATH.

### Preview doesn't start
Check that mpv is installed and working:
```bash
mpv --version
```

### Download fails
Verify yt-dlp is up to date:
```bash
yt-dlp --update  # or: brew upgrade yt-dlp
```

### UI looks broken
Ensure your terminal supports:
- 256 colors
- UTF-8 encoding
- Modern terminal emulator (iTerm2, Alacritty, etc.)

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This project is provided as-is for educational and personal use.

## Credits

Built with:
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - YouTube downloader
- [mpv](https://mpv.io/) - Media player
- [ffmpeg](https://ffmpeg.org/) - Audio processing

---

Made with ❤️ using Go and Bubbletea
