package utils

import (
	"compress/gzip"
	"io"
	"os"
)

// UnzipFile unzips the specified gzip file to the specified output filepath.
func UnzipFile(gzipPath, jsonPath string) error {
	file, err := os.Open(gzipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	outFile, err := os.Create(jsonPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Use a buffer to speed up the copying process
	buf := make([]byte, 32*1024) // 32KB buffer
	_, err = io.CopyBuffer(outFile, gzipReader, buf)
	return err
}
