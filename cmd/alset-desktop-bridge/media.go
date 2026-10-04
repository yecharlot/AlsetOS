package main

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Media roots: user data + typical removable/network mounts on Linux/TinyCore.
func (b *bridge) mediaRoots() []string {
	roots := []string{b.dataDir}
	for _, r := range []string{"/media", "/mnt", "/home/tc", "/tmp"} {
		if st, err := os.Stat(r); err == nil && st.IsDir() {
			roots = append(roots, r)
		}
	}
	return roots
}

func (b *bridge) safeMediaPath(rel string) (string, error) {
	if rel == "" || rel == "." {
		return b.dataDir, nil
	}
	var candidate string
	if filepath.IsAbs(rel) {
		candidate = rel
	} else {
		candidate = filepath.Join(b.dataDir, rel)
	}
	abs, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	for _, root := range b.mediaRoots() {
		if under(root, abs) {
			return abs, nil
		}
	}
	return "", errAuth("path not in allowed media roots")
}

func (b *bridge) handleMediaRoots(w http.ResponseWriter, r *http.Request) {
	type root struct {
		Path  string `json:"path"`
		Label string `json:"label"`
	}
	var list []root
	for _, p := range b.mediaRoots() {
		list = append(list, root{Path: p, Label: p})
	}
	writeJSON(w, map[string]any{"ok": true, "roots": list})
}

func (b *bridge) handleMediaList(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	filter := strings.ToLower(r.URL.Query().Get("filter")) // audio|video|image|all
	abs, err := b.safeMediaPath(path)
	if err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	type ent struct {
		Name string `json:"name"`
		Dir  bool   `json:"dir"`
		Path string `json:"path"`
		URL  string `json:"url,omitempty"`
	}
	var list []ent
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		full := filepath.Join(abs, name)
		item := ent{Name: name, Dir: e.IsDir(), Path: full}
		if !e.IsDir() {
			if filter != "" && filter != "all" && !mediaMatch(name, filter) {
				continue
			}
			item.URL = "/v1/media/file?path=" + url.QueryEscape(full)
		}
		list = append(list, item)
	}
	writeJSON(w, map[string]any{"ok": true, "path": abs, "entries": list})
}

func mediaMatch(name, filter string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch filter {
	case "audio":
		return ext == ".mp3" || ext == ".ogg" || ext == ".wav" || ext == ".flac" || ext == ".m4a" || ext == ".aac"
	case "video":
		return ext == ".mp4" || ext == ".webm" || ext == ".avi" || ext == ".mkv" || ext == ".mpeg" || ext == ".mpg" || ext == ".mov"
	case "image":
		return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" || ext == ".bmp" || ext == ".svg"
	default:
		return true
	}
}

func (b *bridge) handleMediaFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	abs, err := b.safeMediaPath(path)
	if err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	st, err := os.Stat(abs)
	if err != nil || st.IsDir() {
		http.Error(w, "not found", 404)
		return
	}
	// size limit 200MB for media stream
	if st.Size() > 200<<20 {
		http.Error(w, "file too large", 413)
		return
	}
	http.ServeFile(w, r, abs)
}
