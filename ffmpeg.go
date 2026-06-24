package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FileMeta is the cached metadata extracted via ffprobe.
type FileMeta struct {
	Duration    string `json:"duration"`
	DurationSec int    `json:"durationSec"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	GeneratedAt int64  `json:"generatedAt"`
}

var (
	thumbInFlight sync.Map // name -> struct{}{} to dedupe thumbnail generation
	ffmpegOnce    sync.Once
	ffmpegPath    string
	ffmpegErr     error

	ffprobeOnce sync.Once
	ffprobePath string
	ffprobeErr  error

	thumbWork = make(chan struct{}, 2)
)

// loadMeta returns cached metadata for name, or an empty FileMeta if none.
// Cache generation is controlled by the explicit library cache job.
func (s *Server) loadMeta(name string) FileMeta {
	data, err := os.ReadFile(s.metaPath(name))
	if err != nil {
		return FileMeta{}
	}
	var m FileMeta
	if json.Unmarshal(data, &m) != nil {
		return FileMeta{}
	}
	return m
}

func (s *Server) saveMeta(name string, m FileMeta) {
	dir := filepath.Join(s.cache, "meta")
	_ = os.MkdirAll(dir, 0o755)
	m.GeneratedAt = time.Now().Unix()
	data, _ := json.Marshal(m)
	_ = os.WriteFile(filepath.Join(dir, hashName(name)+".json"), data, 0o644)
}

// ensureThumb generates the thumbnail in the background if it doesn't exist.
// It dedupes concurrent calls for the same file via thumbInFlight.
func (s *Server) ensureThumb(name string) {
	if _, err := os.Stat(s.thumbPath(name)); err == nil {
		return // already done
	}
	if _, loaded := thumbInFlight.LoadOrStore(name, struct{}{}); loaded {
		return // someone else is on it
	}
	go func() {
		defer thumbInFlight.Delete(name)
		thumbWork <- struct{}{}
		defer func() { <-thumbWork }()
		_ = generateThumb(filepath.Join(s.root, filepath.FromSlash(name)), s.thumbPath(name))
	}()
}

// generateThumb seeks to ~20% of the movie and grabs one JPEG frame.
func generateThumb(src, dst string) error {
	ffmpeg, err := ffmpegExecutable()
	if err != nil {
		return err
	}
	// Read duration first so we can pick an interesting frame that isn't a
	// black intro. Fall back to a fixed offset if probing fails.
	seek := "00:00:03"
	if m, err := probeMeta(src); err == nil && m.DurationSec > 0 {
		pos := m.DurationSec * 20 / 100
		if pos < 1 {
			pos = 1
		}
		seek = strconv.Itoa(pos)
	}
	// tmp file so a half-written image is never served. Keep the .jpg suffix so
	// ffmpeg can infer the output format (a bare .tmp makes it error out).
	tmp := dst + ".generating.jpg"
	cmd := exec.Command(ffmpeg,
		"-ss", seek,
		"-i", src,
		"-frames:v", "1",
		"-vf", "scale=640:-2",
		"-q:v", "4",
		"-y",
		tmp,
	)
	if err := cmd.Run(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dst)
}

// probeMeta runs ffprobe to get duration + dimensions.
func probeMeta(path string) (FileMeta, error) {
	ffprobe, err := ffprobeExecutable()
	if err != nil {
		return FileMeta{}, err
	}
	out, err := exec.Command(ffprobe,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height:format=duration",
		"-of", "default=noprint_wrappers=1",
		path,
	).Output()
	if err != nil {
		return FileMeta{}, err
	}
	m := FileMeta{}
	for _, line := range strings.Split(string(out), "\n") {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(k) {
		case "width":
			m.Width, _ = strconv.Atoi(strings.TrimSpace(v))
		case "height":
			m.Height, _ = strconv.Atoi(strings.TrimSpace(v))
		case "duration":
			f, _ := strconv.ParseFloat(strings.TrimSpace(v), 64)
			m.DurationSec = int(f)
			m.Duration = formatDuration(m.DurationSec)
		}
	}
	if m.DurationSec == 0 {
		m.Duration = ""
	}
	return m, nil
}

func ffmpegExecutable() (string, error) {
	ffmpegOnce.Do(func() {
		ffmpegPath, ffmpegErr = exec.LookPath("ffmpeg")
		if ffmpegErr != nil {
			ffmpegErr = fmt.Errorf("ffmpeg not found in PATH")
		}
	})
	return ffmpegPath, ffmpegErr
}

func ffprobeExecutable() (string, error) {
	ffprobeOnce.Do(func() {
		ffprobePath, ffprobeErr = exec.LookPath("ffprobe")
		if ffprobeErr != nil {
			ffprobeErr = fmt.Errorf("ffprobe not found in PATH")
		}
	})
	return ffprobePath, ffprobeErr
}

func formatDuration(sec int) string {
	if sec <= 0 {
		return ""
	}
	h := sec / 3600
	mn := (sec % 3600) / 60
	ss := sec % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, mn, ss)
	}
	return fmt.Sprintf("%d:%02d", mn, ss)
}

// ---- view-count persistence ------------------------------------------------

func loadViews(cache string) map[string]int64 {
	path := filepath.Join(cache, "views.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]int64{}
	}
	var v map[string]int64
	if json.Unmarshal(data, &v) != nil {
		return map[string]int64{}
	}
	return v
}

func (s *Server) saveViews() {
	s.mu.Lock()
	data, _ := json.Marshal(s.views)
	s.mu.Unlock()
	_ = os.WriteFile(filepath.Join(s.cache, "views.json"), data, 0o644)
}

// ---- hashing for stable cache filenames ------------------------------------

func hashName(name string) string {
	// A short, deterministic, filesystem-safe id. Not cryptographic — just
	// meant to map an arbitrary filename to one cache file. FNV-1a is enough.
	const (
		offset uint64 = 1469598103934665603
		prime  uint64 = 1099511628211
	)
	h := offset
	for i := 0; i < len(name); i++ {
		h ^= uint64(name[i])
		h *= prime
	}
	return strconv.FormatUint(h, 36)
}
