package utils

import (
	"encoding/json"
	"io"
	"log"
	"strings"
	"sync"
)

// InNetworkFile represents an entry in the in_network_files array
type InNetworkFile struct {
	Description string `json:"description"`
	Location    string `json:"location"`
}

// AllowedAmountFile represents the allowed amount file structure
type AllowedAmountFile struct {
	Description string `json:"description"`
	Location    string `json:"location"`
}

// ReportingPlan represents a reporting plan in the JSON structure
type ReportingPlan struct {
	PlanName       string `json:"plan_name"`
	PlanIDType     string `json:"plan_id_type"`
	PlanID         string `json:"plan_id"`
	PlanMarketType string `json:"plan_market_type"`
}

// ReportingStructure represents the nested structure within reporting_structure
type ReportingStructure struct {
	ReportingPlans    []ReportingPlan   `json:"reporting_plans"`
	InNetworkFiles    []InNetworkFile   `json:"in_network_files"`
	AllowedAmountFile AllowedAmountFile `json:"allowed_amount_file"`
}

// TableOfContents represents the top-level structure of the JSON file
type TableOfContents struct {
	ReportingEntityName string               `json:"reporting_entity_name"`
	ReportingEntityType string               `json:"reporting_entity_type"`
	ReportingStructures []ReportingStructure `json:"reporting_structure"`
}

// ParseAndFilterJSONFromReader parses the JSON from an io.Reader and filters URLs for the specified criteria.
// Results are sent to the provided channel.
func ParseAndFilterJSONFromReader(reader io.Reader, urlsChan chan<- string, done <-chan struct{}) error {
	// Read the entire JSON content
	jsonData, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	// Unmarshal the JSON content into the TableOfContents struct
	var toc TableOfContents
	if err := json.Unmarshal(jsonData, &toc); err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return err
	}

	var wg sync.WaitGroup
	workers := 4 // Number of concurrent workers
	tasks := make(chan ReportingStructure, len(toc.ReportingStructures))

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entry := range tasks {
				select {
				case <-done:
					return
				default:
					for _, plan := range entry.ReportingPlans {
						// Log the plan names for debugging
						log.Printf("Processing plan: %s", plan.PlanName)
						// Check if the plan matches the criteria
						if strings.Contains(plan.PlanName, "Anthem PPO") && (strings.Contains(toc.ReportingEntityName, "NY") || strings.Contains(toc.ReportingEntityName, "New York")) {
							for _, file := range entry.InNetworkFiles {
								urlsChan <- file.Location
								// Debug: Log the URL being sent to the channel
								log.Println("URL found:", file.Location)
							}
						}
					}
				}
			}
		}()
	}

	// Send tasks to workers
	for _, entry := range toc.ReportingStructures {
		tasks <- entry
	}
	close(tasks)

	// Wait for workers to finish
	wg.Wait()
	return nil
}
