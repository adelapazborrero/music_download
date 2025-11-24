dlmp3 () {
	for cmd in yt-dlp fzf ffmpeg jq numfmt mpv
	do
		if ! command -v "$cmd" > /dev/null 2>&1
		then
			echo "❌ Required tool '$cmd' is not installed. Please install it first."
			return 1
		fi
	done
	if [[ $# -eq 0 ]]
	then
		echo "❗️ Usage: $0 <search terms>"
		return 1
	fi
	SEARCH="$*"
	echo "🔎 Searching for: $SEARCH"
	RESULTS=$(yt-dlp "ytsearch20:$SEARCH" --flat-playlist --print "%(title)s | %(id)s")
	if [[ -z "$RESULTS" ]]
	then
		echo "❌ No results found."
		return 1
	fi
	CHOICE=$(echo "$RESULTS" | fzf --prompt="🎵 Choose a video: " --height=30 --border)
	if [[ -z "$CHOICE" ]]
	then
		echo "❌ No selection made."
		return 1
	fi
	VIDEO_ID=$(echo "$CHOICE" | awk -F ' | ' '{print $NF}')
	URL="https://www.youtube.com/watch?v=$VIDEO_ID"
	META=$(yt-dlp -j "$URL")
	TITLE=$(echo "$META" | jq -r '.title')
	CHANNEL=$(echo "$META" | jq -r '.channel')
	DURATION=$(echo "$META" | jq -r '.duration' | awk '{printf "%d:%02d", $1/60, $1%60}')
	VIEWS=$(echo "$META" | jq -r '.view_count' | numfmt --grouping)
	echo ""
	echo "📄 Video details:"
	echo "   ▶️ Title:    $TITLE"
	echo "   👤 Channel:  $CHANNEL"
	echo "   ⏱  Duration: $DURATION"
	echo "   👁️ Views:    $VIEWS"
	echo ""
	echo "What would you like to do?"
	select OPTION in "Preview" "Download" "Cancel"
	do
	case $OPTION in
			(Preview) echo "🎧 Previewing..."
				mpv --no-video --ytdl-format=bestaudio "$URL" ;;
			(Download) break ;;
			(Cancel) echo "❌ Cancelled."
				return 0 ;;
		esac
	done
	echo "⬇️ Downloading high-quality MP3 with embedded cover art..."
	yt-dlp -f bestaudio --extract-audio --audio-format mp3 --audio-quality 0 --embed-thumbnail --add-metadata -o "%(title)s.%(ext)s" "$URL"
	echo "✅ Download complete!"
}
