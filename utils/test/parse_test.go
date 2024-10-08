package utils_test

import (
	"strings"
	"testing"

	"example.com/takehome/utils"
)

func TestParseAndFilterJSONFromReader(t *testing.T) {
	// Mock JSON data
	jsonData := `
	{
		"reporting_entity_name": "Anthem NY",
		"reporting_entity_type": "Insurance Company",
		"reporting_structure": [
			{
				"reporting_plans": [
					{
						"plan_name": "Anthem PPO",
						"plan_id_type": "HIOS",
						"plan_id": "12345",
						"plan_market_type": "Individual"
					}
				],
				"in_network_files": [
					{
						"description": "In-Network Providers",
						"location": "https://example.com/in_network_file.json"
					}
				],
				"allowed_amount_file": {
					"description": "Allowed Amounts",
					"location": "https://example.com/allowed_amounts.json"
				}
			}
		]
	}`

	// Create a reader from the mock JSON data
	reader := strings.NewReader(jsonData)

	// Create a channel to receive URLs
	urlsChan := make(chan string, 10)
	// Create a done channel to simulate early termination
	done := make(chan struct{})

	// Run the function in a goroutine
	go func() {
		if err := utils.ParseAndFilterJSONFromReader(reader, urlsChan, done); err != nil {
			t.Fatalf("ParseAndFilterJSONFromReader returned an error: %v", err)
		}
		close(urlsChan)
	}()

	// Collect the results
	var urls []string
	for url := range urlsChan {
		urls = append(urls, url)
	}

	// Check if the correct URLs were extracted
	expectedURL := "https://example.com/in_network_file.json"
	if len(urls) != 1 || urls[0] != expectedURL {
		t.Errorf("Expected URL %s, but got %v", expectedURL, urls)
	}

	// Test the early termination scenario
	done = make(chan struct{})           // reset done channel
	reader = strings.NewReader(jsonData) // reset reader

	go func() {
		close(done) // Signal to stop processing immediately
	}()

	// Call the function again with done channel closing early
	urlsChan = make(chan string, 10)
	go func() {
		if err := utils.ParseAndFilterJSONFromReader(reader, urlsChan, done); err != nil {
			t.Fatalf("ParseAndFilterJSONFromReader returned an error: %v", err)
		}
		close(urlsChan)
	}()

	// There should be no URLs processed after done is closed
	urls = nil
	for url := range urlsChan {
		urls = append(urls, url)
	}

	if len(urls) > 0 {
		t.Errorf("Expected no URLs to be processed after done is closed, but got %v", urls)
	}
}
