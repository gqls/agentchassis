// FILE: platform/orchestration/datahelpers/nav_labels.go
// Navigation label utilities - shared across actions

package datahelpers

import (
	"strings"
)

// SimpleNavLabels maps page names to clean nav labels
var SimpleNavLabels = map[string]string{
	"index":            "Home",
	"home":             "Home",
	"about":            "About",
	"about-us":         "About",
	"services":         "Services",
	"contact":          "Contact",
	"contact-us":       "Contact",
	"insights":         "Insights",
	"blog":             "Blog",
	"careers":          "Careers",
	"team":             "Team",
	"leadership":       "Team",
	"pricing":          "Pricing",
	"faq":              "FAQ",
	"support":          "Support",
	"features":         "Features",
	"products":         "Products",
	"portfolio":        "Portfolio",
	"work":             "Work",
	"case-studies":     "Work",
	"clients":          "Clients",
	"resources":        "Resources",
	"privacy":          "Privacy",
	"privacy-policy":   "Privacy",
	"terms":            "Terms",
	"terms-of-service": "Terms",
}

// SimplifyNavLabel creates a clean, simple navigation label
// Handles verbose titles like "About Leopardess Consulting | Home" -> "About"
func SimplifyNavLabel(label, pageName string) string {
	if label == "" && pageName == "" {
		return ""
	}

	// If label is already short, keep it
	if len(label) <= 15 && label != "" {
		return label
	}

	// Clean up label - remove title separators
	cleanLabel := label
	if idx := strings.Index(cleanLabel, "|"); idx > 0 {
		cleanLabel = strings.TrimSpace(cleanLabel[:idx])
	}
	if idx := strings.Index(cleanLabel, " - "); idx > 0 {
		cleanLabel = strings.TrimSpace(cleanLabel[:idx])
	}

	// Try exact match on page name first
	pageNameLower := strings.ToLower(pageName)
	if simple, ok := SimpleNavLabels[pageNameLower]; ok {
		return simple
	}

	// Try prefix matching on page name
	for key, simple := range SimpleNavLabels {
		if strings.HasPrefix(pageNameLower, key) {
			return simple
		}
	}

	// Try matching first word of cleaned label
	words := strings.Fields(cleanLabel)
	if len(words) >= 1 {
		firstWordLower := strings.ToLower(words[0])
		if simple, ok := SimpleNavLabels[firstWordLower]; ok {
			return simple
		}
	}

	// If cleaned label is reasonable length, use it
	if len(cleanLabel) <= 20 && cleanLabel != "" {
		return cleanLabel
	}

	// Take first word if still long
	if len(words) > 0 {
		return strings.Title(strings.ToLower(words[0]))
	}

	// Fallback: capitalize page name
	if pageName != "" {
		return strings.Title(strings.ReplaceAll(pageName, "-", " "))
	}

	return label
}

// ExtractNameFromURL gets page name from URL path
// "/about.html" -> "about"
// "/services/consulting.html" -> "consulting"
func ExtractNameFromURL(url string) string {
	url = strings.TrimPrefix(url, "/")

	// Handle paths with directories
	if idx := strings.LastIndex(url, "/"); idx >= 0 {
		url = url[idx+1:]
	}

	url = strings.TrimSuffix(url, ".html")
	url = strings.TrimSuffix(url, ".htm")

	return url
}
