package main

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func spaFileServer(dir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if file, ok := staticFilePath(dir, r.URL.Path); ok {
			http.ServeFile(w, r, file)
			return
		}
		indexPath := filepath.Join(dir, "index.html")
		if info, err := os.Stat(indexPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, indexPath)
			return
		}
		http.NotFound(w, r)
	})
}

func staticFilePath(dir string, requestPath string) (string, bool) {
	cleanPath := strings.TrimPrefix(path.Clean("/"+requestPath), "/")
	if cleanPath == "" || cleanPath == "." {
		return "", false
	}
	file := filepath.Join(dir, filepath.FromSlash(cleanPath))
	info, err := os.Stat(file)
	if err != nil || info.IsDir() {
		return "", false
	}
	return file, true
}
