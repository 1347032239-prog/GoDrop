package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

func newHTTPHandler(result ScanResult, root *os.Root) http.Handler {

	mux := http.NewServeMux()

	allowedPaths := make(map[string]FileEntry, len(result.Files))

	for _, file := range result.Files {

		allowedPaths[file.Path] = file

	}

	mux.HandleFunc("GET /files", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		if encodeErr := json.NewEncoder(w).Encode(result); encodeErr != nil {

			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)

		}

	})

	mux.HandleFunc("GET /download", func(w http.ResponseWriter, r *http.Request) {

		path := r.URL.Query().Get("path")

		if path == "" || path == "." {

			http.Error(w, "路径非法", http.StatusBadRequest)

			return

		}

		if !filepath.IsLocal(path) {

			http.Error(w, "可能逃逸", http.StatusNotFound)

			return

		}

		path = filepath.Clean(path)

		_, existed := allowedPaths[path]

		if !existed {

			http.Error(w, "路径不在索引集合内", http.StatusNotFound)

			return

		}

		file, err := root.Open(path)

		if err != nil {

			http.Error(w, "文件打开失败", http.StatusInternalServerError)

			return

		}

		defer file.Close()

		info, err := file.Stat()

		if err != nil || !info.Mode().IsRegular() {

			http.Error(w, "Not a regular file", http.StatusNotFound)

			return

		}

		w.Header().Set("X-GoDrop-SHA256", allowedPaths[path].SHA256)
		w.Header().Set("Content-Type", "application/pdf")

		_, err = io.Copy(w, file)
		if err != nil {

			slog.Error("文件流式传输失败", "path", path, "error", err)

		}

	})

	mux.HandleFunc("POST /files", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Allow", "GET")

		http.Error(w, "Not Allowed", http.StatusMethodNotAllowed)

	})

	mux.HandleFunc("GET /unknown", func(w http.ResponseWriter, r *http.Request) {

		http.Error(w, "File not found", http.StatusNotFound)

	})

	return mux

}
