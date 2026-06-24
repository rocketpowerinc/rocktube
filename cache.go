package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type CacheStatus struct {
	Running     bool   `json:"running"`
	Force       bool   `json:"force"`
	Scanning    bool   `json:"scanning"`
	Total       int    `json:"total"`
	Done        int    `json:"done"`
	MetaCached  int    `json:"metaCached"`
	ThumbCached int    `json:"thumbCached"`
	Errors      int    `json:"errors"`
	Current     string `json:"current"`
	Message     string `json:"message"`
	StartedAt   int64  `json:"startedAt"`
	FinishedAt  int64  `json:"finishedAt"`
}

func (s *Server) handleCacheStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errMap(fmt.Errorf("GET required")))
		return
	}
	writeJSON(w, http.StatusOK, s.cacheStatus())
}

func (s *Server) handleCacheStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errMap(fmt.Errorf("POST required")))
		return
	}
	force := r.URL.Query().Get("force") == "1"
	status := s.startCacheBuild(force)
	writeJSON(w, http.StatusAccepted, status)
}

func (s *Server) cacheStatus() CacheStatus {
	s.cacheMu.Lock()
	status := s.cacheRun
	s.cacheMu.Unlock()
	if !status.Running && status.Total == 0 {
		status = s.cacheSummary()
	}
	return status
}

func (s *Server) startCacheBuild(force bool) CacheStatus {
	s.cacheMu.Lock()
	if s.cacheRun.Running {
		status := s.cacheRun
		s.cacheMu.Unlock()
		return status
	}
	s.cacheRun = CacheStatus{
		Running:   true,
		Force:     force,
		Scanning:  true,
		Message:   "Scanning library",
		StartedAt: time.Now().Unix(),
	}
	status := s.cacheRun
	s.cacheMu.Unlock()

	go s.runCacheBuild(force)
	return status
}

func (s *Server) runCacheBuild(force bool) {
	names, err := s.listVideoNames()
	if err != nil {
		s.updateCacheStatus(func(st *CacheStatus) {
			st.Running = false
			st.Scanning = false
			st.Message = err.Error()
			st.Errors++
			st.FinishedAt = time.Now().Unix()
		})
		return
	}

	metaCached, thumbCached := 0, 0
	if !force {
		for _, name := range names {
			if fileExists(s.metaPath(name)) {
				metaCached++
			}
			if fileExists(s.thumbPath(name)) {
				thumbCached++
			}
		}
	}

	s.updateCacheStatus(func(st *CacheStatus) {
		st.Scanning = false
		st.Total = len(names)
		st.MetaCached = metaCached
		st.ThumbCached = thumbCached
		st.Message = "Building cache"
	})

	jobs := make(chan string)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for name := range jobs {
				s.cacheOne(name, force)
			}
		}()
	}
	for _, name := range names {
		jobs <- name
	}
	close(jobs)
	wg.Wait()

	s.updateCacheStatus(func(st *CacheStatus) {
		st.Running = false
		st.Current = ""
		st.Message = "Cache ready"
		st.FinishedAt = time.Now().Unix()
	})
}

func (s *Server) cacheOne(name string, force bool) {
	s.updateCacheStatus(func(st *CacheStatus) {
		st.Current = name
	})

	if force || !fileExists(s.metaPath(name)) {
		m, err := probeMeta(filepath.Join(s.root, filepath.FromSlash(name)))
		if err == nil {
			s.saveMeta(name, m)
			s.updateCacheStatus(func(st *CacheStatus) {
				st.MetaCached++
			})
		} else {
			s.updateCacheStatus(func(st *CacheStatus) {
				st.Errors++
			})
		}
	}

	if force || !fileExists(s.thumbPath(name)) {
		err := generateThumb(filepath.Join(s.root, filepath.FromSlash(name)), s.thumbPath(name))
		if err == nil {
			s.updateCacheStatus(func(st *CacheStatus) {
				st.ThumbCached++
			})
		} else {
			s.updateCacheStatus(func(st *CacheStatus) {
				st.Errors++
			})
		}
	}

	s.updateCacheStatus(func(st *CacheStatus) {
		st.Done++
	})
}

func (s *Server) updateCacheStatus(fn func(*CacheStatus)) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	fn(&s.cacheRun)
}

func (s *Server) cacheSummary() CacheStatus {
	names, err := s.listVideoNames()
	status := CacheStatus{
		Message: "Cache status",
		Total:   len(names),
	}
	if err != nil {
		status.Message = err.Error()
		status.Errors = 1
		return status
	}
	for _, name := range names {
		if fileExists(s.metaPath(name)) {
			status.MetaCached++
		}
		if fileExists(s.thumbPath(name)) {
			status.ThumbCached++
		}
	}
	status.Done = status.Total
	if status.Total == 0 {
		status.Message = "No videos found"
	} else if status.MetaCached == status.Total && status.ThumbCached == status.Total {
		status.Message = "Cache ready"
	} else {
		status.Message = "Cache incomplete"
	}
	return status
}

func (s *Server) listVideoNames() ([]string, error) {
	var names []string
	err := filepath.WalkDir(s.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
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
		names = append(names, filepath.ToSlash(rel))
		return nil
	})
	return names, err
}

func (s *Server) metaPath(name string) string {
	return filepath.Join(s.cache, "meta", hashName(name)+".json")
}

func (s *Server) thumbPath(name string) string {
	return filepath.Join(s.cache, "thumbs", hashName(name)+".jpg")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
