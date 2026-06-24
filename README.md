# RockTube

A tiny, self-hosted YouTube clone. **Drop the binary in any folder of videos, run it, and watch them in a YouTube-style web app.** One Go binary, no dependencies, no install.

```
   your video folder/
   ├── rocktube.exe          <- the only thing you need
   ├── movie1.mp4
   ├── vacation.mkv
   └── ...
```

```cmd
rocktube.exe
```

A browser opens at `http://localhost:8090` with your videos.

---

## Features

- **One binary, drop-anywhere** — serves videos from the folder it runs in (or `--folder PATH`)
- **YouTube-style UI** — dark theme, video grid, thumbnails, durations, search, "up next" rail
- **Seek / scrub through videos** — full HTTP range streaming
- **Auto thumbnails & duration** — generated with `ffmpeg` / `ffprobe`, cached on disk
- **View counter** — tracks plays, persisted in a local cache file
- **Like / dislike** — thumbs-up & thumbs-down with per-browser vote state (toggle behaviour, just like YouTube)
- **Comments** — post, list (newest first with relative timestamps), and delete comments per video
- **Subtitles** — drop a `.srt` or `.vtt` next to a video (same name) and it loads automatically
- **Single-file** — the whole web app (HTML/CSS/JS) is embedded in the binary

## Requirements

- The `rocktube` binary (Linux/macOS/Windows)
- **`ffmpeg` + `ffprobe` on your PATH** — only needed for thumbnails & duration display. The app still works without them (you just get placeholder thumbnails).

## Build

```cmd
go build -o rocktube.exe .
```

No external Go modules — pure standard library.

## Usage

```cmd
rocktube                       :: serve videos in the current folder
rocktube --folder "D:\Movies"  :: serve a specific folder
rocktube --port 9000           :: use a different port
rocktube --host 0.0.0.0        :: listen on all interfaces (share on LAN)
rocktube --no-browser          :: don't auto-open a browser
rocktube --version
```

If the default port `8090` is taken, it automatically tries `8091`–`8095`.

### Watching on your phone / TV (LAN)

```cmd
rocktube --host 0.0.0.0
```

Then open `http://<your-pc-name>:8090` or `http://<your-pc-ip>:8090` from any device on the same network.

## Supported formats

`.mp4 .mkv .webm .mov .avi .m4v .wmv .flv .mpg .mpeg .ts .3gp .ogv`

> Browser playback depends on the browser's codec support. `.mp4`/`.webm` play everywhere; formats like `.mkv`/`.avi` may not play in all browsers even though the server streams them fine. For best results use H.264 `.mp4`.

## Cache

RockTube writes a `.rocktube/` folder next to your videos:

```
.rocktube/
├── thumbs/*.jpg    :: generated thumbnails
├── meta/*.json     :: ffprobe duration / resolution, cached
└── views.json      :: view counts
```

Delete `.rocktube/` any time to regenerate everything. It's safe to ignore in version control — this isn't something you commit.

## How it works

| Endpoint            | Purpose                                  |
|---------------------|------------------------------------------|
| `GET /`             | The single-page app (embedded in binary) |
| `GET /api/videos`   | JSON list of videos + metadata           |
| `GET /api/search?q=`| Search by title                          |
| `POST /api/view?id=`| Increment a video's view count           |
| `GET /stream/<id>`  | Video bytes (HTTP range / seek)          |
| `GET /thumb/<id>`   | Thumbnail JPEG                           |
| `GET /subtitle/<id>`| `.vtt` / `.srt` if present               |

## Project layout

```
main.go     entry point: flags, port fallback, opens browser
server.go   scanning, API, range streaming, view tracking
ffmpeg.go   thumbnails, ffprobe metadata, caching, hashing
web.go      the embedded HTML/CSS/JS SPA
```

## Notes & limitations

- **Read-only, single folder.** It serves one flat directory (no subfolders). That keeps it dead simple — the whole point.
- **No auth.** Anyone who can reach the port can watch. Use `--host localhost` (default) to keep it private, or only bind `0.0.0.0` on a trusted LAN.
- **No uploads.** It reads the files already on disk.
