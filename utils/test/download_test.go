package utils_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"example.com/takehome/utils"
)

func TestDownloadFile(t *testing.T) {
	// Create a temporary server to serve a test file
	fileContent := "This is a test file for download"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			// Simulate range requests
			var start, end int
			fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
			if end == 0 {
				end = len(fileContent) - 1
			}
			if start < 0 || start > len(fileContent)-1 {
				start = 0
			}
			if end < 0 || end > len(fileContent)-1 {
				end = len(fileContent) - 1
			}
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(fileContent)))
			w.WriteHeader(http.StatusPartialContent)
			w.Write([]byte(fileContent[start : end+1]))
		} else {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileContent)))
			w.Write([]byte(fileContent))
		}
	}))
	defer server.Close()

	// Temporary file path
	tmpFilePath := "testdata/test_download.txt"
	defer os.Remove(tmpFilePath)

	// Call DownloadFile function
	err := utils.DownloadFile(server.URL, tmpFilePath)
	if err != nil {
		t.Fatalf("DownloadFile returned an error: %v", err)
	}

	// Read the downloaded file and compare the content
	downloadedContent, err := os.ReadFile(tmpFilePath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(downloadedContent) != fileContent {
		t.Errorf("Downloaded content does not match expected content.\nExpected: %s\nGot: %s", fileContent, downloadedContent)
	}
}
