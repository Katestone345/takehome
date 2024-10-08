package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"example.com/takehome/utils"
)

func main() {
	url := "https://antm-pt-prod-dataz-nogbd-nophi-us-east1.s3.amazonaws.com/anthem/2024-08-01_anthem_index.json.gz"
	gzipPath := "index.json.gz"
	jsonPath := "index.json"
	outputFilePath := "anthem_ny_ppo_urls.txt"

	// Check if the gzip file already exists
	if _, err := os.Stat(gzipPath); os.IsNotExist(err) {
		// Download the file using parallel downloads
		fmt.Println("Downloading file...")
		if err := utils.DownloadFile(url, gzipPath); err != nil {
			log.Fatalf("Error downloading file: %v", err)
		}
	} else {
		fmt.Println("Gzip file already exists, skipping download.")
	}

	// Check if the JSON file already exists
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		// Unzip the file
		fmt.Println("Unzipping file...")
		if err := utils.UnzipFile(gzipPath, jsonPath); err != nil {
			log.Fatalf("Error unzipping file: %v", err)
		}
	} else {
		fmt.Println("JSON file already exists, skipping unzip.")
	}

	// Open the unzipped JSON file for reading
	jsonFile, err := os.Open(jsonPath)
	if err != nil {
		log.Fatalf("Error opening JSON file: %v", err)
	}
	defer jsonFile.Close()

	done := make(chan struct{})
	urlsChan := make(chan string, 100) // Buffered channel to hold URLs

	var parseErr error
	var wg sync.WaitGroup

	// Concurrently parse the JSON and send URLs to channel
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Parsing and filtering JSON...")
		parseErr = utils.ParseAndFilterJSONFromReader(jsonFile, urlsChan, done)
		close(urlsChan) // Close the channel once parsing is done
	}()

	// Wait for parsing to complete
	wg.Wait()

	if parseErr != nil {
		log.Fatalf("Error parsing JSON: %v", parseErr)
	}

	// Save URLs to file
	outFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer outFile.Close()

	fmt.Println("Saving URLs to file...")
	for url := range urlsChan {
		outFile.WriteString(url + "\n")
	}

	fmt.Printf("URLs for Anthem PPO in New York State saved to %s\n", outputFilePath)
}
