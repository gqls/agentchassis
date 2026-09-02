// FILE: platform/orchestration/actions/queryresolve/upcoming_events_test.go
//
// bugs_open/427: resolveUpcomingEvents reads a site's evidence_base register
// directly off the raw jsonb map (never through the typed EvidenceFact
// struct, which has no event_date/venue/participants/broadcaster fields —
// RFC_025 §9 Q2's untyped-map convention). These tests pin: selection
// (event_date present, parseable, not in the past), exclusion of an
// unparseable date (never guessed), ordering, the limit/cap, and escaping.

package queryresolve

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func registerRow(t *testing.T, facts []map[string]interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(map[string]interface{}{"facts": facts})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	return b
}

func TestResolveUpcomingEvents_NoRegisterIsEmptyNotError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM site_specs").WillReturnRows(sqlmock.NewRows([]string{"data"}))

	got, err := resolveUpcomingEvents(context.Background(), db, uuid.New(), 0, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if items := got.([]map[string]interface{}); len(items) != 0 {
		t.Fatalf("want empty, got %#v", items)
	}
}

func TestResolveUpcomingEvents_SelectionOrderingAndExclusions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	future1 := time.Now().UTC().AddDate(0, 0, 30).Format("2006-01-02")
	future2 := time.Now().UTC().AddDate(0, 0, 10).Format("2006-01-02")
	past := time.Now().UTC().AddDate(0, 0, -5).Format("2006-01-02")

	facts := []map[string]interface{}{
		{ // no event_date at all — not an event fact, excluded silently
			"id": "F-plain", "claim": "some other fact", "kind": "metric", "value": 42.0,
		},
		{ // the later of the two future events
			"id": "EVT-later", "kind": "entity", "event_date": future1,
			"participants": []interface{}{"Tyson Fury", "Oleksandr Usyk"},
			"venue":        "Wembley Stadium", "broadcaster": "DAZN",
			"source": map[string]interface{}{"citation": map[string]interface{}{
				"title": "Fury vs Usyk announced", "url": "https://example.com/fury-usyk",
				"quote": "Fury will fight Usyk",
			}},
		},
		{ // the sooner of the two future events — must sort first
			"id": "EVT-sooner", "kind": "entity", "event_date": future2,
			"participants": []interface{}{"Anthony Joshua", "Deontay Wilder"},
			"source": map[string]interface{}{"citation": map[string]interface{}{
				"url": "https://example.com/joshua-wilder", "quote": "Joshua vs Wilder confirmed",
			}},
		},
		{ // unparseable date — excluded, never guessed
			"id": "EVT-bad-date", "kind": "entity", "event_date": "sometime in March",
			"participants": []interface{}{"A", "B"},
			"source": map[string]interface{}{"citation": map[string]interface{}{
				"url": "https://example.com/ab", "quote": "A vs B",
			}},
		},
		{ // a concluded event — not "upcoming"
			"id": "EVT-past", "kind": "entity", "event_date": past,
			"participants": []interface{}{"C", "D"},
			"source": map[string]interface{}{"citation": map[string]interface{}{
				"url": "https://example.com/cd", "quote": "C vs D",
			}},
		},
		{ // has a citation URL but no quote — must be excluded, not just warned
			"id": "EVT-no-quote", "kind": "entity", "event_date": future2,
			"participants": []interface{}{"E", "F"},
			"source": map[string]interface{}{"citation": map[string]interface{}{
				"url": "https://example.com/ef",
			}},
		},
		{ // has a citation quote but no url — must be excluded
			"id": "EVT-no-url", "kind": "entity", "event_date": future2,
			"participants": []interface{}{"G", "H"},
			"source": map[string]interface{}{"citation": map[string]interface{}{
				"quote": "G vs H",
			}},
		},
	}

	mock.ExpectQuery("FROM site_specs").WillReturnRows(
		sqlmock.NewRows([]string{"data"}).AddRow(registerRow(t, facts)))

	got, err := resolveUpcomingEvents(context.Background(), db, uuid.New(), 0, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items := got.([]map[string]interface{})
	if len(items) != 2 {
		t.Fatalf("want exactly the 2 valid future events, got %d: %#v", len(items), items)
	}
	if items[0]["fact_id"] != "EVT-sooner" || items[1]["fact_id"] != "EVT-later" {
		t.Errorf("want ascending date order (sooner first), got %v then %v", items[0]["fact_id"], items[1]["fact_id"])
	}
	if items[1]["title"] != "Tyson Fury vs Oleksandr Usyk" {
		t.Errorf("want participants joined with \" vs \" as the title, got %v", items[1]["title"])
	}
	if items[1]["venue"] != "Wembley Stadium" || items[1]["broadcaster"] != "DAZN" {
		t.Errorf("venue/broadcaster missing or wrong: %#v", items[1])
	}
	if items[1]["source_url"] != "https://example.com/fury-usyk" {
		t.Errorf("source_url missing or wrong: %#v", items[1])
	}
	// The sooner event carries no venue/broadcaster — absent keys, not
	// empty-string placeholders, so a template's {{if}} can tell.
	if _, has := items[0]["venue"]; has {
		t.Errorf("event with no stated venue must omit the key entirely, got %#v", items[0])
	}
	// Every item carries the disclaimer regardless of how complete it is —
	// compliance's caveat-travels-with-the-claim requirement (council REVISE
	// 08f56b7e).
	for _, it := range items {
		if it["disclaimer"] != upcomingEventDisclaimer {
			t.Errorf("every item must carry the disclaimer, got %#v", it)
		}
	}
}

