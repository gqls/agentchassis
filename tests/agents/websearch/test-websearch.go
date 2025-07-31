// test-websearch.go
package main

import (
	"fmt"
	"time"
)

func main() {
	// Test different scenarios
	tests := []struct {
		name     string
		provider string
		query    string
	}{
		{"Default provider", "", "latest AI news 2024"},
		{"Specific Firecrawl", "firecrawl", "kubernetes best practices"},
		{"Specific ScrapingBee", "scrapingbee", "golang web scraping"},
		{"Invalid provider (should fallback)", "invalid", "cloud computing trends"},
	}

	for _, test := range tests {
		fmt.Printf("\n=== Test: %s ===\n", test.name)
		sendSearchRequest(test.query, test.provider)
		time.Sleep(2 * time.Second)
	}
}

func sendSearchRequest(query, provider string) {
	// ... implement similar to your content-creator test
}
