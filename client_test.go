package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestFetchIndex(t *testing.T) {
	t.Run("fetches an index with GET", func(t *testing.T) {
		want := ScanResult{
			Files: []FileEntry{
				{Path: "a.txt", Size: 3},
				{Path: "empty.txt", Size: 0},
				{Path: "nested/b.txt", Size: 5},
			},
			TotalSize: 8,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("request method: want %s, got %s", http.MethodGet, r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(want); err != nil {
				t.Errorf("encode response: %v", err)
			}
		}))
		defer server.Close()

		got, err := fetchIndex(server.URL)
		if err != nil {
			t.Fatalf("fetchIndex returned an unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("result: want %+v, got %+v", want, got)
		}
	})

	t.Run("rejects a non-200 response before decoding", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		}))
		defer server.Close()

		_, err := fetchIndex(server.URL)
		if err == nil {
			t.Fatal("fetchIndex should reject a non-200 response")
		}
		if !strings.Contains(err.Error(), "404") {
			t.Errorf("error should include status 404, got %q", err)
		}
	})

	t.Run("rejects malformed JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"files":`))
		}))
		defer server.Close()

		if _, err := fetchIndex(server.URL); err == nil {
			t.Fatal("fetchIndex should reject malformed JSON")
		}
	})

	t.Run("returns a network error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		serverURL := server.URL
		server.Close()

		if _, err := fetchIndex(serverURL); err == nil {
			t.Fatal("fetchIndex should return an error when the server is unavailable")
		}
	})
}
