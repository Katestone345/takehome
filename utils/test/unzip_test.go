package utils_test

import (
	"bytes"
	"compress/gzip"
	"os"
	"testing"

	"example.com/takehome/utils"
)

func TestUnzipFile(t *testing.T) {
	// Create a sample gzip file
	inputData := []byte("This is a test content")
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err := gzipWriter.Write(inputData)
	if err != nil {
		t.Fatalf("Failed to write to gzip writer: %v", err)
	}
	gzipWriter.Close()

	gzipFilePath := "testdata/test_input.gz"
	jsonFilePath := "testdata/test_output.json"

	// Create test directory if it doesn't exist
	if err := os.MkdirAll("testdata", os.ModePerm); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Write the gzip data to a file
	if err := os.WriteFile(gzipFilePath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("Failed to write gzip file: %v", err)
	}

	// Ensure the output file is removed after test
	defer os.Remove(gzipFilePath)
	defer os.Remove(jsonFilePath)

	// Call UnzipFile function
	err = utils.UnzipFile(gzipFilePath, jsonFilePath)
	if err != nil {
		t.Fatalf("UnzipFile returned an error: %v", err)
	}

	// Read the output file and compare content
	outputData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !bytes.Equal(outputData, inputData) {
		t.Errorf("Output file content does not match input data.\nExpected: %s\nGot: %s", inputData, outputData)
	}
}