// TestResolveUpcomingEvents_RequiresCitation — council REVISE 08f56b7e,
// compliance HIGH: a fact must not render real-world scheduling claims
// (fights, venues, real people) with nothing checking it is actually
// evidenced. Both halves of a citation (url, quote) are required; either
// alone is not enough, and a fact with neither or only one is excluded and
// logged exactly like an unparseable date, not silently rendered.
func TestResolveUpcomingEvents_RequiresCitation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	future := time.Now().UTC().AddDate(0, 0, 5).Format("2006-01-02")
	facts := []map[string]interface{}{
		{"id": "EVT-none", "kind": "entity", "event_date": future, "participants": []interface{}{"A", "B"}},
		{"id": "EVT-url-only", "kind": "entity", "event_date": future, "participants": []interface{}{"C", "D"},
			"source": map[string]interface{}{"citation": map[string]interface{}{"url": "https://example.com/cd"}}},
		{"id": "EVT-quote-only", "kind": "entity", "event_date": future, "participants": []interface{}{"E", "F"},
			"source": map[string]interface{}{"citation": map[string]interface{}{"quote": "E vs F"}}},
		{"id": "EVT-both", "kind": "entity", "event_date": future, "participants": []interface{}{"G", "H"},
			"source": map[string]interface{}{"citation": map[string]interface{}{"url": "https://example.com/gh", "quote": "G vs H"}}},
	}
	mock.ExpectQuery("FROM site_specs").WillReturnRows(
		sqlmock.NewRows([]string{"data"}).AddRow(registerRow(t, facts)))

	got, err := resolveUpcomingEvents(context.Background(), db, uuid.New(), 0, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items := got.([]map[string]interface{})
	if len(items) != 1 || items[0]["fact_id"] != "EVT-both" {
		t.Fatalf("want exactly EVT-both (the only fact with both url and quote), got %#v", items)
	}
}

func TestResolveUpcomingEvents_EscapesUntrustedText(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	future := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	facts := []map[string]interface{}{
		{
			"id": "EVT-xss", "kind": "entity", "event_date": future,
			"participants": []interface{}{"<script>alert(1)</script>"},
			"venue":        "Arena & Co",
			"source": map[string]interface{}{"citation": map[string]interface{}{
				"url": "https://example.com/xss", "quote": "the fixture",
			}},
		},
	}
	mock.ExpectQuery("FROM site_specs").WillReturnRows(
		sqlmock.NewRows([]string{"data"}).AddRow(registerRow(t, facts)))

	got, err := resolveUpcomingEvents(context.Background(), db, uuid.New(), 0, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items := got.([]map[string]interface{})
	if len(items) != 1 {
		t.Fatalf("want 1 item, got %#v", items)
	}
	if items[0]["title"] == "<script>alert(1)</script>" {
		t.Errorf("participant text must be HTML-escaped, got raw: %v", items[0]["title"])
	}
	if items[0]["venue"] == "Arena & Co" {
		t.Errorf("venue text must be HTML-escaped, got raw: %v", items[0]["venue"])
	}
}

func TestResolveUpcomingEvents_LimitDefaultAndCap(t *testing.T) {
	makeFacts := func(n int) []map[string]interface{} {
		var facts []map[string]interface{}
		for i := 0; i < n; i++ {
			facts = append(facts, map[string]interface{}{
				"id": uuid.New().String(), "kind": "entity",
				"event_date":   time.Now().UTC().AddDate(0, 0, i+1).Format("2006-01-02"),
				"participants": []interface{}{"A", "B"},
				"source": map[string]interface{}{"citation": map[string]interface{}{
					"url": "https://example.com/x", "quote": "A vs B",
				}},
			})
		}
		return facts
	}

	for _, c := range []struct {
		name        string
		nFacts      int
		limit       int
		wantResults int
	}{
		{"default cap is 20", 25, 0, upcomingEventsDefaultLimit},
		{"explicit limit under cap", 10, 5, 5},
		{"explicit limit over the hard cap is clamped", 60, 100, upcomingEventsMaxLimit},
	} {
		t.Run(c.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()
			mock.ExpectQuery("FROM site_specs").WillReturnRows(
				sqlmock.NewRows([]string{"data"}).AddRow(registerRow(t, makeFacts(c.nFacts))))

			got, err := resolveUpcomingEvents(context.Background(), db, uuid.New(), c.limit, zap.NewNop())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if items := got.([]map[string]interface{}); len(items) != c.wantResults {
				t.Errorf("want %d items, got %d", c.wantResults, len(items))
			}
		})
	}
}

func TestParseEventDate(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"2027-03-03", true},
		{"2027-03", true},
		{"2027", false},    // bare year: not precise enough to list
		{"March 3", false}, // not the stored ISO-ish form
		{"", false},
		{"  2027-03-03  ", true}, // surrounding whitespace tolerated
	}
	for _, c := range cases {
		if _, ok := parseEventDate(c.in); ok != c.want {
			t.Errorf("parseEventDate(%q) ok = %v, want %v", c.in, ok, c.want)
		}
	}
}
