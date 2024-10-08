package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
)

const numParts = 8 // Number of parts to split the download into

// DownloadFile downloads the file from the specified URL and saves it to the specified filepath using parallel downloads.
func DownloadFile(url, filepath string) error {
	// Get the size of the file
	resp, err := http.Head(url)
	if err != nil {
		return fmt.Errorf("failed to get file size: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status: %s", resp.Status)
	}
	size, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse content length: %v", err)
	}

	// Create a file to save the downloaded content
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	var wg sync.WaitGroup
	partSize := size / numParts

	for i := 0; i < numParts; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			start := int64(i) * partSize
			end := start + partSize - 1
			if i == numParts-1 {
				end = size - 1
			}
			if err := downloadPart(url, start, end, out); err != nil {
				fmt.Printf("failed to download part %d: %v\n", i, err)
			}
		}(i)
	}

	wg.Wait()

	return nil
}

// downloadPart downloads a part of the file and writes it to the specified file.
func downloadPart(url string, start, end int64, out *os.File) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download part: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("server returned non-206 status: %s", resp.Status)
	}

	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, err := out.WriteAt(buf[:n], start); err != nil {
				return fmt.Errorf("failed to write part: %v", err)
			}
			start += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read part: %v", err)
		}
	}

	return nil
}
