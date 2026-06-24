package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// social.go adds the community layer on top of the read-only video server:
// per-video comments and a thumbs-up / thumbs-down rating. Everything is
// persisted as JSON inside the .rocktube cache folder so it survives restarts.

// Comment is a single stored comment.
type Comment struct {
	ID        string `json:"id"`
	VideoName string `json:"video"` // the file name this comment belongs to
	Author    string `json:"author"`
	Text      string `json:"text"`
	CreatedAt int64  `json:"createdAt"` // unix seconds
}

// socialStore owns the in-memory comment + rating state with a mutex.
type socialStore struct {
	mu        sync.Mutex
	cache     string
	comments  map[string][]Comment         // videoName -> comments (oldest first)
	likes     map[string]int               // videoName -> count
	dislikes  map[string]int               // videoName -> count
	userVotes map[string]map[string]string // videoName -> clientKey -> "like"|"dislike"
}

func newSocialStore(cache string) *socialStore {
	s := &socialStore{
		cache:     cache,
		comments:  map[string][]Comment{},
		likes:     map[string]int{},
		dislikes:  map[string]int{},
		userVotes: map[string]map[string]string{},
	}
	s.load()
	return s
}

// ---- persistence -----------------------------------------------------------

func (s *socialStore) load() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if data, err := os.ReadFile(filepath.Join(s.cache, "comments.json")); err == nil {
		_ = json.Unmarshal(data, &s.comments)
	}
	if data, err := os.ReadFile(filepath.Join(s.cache, "ratings.json")); err == nil {
		var both struct {
			Likes     map[string]int               `json:"likes"`
			Dislikes  map[string]int               `json:"dislikes"`
			UserVotes map[string]map[string]string `json:"userVotes"`
		}
		if json.Unmarshal(data, &both) == nil {
			if both.Likes != nil {
				s.likes = both.Likes
			}
			if both.Dislikes != nil {
				s.dislikes = both.Dislikes
			}
			if both.UserVotes != nil {
				s.userVotes = both.UserVotes
			}
		}
	}
}

func (s *socialStore) save() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c, err := json.Marshal(s.comments); err == nil {
		_ = os.WriteFile(filepath.Join(s.cache, "comments.json"), c, 0o644)
	}
	both := struct {
		Likes     map[string]int               `json:"likes"`
		Dislikes  map[string]int               `json:"dislikes"`
		UserVotes map[string]map[string]string `json:"userVotes"`
	}{s.likes, s.dislikes, s.userVotes}
	if r, err := json.Marshal(both); err == nil {
		_ = os.WriteFile(filepath.Join(s.cache, "ratings.json"), r, 0o644)
	}
}

// ---- comments --------------------------------------------------------------

// addComment appends a comment and returns it. It validates + sanitizes input.
func (s *socialStore) addComment(videoName, author, text string) (Comment, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return Comment{}, fmt.Errorf("comment cannot be empty")
	}
	if len([]rune(text)) > 2000 {
		return Comment{}, fmt.Errorf("comment too long (max 2000 characters)")
	}
	author = strings.TrimSpace(author)
	if author == "" {
		author = "Anonymous"
	}
	if len([]rune(author)) > 40 {
		author = author[:40]
	}
	c := Comment{
		ID:        newID(),
		VideoName: videoName,
		Author:    author,
		Text:      text,
		CreatedAt: time.Now().Unix(),
	}
	s.mu.Lock()
	s.comments[videoName] = append(s.comments[videoName], c)
	s.mu.Unlock()
	s.save()
	return c, nil
}

func (s *socialStore) deleteComment(videoName, id string) bool {
	s.mu.Lock()
	list := s.comments[videoName]
	found := false
	for i, c := range list {
		if c.ID == id {
			s.comments[videoName] = append(list[:i], list[i+1:]...)
			found = true
			break
		}
	}
	s.mu.Unlock()
	if found {
		s.save()
	}
	return found
}

func (s *socialStore) commentsFor(videoName string) []Comment {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Comment, len(s.comments[videoName]))
	copy(out, s.comments[videoName])
	return out
}

// ---- ratings ---------------------------------------------------------------

// setRating records a vote. action is "like", "dislike", or "none".
// Each client can have one vote per video; changing the vote moves the count,
// and voting the same way again removes it (toggle behaviour).
func (s *socialStore) setRating(videoName, clientKey, action string) (likes, dislikes int, myVote string) {
	s.mu.Lock()
	prev := ""
	if m, ok := s.userVotes[videoName]; ok {
		prev = m[clientKey]
	}
	// undo previous vote
	switch prev {
	case "like":
		if s.likes[videoName] > 0 {
			s.likes[videoName]--
		}
	case "dislike":
		if s.dislikes[videoName] > 0 {
			s.dislikes[videoName]--
		}
	}
	// toggle: voting the same way again clears it
	myVote = "none"
	effective := action
	if prev == action {
		effective = "none"
	}
	switch effective {
	case "like":
		s.likes[videoName]++
		myVote = "like"
	case "dislike":
		s.dislikes[videoName]++
		myVote = "dislike"
	}
	if s.userVotes[videoName] == nil {
		s.userVotes[videoName] = map[string]string{}
	}
	if effective == "none" {
		delete(s.userVotes[videoName], clientKey)
	} else {
		s.userVotes[videoName][clientKey] = effective
	}
	likes = s.likes[videoName]
	dislikes = s.dislikes[videoName]
	s.mu.Unlock()
	s.save()
	return likes, dislikes, myVote
}

func (s *socialStore) ratingsFor(videoName string) (likes, dislikes int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.likes[videoName], s.dislikes[videoName]
}

func (s *socialStore) myVote(videoName, clientKey string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m, ok := s.userVotes[videoName]; ok {
		return m[clientKey]
	}
	return ""
}

// newID returns a short unique id for a comment (FNV over nanosecond time).
func newID() string {
	return hashName(fmt.Sprintf("%d", time.Now().UnixNano()))
}
