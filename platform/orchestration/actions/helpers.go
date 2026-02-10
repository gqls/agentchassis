package actions

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// PageInfo holds page details for rendering
type PageInfo struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	SiteID      uuid.UUID
	AreaID      *uuid.UUID // nullable
	Filename    string     `json:"filename"`
	MetaDesc    string     `json:"meta_desc"`
	Description string     `json:"description"`
	Domain      string
}

func getDomainForSite(ctx context.Context, db *sql.DB, siteID uuid.UUID) (string, error) {
	var domain string
	err := db.QueryRowContext(ctx, `SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain)
	return domain, err
}

// extractDomain returns the root domain from a URL (e.g. "www.example.co.uk" -> "example.co.uk")
func extractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := strings.ToLower(parsed.Hostname())
	host = strings.TrimPrefix(host, "www.")
	return host
}

// extractRootURL returns the scheme + host (without path) and without www
func extractRootURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	host := strings.TrimPrefix(parsed.Hostname(), "www.")
	return fmt.Sprintf("https://%s", host)
}

// sameWebsite checks if two URLs point to the same website
func sameWebsite(url1, url2 string) bool {
	d1 := extractDomain(url1)
	d2 := extractDomain(url2)
	return d1 != "" && d1 == d2
}

// extractPracticeName tries to pull a practice name from a search result title
// Titles tend to be "Practice Name - Location | Something" or "Practice Name | Vets in ..."
func extractPracticeName(title string) string {
	title = strings.TrimSpace(title)
	// Split on common separators
	for _, sep := range []string{" | ", " - ", " – ", " — "} {
		parts := strings.SplitN(title, sep, 2)
		if len(parts) > 0 && len(parts[0]) > 3 {
			return strings.TrimSpace(parts[0])
		}
	}
	return title
}

// extractUKPostcode finds a UK postcode pattern in text
// Pattern: 1-2 letters, 1-2 digits, optional space, digit, 2 letters
func extractUKPostcode(text string) string {
	text = strings.ToUpper(text)
	// Simple scan - look for postcode-shaped sequences
	words := strings.Fields(text)
	for i := 0; i < len(words)-1; i++ {
		candidate := words[i] + " " + words[i+1]
		if looksLikeUKPostcode(candidate) {
			return candidate
		}
		// Also try without space
		if looksLikeUKPostcode(words[i]) {
			// Insert space before last 3 chars
			if len(words[i]) >= 5 {
				pc := words[i][:len(words[i])-3] + " " + words[i][len(words[i])-3:]
				return pc
			}
		}
	}
	return ""
}

func looksLikeUKPostcode(s string) bool {
	s = strings.TrimSpace(s)
	// Must be 6-8 chars (with space)
	if len(s) < 6 || len(s) > 8 {
		return false
	}
	// Must contain a space
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return false
	}
	// Second part must be digit + 2 letters
	if len(parts[1]) != 3 {
		return false
	}
	if parts[1][0] < '0' || parts[1][0] > '9' {
		return false
	}
	if parts[1][1] < 'A' || parts[1][1] > 'Z' {
		return false
	}
	if parts[1][2] < 'A' || parts[1][2] > 'Z' {
		return false
	}
	// First part: 2-4 chars, starts with letter
	if len(parts[0]) < 2 || len(parts[0]) > 4 {
		return false
	}
	if parts[0][0] < 'A' || parts[0][0] > 'Z' {
		return false
	}
	return true
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
