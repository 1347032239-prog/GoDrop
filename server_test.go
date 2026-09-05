//go:build ignore

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestNewHTTPHandler(t *testing.T) {
	t.Run("GET /files returns an index", func(t *testing.T) {
		want := ScanResult{
			Files: []FileEntry{
				{Path: "a.txt", Size: 3},
				{Path: "empty.txt", Size: 0},
				{Path: "nested/b.txt", Size: 5},
			},
			TotalSize: 8,
		}

		req := httptest.NewRequest(http.MethodGet, "/files", nil)
		recorder := httptest.NewRecorder()
		newHTTPHandler(want).ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("状态码错误: 期望 %d，实际 %d", http.StatusOK, recorder.Code)
		}
		if contentType := recorder.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
			t.Errorf("Content-Type 错误: 期望 application/json，实际 %q", contentType)
		}

		var got ScanResult
		if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
			t.Fatalf("响应正文不是有效 JSON: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("响应索引错误: 期望 %+v，实际 %+v", want, got)
		}
	})

	t.Run("GET /files preserves an empty file list", func(t *testing.T) {
		want := ScanResult{
			Files: make([]FileEntry, 0),
		}

		req := httptest.NewRequest(http.MethodGet, "/files", nil)
		recorder := httptest.NewRecorder()
		newHTTPHandler(want).ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("状态码错误: 期望 %d，实际 %d", http.StatusOK, recorder.Code)
		}

		var got ScanResult
		if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
			t.Fatalf("响应正文不是有效 JSON: %v", err)
		}
		if got.Files == nil {
			t.Error("空文件列表应编码为 []，不应为 null")
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("响应索引错误: 期望 %+v，实际 %+v", want, got)
		}
	})

	t.Run("unknown path returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
		recorder := httptest.NewRecorder()
		newHTTPHandler(ScanResult{}).ServeHTTP(recorder, req)

		if recorder.Code != http.StatusNotFound {
			t.Errorf("状态码错误: 期望 %d，实际 %d", http.StatusNotFound, recorder.Code)
		}
	})

	t.Run("POST /files returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/files", nil)
		recorder := httptest.NewRecorder()
		newHTTPHandler(ScanResult{}).ServeHTTP(recorder, req)

		if recorder.Code != http.StatusMethodNotAllowed {
			t.Errorf("状态码错误: 期望 %d，实际 %d", http.StatusMethodNotAllowed, recorder.Code)
		}
		if allow := recorder.Header().Get("Allow"); !strings.Contains(allow, http.MethodGet) {
			t.Errorf("Allow 响应头应包含 GET，实际为 %q", allow)
		}
	})
}
