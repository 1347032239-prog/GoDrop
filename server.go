package main

import (
	"encoding/json"
	"net/http"
)

func newHTTPHandler(result ScanResult) http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /files", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		if encodeErr := json.NewEncoder(w).Encode(result); encodeErr != nil {

			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)

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
