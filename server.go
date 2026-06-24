package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// videoExt lists the file extensions treated as videos (lowercase, with dot).
var videoExt = map[string]bool{
	".mp4": true, ".mkv": true, ".webm": true, ".mov": true, ".avi": true,
	".m4v": true, ".wmv": true, ".flv": true, ".mpg": true, ".mpeg": true,
	".ts": true, ".3gp": true, ".ogv": true,
}

// Server holds the state for the video server.
type Server struct {
	root     string // absolute path to the served folder
	cache    string // absolute path to the .rocktube cache folder
	router   *http.ServeMux
	mu       sync.Mutex
	views    map[string]int64 // videoName -> view count
	lastScan time.Time
	social   *socialStore
}

// NewServer constructs a server serving videos from root.
func NewServer(root string) *Server {
	cache := filepath.Join(root, ".rocktube")
	if err := os.MkdirAll(filepath.Join(cache, "thumbs"), 0o755); err != nil {
		log.Printf("warning: could not create cache dir: %v", err)
	}
	s := &Server{
		root:   root,
		cache:  cache,
		views:  loadViews(cache),
		social: newSocialStore(cache),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/videos", s.handleVideos)
	mux.HandleFunc("/api/folders", s.handleFolders)
	mux.HandleFunc("/api/view", s.handleView)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/comments", s.handleComments)
	mux.HandleFunc("/api/rate", s.handleRate)
	mux.HandleFunc("/thumb/", s.handleThumb)
	mux.HandleFunc("/stream/", s.handleStream)
	mux.HandleFunc("/subtitle/", s.handleSubtitle)
	mux.HandleFunc("/favicon.ico", s.handleFavicon)
	mux.HandleFunc("/", s.handleIndex)
	s.router = mux
}

// ---- scanning --------------------------------------------------------------

func (s *Server) scanVideos(folder string) ([]VideoInfo, error) {
	s.mu.Lock()
	s.lastScan = time.Now()
	s.mu.Unlock()

	var out []VideoInfo
	dir, folder, err := s.resolveFolder(folder)
	if err != nil {
		return nil, err
	}
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		// Skip hidden directories (like .rocktube) and their contents.
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		// If a folder filter is set, only return videos directly in that folder
		// (not in subdirectories), unless folder is "" (show everything).
		if folder != "" && d.IsDir() && path != dir {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if !videoExt[strings.ToLower(filepath.Ext(name))] {
			return nil
		}
		// Compute relative path from root (e.g., "Kids/Toddlers/video.mp4").
		rel, err := filepath.Rel(s.root, path)
		if err != nil {
			return nil
		}
		// Normalise to forward slashes for IDs and API keys.
		rel = filepath.ToSlash(rel)
		info, _ := d.Info()
		out = append(out, s.infoFor(rel, info))
		return nil
	})
	return out, err
}

