// FILE: internal/adapters/websearch/providers/published_at.go
package providers

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// relativeDateRe matches the human-readable relative dates Google-backed news
// APIs return, e.g. "3 hours ago", "2 days ago" (Firecrawl v2 documents news
// dates in exactly this form).
var relativeDateRe = regexp.MustCompile(`^(\d+)\s+(second|minute|hour|day|week|month|year)s?\s+ago$`)

var relativeUnits = map[string]time.Duration{
	"second": time.Second,
	"minute": time.Minute,
	"hour":   time.Hour,
	"day":    24 * time.Hour,
	"week":   7 * 24 * time.Hour,
	// Calendar approximations — news recency does not need day-exact months.
	"month": 30 * 24 * time.Hour,
	"year":  365 * 24 * time.Hour,
}

// absoluteDateLayouts are tried in order for absolute dates that are not
// already RFC3339.
var absoluteDateLayouts = []string{
	"2006-01-02T15:04:05.000Z",
	"2006-01-02 15:04:05",
	"2006-01-02",
	"Jan 2, 2006",
	"2 Jan 2006",
	"January 2, 2006",
}

// normalisePublishedAt converts the date formats search APIs actually return
// into RFC3339, because the feed writer parses published_at with
// time.Parse(time.RFC3339, …) and silently stores NULL for anything else
// (WriteFeedItemsAction, feed_actions.go). An unrecognised value is returned
// unchanged, which downstream treats exactly as it does today.
func normalisePublishedAt(raw string, now time.Time) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}

	if _, err := time.Parse(time.RFC3339, s); err == nil {
		return s
	}

	for _, layout := range absoluteDateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC().Format(time.RFC3339)
		}
	}

	lower := strings.ToLower(s)
	if m := relativeDateRe.FindStringSubmatch(lower); m != nil {
		n, err := strconv.Atoi(m[1])
		if err == nil {
			return now.Add(-time.Duration(n) * relativeUnits[m[2]]).UTC().Format(time.RFC3339)
		}
	}
	switch lower {
	case "just now", "today":
		return now.UTC().Format(time.RFC3339)
	case "yesterday":
		return now.Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	}

	return raw
}