// VideoInfo is the JSON shape sent to the client.
type VideoInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Path        string `json:"path"` // relative folder, e.g. "Kids/Toddlers"
	Size        int64  `json:"size"`
	Duration    string `json:"duration"`
	DurationSec int    `json:"durationSec"`
	Views       int64  `json:"views"`
	Uploaded    string `json:"uploaded"`
	HasSubs     bool   `json:"hasSubs"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Likes       int    `json:"likes"`
	Dislikes    int    `json:"dislikes"`
	Comments    int    `json:"comments"`
}

func (s *Server) infoFor(relPath string, fi os.FileInfo) VideoInfo {
	id := idFromName(relPath)
	meta := s.loadMeta(relPath)
	// relPath uses forward slashes, so we can split on "/" for the folder.
	dir, file := pathSplit(relPath)
	dir = strings.TrimSuffix(dir, "/") // remove trailing "/"
	v := VideoInfo{
		ID:          id,
		Name:        relPath,
		Title:       strings.TrimSuffix(file, filepath.Ext(file)),
		Path:        dir,
		Duration:    meta.Duration,
		DurationSec: meta.DurationSec,
		Width:       meta.Width,
		Height:      meta.Height,
		HasSubs:     s.hasSubs(relPath),
	}
	if fi != nil {
		v.Size = fi.Size()
		v.Uploaded = humanTime(fi.ModTime())
	}
	s.mu.Lock()
	v.Views = s.views[relPath]
	s.mu.Unlock()
	v.Likes, v.Dislikes = s.social.ratingsFor(relPath)
	v.Comments = len(s.social.commentsFor(relPath))
	return v
}

// idFromName is the stable client-facing ID. It stays as the relative path;
// each caller encodes it for the specific URL shape it is building.
func idFromName(name string) string {
	return name
}

func nameFromID(id string) string {
	n, err := url.PathUnescape(id)
	if err != nil {
		return id
	}
	return n
}

// ---- API handlers ----------------------------------------------------------

func (s *Server) handleVideos(w http.ResponseWriter, r *http.Request) {
	folder := r.URL.Query().Get("folder")
	videos, err := s.scanVideos(folder)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errMap(err))
		return
	}
	// Kick off thumbnail generation for anything missing (non-blocking).
	for _, v := range videos {
		s.ensureThumb(v.Name)
	}
	writeJSON(w, http.StatusOK, map[string]any{"videos": videos})
}

func (s *Server) handleView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errMap(fmt.Errorf("POST required")))
		return
	}
	name := nameFromID(r.URL.Query().Get("id"))
	if !s.isVideo(name) {
		writeJSON(w, http.StatusNotFound, errMap(fmt.Errorf("not found")))
		return
	}
	s.mu.Lock()
	s.views[name]++
	cnt := s.views[name]
	s.mu.Unlock()
	s.saveViews()
	writeJSON(w, http.StatusOK, map[string]any{"views": cnt})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	videos, err := s.scanVideos("")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errMap(err))
		return
	}
	var out []VideoInfo
	for _, v := range videos {
		if q == "" || strings.Contains(strings.ToLower(v.Title), q) {
			out = append(out, v)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"videos": out, "query": q})
}

// FolderNode is the JSON shape for the /api/folders endpoint.
type FolderNode struct {
	Name       string        `json:"name"`
	Path       string        `json:"path"` // relative path from root, "" for root
	Count      int           `json:"count"`
	TotalCount int           `json:"totalCount"`
	Children   []*FolderNode `json:"children,omitempty"`
}

// handleFolders returns the complete folder tree with video counts per folder.
func (s *Server) handleFolders(w http.ResponseWriter, r *http.Request) {
	root := s.buildFolderTree()
	writeJSON(w, http.StatusOK, map[string]any{"folders": root})
}

func (s *Server) buildFolderTree() []*FolderNode {
	// First pass: walk the entire tree and count videos per folder.
	counts := map[string]int{} // relative folder path -> video count
	filepath.WalkDir(s.root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			if d != nil && d.IsDir() && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !videoExt[strings.ToLower(filepath.Ext(d.Name()))] {
			return nil
		}
		rel, err := filepath.Rel(s.root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		dir, _ := pathSplit(rel)
		dir = strings.TrimSuffix(dir, "/")
		counts[dir]++
		return nil
	})

	root := buildNode(filepath.Base(s.root), "", counts, s.root)
	return []*FolderNode{root}
}

func buildNode(name, nodePath string, counts map[string]int, absPath string) *FolderNode {
	node := &FolderNode{
		Name:       name,
		Path:       nodePath,
		Count:      counts[nodePath],
		TotalCount: counts[nodePath],
	}
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return node
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		childPath := filepath.ToSlash(filepath.Join(nodePath, e.Name()))
		child := buildNode(e.Name(), childPath, counts, filepath.Join(absPath, e.Name()))
		node.Children = append(node.Children, child)
		node.TotalCount += child.TotalCount
	}
	return node
}

// ---- comments & ratings ----------------------------------------------------

// clientKey returns the per-client identity used to dedupe votes. If the
// request already carries our cookie it's reused; otherwise a new value is
// minted, set on the response, AND returned so this very request uses the same
// key (otherwise the first vote would be filed under a different identity than
// every subsequent one, breaking toggling).
func (s *Server) clientKey(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie("rt_client"); err == nil && c.Value != "" {
		return c.Value
	}
	v := hashName(r.RemoteAddr + time.Now().String())
	http.SetCookie(w, &http.Cookie{
		Name:     "rt_client",
		Value:    v,
		Path:     "/",
		MaxAge:   10 * 365 * 24 * 3600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return v
}

// handleComments handles GET (list), POST (add), DELETE (remove).
// The video is identified by ?id=<urlencoded name>.
func (s *Server) handleComments(w http.ResponseWriter, r *http.Request) {
	name := nameFromID(r.URL.Query().Get("id"))
	if !s.isVideo(name) {
		writeJSON(w, http.StatusNotFound, errMap(fmt.Errorf("video not found")))
		return
	}

	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{
			"comments": s.social.commentsFor(name),
			"count":    len(s.social.commentsFor(name)),
		})

	case http.MethodPost:
		var body struct {
			Author string `json:"author"`
			Text   string `json:"text"`
		}
		// Limit how much we read so nobody can OOM us with a huge body.
		dec := json.NewDecoder(io.LimitReader(r.Body, 64*1024))
		if err := dec.Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, errMap(fmt.Errorf("invalid JSON: %w", err)))
			return
		}
		c, err := s.social.addComment(name, body.Author, body.Text)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errMap(err))
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"comment": c})

	case http.MethodDelete:
		cid := r.URL.Query().Get("cid")
		if cid == "" {
			writeJSON(w, http.StatusBadRequest, errMap(fmt.Errorf("missing cid")))
			return
		}
		if s.social.deleteComment(name, cid) {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		} else {
			writeJSON(w, http.StatusNotFound, errMap(fmt.Errorf("comment not found")))
		}

	default:
		w.Header().Set("Allow", "GET, POST, DELETE")
		writeJSON(w, http.StatusMethodNotAllowed, errMap(fmt.Errorf("method not allowed")))
	}
}

// handleRate records a thumbs-up / thumbs-down vote.
//
//	POST /api/rate?id=<name>&action=like|dislike|none
//
// Returns the resulting counts and this client's current vote.
func (s *Server) handleRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeJSON(w, http.StatusMethodNotAllowed, errMap(fmt.Errorf("POST required")))
		return
	}
	name := nameFromID(r.URL.Query().Get("id"))
	if !s.isVideo(name) {
		writeJSON(w, http.StatusNotFound, errMap(fmt.Errorf("video not found")))
		return
	}
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "like", "dislike", "none", "":
		// ok
	default:
		writeJSON(w, http.StatusBadRequest, errMap(fmt.Errorf("action must be like, dislike, or none")))
		return
	}
	if action == "" {
		action = "none"
	}
	key := s.clientKey(w, r)
	likes, dislikes, myVote := s.social.setRating(name, key, action)
	writeJSON(w, http.StatusOK, map[string]any{
		"likes":    likes,
		"dislikes": dislikes,
		"myVote":   myVote,
	})
}

// ---- streaming & media -----------------------------------------------------

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	name := nameFromID(strings.TrimPrefix(r.URL.Path, "/stream/"))
	if !s.isVideo(name) {
		http.NotFound(w, r)
		return
	}
	path := filepath.Join(s.root, name)
	f, err := os.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		http.Error(w, "stat error", http.StatusInternalServerError)
		return
	}

	ctype := mime.TypeByExtension(strings.ToLower(filepath.Ext(name)))
	if ctype == "" {
		ctype = "video/mp4"
	}
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// http.ServeContent handles Range requests, HEAD, and 304 on its own,
	// so seeking "just works" in the <video> element.
	http.ServeContent(w, r, name, fi.ModTime(), f)
}

func (s *Server) handleThumb(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/thumb/")
	name := nameFromID(id)
	if !s.isVideo(name) {
		http.NotFound(w, r)
		return
	}
	thumbPath := filepath.Join(s.cache, "thumbs", hashName(name)+".jpg")
	if _, err := os.Stat(thumbPath); err != nil {
		if err := generateThumb(filepath.Join(s.root, name), thumbPath); err != nil {
			servePlaceholder(w)
			return
		}
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, thumbPath)
}

func (s *Server) handleSubtitle(w http.ResponseWriter, r *http.Request) {
	name := nameFromID(strings.TrimPrefix(r.URL.Path, "/subtitle/"))
	if !s.isVideo(name) {
		http.NotFound(w, r)
		return
	}
	base := strings.TrimSuffix(name, filepath.Ext(name))
	for _, sub := range []string{".vtt", ".srt", ".VTT", ".SRT"} {
		candidate := filepath.Join(s.root, base+sub)
		if _, err := os.Stat(candidate); err == nil {
			ctype := "text/vtt; charset=utf-8"
			if strings.EqualFold(sub, ".srt") {
				ctype = "application/x-subrip; charset=utf-8"
			}
			w.Header().Set("Content-Type", ctype)
			http.ServeFile(w, r, candidate)
			return
		}
	}
	http.NotFound(w, r)
}

func (s *Server) handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	fmt.Fprint(w, faviconSVG)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	fmt.Fprint(w, indexHTML)
}

// ---- misc helpers ----------------------------------------------------------

func (s *Server) isVideo(name string) bool {
	if name == "" {
		return false
	}
	resolved, err := s.resolveVideoPath(name)
	if err != nil {
		return false
	}
	st, err := os.Stat(resolved)
	if err != nil || st.IsDir() {
		return false
	}
	return videoExt[strings.ToLower(filepath.Ext(name))]
}

func (s *Server) resolveVideoPath(name string) (string, error) {
	name = filepath.ToSlash(strings.TrimSpace(name))
	if name == "" || strings.HasPrefix(name, "/") || filepath.IsAbs(name) {
		return "", fmt.Errorf("invalid video path")
	}
	resolved := filepath.Clean(filepath.Join(s.root, filepath.FromSlash(name)))
	rel, err := filepath.Rel(s.root, resolved)
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", fmt.Errorf("video path escapes root")
	}
	return resolved, nil
}

func (s *Server) resolveFolder(folder string) (absPath, relPath string, err error) {
	folder = strings.TrimSpace(filepath.ToSlash(folder))
	if folder == "/" {
		folder = ""
	}
	if folder == "" {
		return s.root, "", nil
	}
	if strings.HasPrefix(folder, "/") || filepath.IsAbs(folder) {
		return "", "", fmt.Errorf("invalid folder")
	}
	resolved := filepath.Clean(filepath.Join(s.root, filepath.FromSlash(folder)))
	rel, err := filepath.Rel(s.root, resolved)
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", "", fmt.Errorf("folder escapes root")
	}
	st, err := os.Stat(resolved)
	if err != nil {
		return "", "", err
	}
	if !st.IsDir() {
		return "", "", fmt.Errorf("not a folder")
	}
	return resolved, filepath.ToSlash(rel), nil
}

func (s *Server) hasSubs(relPath string) bool {
	// relPath uses forward slashes; convert to OS path for file access.
	osPath := filepath.FromSlash(relPath)
	base := strings.TrimSuffix(osPath, filepath.Ext(osPath))
	for _, sub := range []string{".vtt", ".srt", ".VTT", ".SRT"} {
		if _, err := os.Stat(filepath.Join(s.root, base+sub)); err == nil {
			return true
		}
	}
	return false
}

func pathSplit(relPath string) (dir, file string) {
	i := strings.LastIndex(relPath, "/")
	if i < 0 {
		return "", relPath
	}
	return relPath[:i+1], relPath[i+1:]
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func errMap(err error) map[string]any {
	return map[string]any{"error": err.Error()}
}

func servePlaceholder(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=300")
	fmt.Fprint(w, placeholderSVG)
}

func humanTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(d.Minutes())+1)
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours())+1)
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24)+1)
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%d months ago", int(d.Hours()/24/30)+1)
	default:
		return fmt.Sprintf("%d years ago", int(d.Hours()/24/365)+1)
	}
}

// humanBytes is exposed to the frontend indirectly via size formatting.
func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

var _ = humanBytes // used by templates if extended
