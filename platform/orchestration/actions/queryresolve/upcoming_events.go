// FILE: platform/orchestration/actions/queryresolve/upcoming_events.go
//
// `query.upcoming_events` — bugs_open/427: the render target for dated,
// correctable event facts (a confirmed fight, launch, hearing — whatever a
// site's evidence_base register has been given a `event_date` for). The
// populator is a sibling change (news_feed_ingestion, commit a7a134af7 and
// its follow-on): it registers a plain citation fact via
// VerifyAndRegisterCitationsAction, kind="entity" (EvidenceFact.Kind's closed
// vocabulary has no "event" kind — see datahelpers/claims.go), with
// event_date/venue/participants/broadcaster as extra top-level keys
// alongside the usual source.citation. This resolver is the first reader of
// those keys; nothing here writes them.
//
// WHY A QUERY SOURCE, NOT A NEW ACTION: the estate already has a designed
// answer to "a component shows rows derived from a store that changes" — see
// news_items.go's header for why HTML-patching and a client-fetched JSON file
// were both rejected in favour of a `query.*`-sourced schema field. The same
// argument applies here, unchanged: a scoped rerender regenerates
// content_data from html_template + stored fields, so items landing in
// content_data via a query source get refreshed, not wiped, and a wrong or
// past date is corrected by the SAME mechanism that corrects a stale news
// item — no new render path to get wrong.
//
// SITE-SCOPED, unlike the directory_entities-backed sources in this package:
// an evidence_base register belongs to one site, the way content_feed_items
// does. `directory_entities`/`directory_claims` is a different, global,
// cross-site registry for an unrelated product (AI-model/company/protocol
// directories) — confirmed unrelated to this bug's "entity-directory" PAGE
// ROLE by name only; see bugs_open/427 §4.4.
//
// NOTHING IS INVENTED: a fact whose event_date does not parse is EXCLUDED and
// logged, never rendered with a guessed date — the research spec's own
// lessons.avoid[] names a wrong fight date as actively harmful. A missing
// venue/broadcaster/participants is an absent key, not a placeholder; the
// template's own {{if}} guards decide what to show for it.

package queryresolve

import (
	"context"
	"database/sql"
	"encoding/json"
	"html"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

const (
	upcomingEventsDefaultLimit = 20
	upcomingEventsMaxLimit     = 50
)

// upcomingEvent is one dated fact, selected and validated but not yet
// projected for template rendering.
type upcomingEvent struct {
	FactID       string
	Date         time.Time
	DateText     string // the fact's own event_date string, as stored
	Venue        string
	Broadcaster  string
	Participants []string
	Claim        string
	SourceTitle  string
	SourceURL    string
}

// parseEventDate accepts the two forms a registered event_date is expected to
// carry: a full date, or (deliberately narrower than
// evidence_citations.go's parseFlexibleDate) a year-month — a calendar entry
// needs at least month precision to be worth listing. A bare year is refused
// here even though the citation-freshness parser accepts one for staleness
// ageing, which asks a different question (how OLD is this source) from the
// one this resolver asks (WHEN is this event, precisely enough to sort and
// show).
func parseEventDate(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{"2006-01-02", "2006-01"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// resolveUpcomingEvents backs `source: "query.upcoming_events"`. Reads the
// site's current evidence_base register directly (not through
// datahelpers.ParseEvidenceBase's typed struct — event_date/venue/
// participants/broadcaster are untyped extra keys by the same RFC_025 §9 Q2
// convention citation/artifact_check already follow, so a typed decode would
// silently see none of them).
func resolveUpcomingEvents(ctx context.Context, db *sql.DB, siteID uuid.UUID, limit int, logger *zap.Logger) (interface{}, error) {
	if limit <= 0 {
		limit = upcomingEventsDefaultLimit
	}
	if limit > upcomingEventsMaxLimit {
		limit = upcomingEventsMaxLimit
	}

	var rawJSON []byte
	err := db.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'evidence_base' AND is_current = true
	`, siteID).Scan(&rawJSON)
	if err == sql.ErrNoRows {
		return []map[string]interface{}{}, nil // no register yet — an empty list, not an error
	}
	if err != nil {
		return nil, err
	}

	var eb map[string]interface{}
	if err := json.Unmarshal(rawJSON, &eb); err != nil {
		logger.Warn("queryresolve: evidence_base did not decode as an object", zap.String("site_id", siteID.String()), zap.Error(err))
		return []map[string]interface{}{}, nil
	}
	factsRaw, _ := eb["facts"].([]interface{})

	today := time.Now().UTC().Truncate(24 * time.Hour)
	var events []upcomingEvent
	skippedUnparseable := 0
	for _, fr := range factsRaw {
		fact, ok := fr.(map[string]interface{})
		if !ok {
			continue
		}
		dateText := strings.TrimSpace(datahelpers.GetStringField(fact, "event_date", ""))
		if dateText == "" {
			continue // not an event fact
		}
		date, ok := parseEventDate(dateText)
		if !ok {
			skippedUnparseable++
			logger.Warn("queryresolve: event fact has an unparseable event_date — excluded, not guessed",
				zap.String("site_id", siteID.String()),
				zap.String("fact_id", datahelpers.GetStringField(fact, "id", "")),
				zap.String("event_date", dateText))
			continue
		}
		if date.Before(today) {
			continue // past events are not "upcoming" — the honest correction this whole fix is about
		}

		src, _ := fact["source"].(map[string]interface{})
		cit, _ := src["citation"].(map[string]interface{})
		events = append(events, upcomingEvent{
			FactID:       datahelpers.GetStringField(fact, "id", ""),
			Date:         date,
			DateText:     dateText,
			Venue:        strings.TrimSpace(datahelpers.GetStringField(fact, "venue", "")),
			Broadcaster:  strings.TrimSpace(datahelpers.GetStringField(fact, "broadcaster", "")),
			Participants: datahelpers.ExtractStringListHelper(fact["participants"]),
			Claim:        strings.TrimSpace(datahelpers.GetStringField(fact, "claim", "")),
			SourceTitle:  strings.TrimSpace(datahelpers.GetStringField(cit, "title", "")),
			SourceURL:    strings.TrimSpace(datahelpers.GetStringField(cit, "url", "")),
		})
	}

	sort.SliceStable(events, func(i, j int) bool {
		if !events[i].Date.Equal(events[j].Date) {
			return events[i].Date.Before(events[j].Date)
		}
		return events[i].Claim < events[j].Claim
	})
	if len(events) > limit {
		events = events[:limit]
	}

	if skippedUnparseable > 0 {
		logger.Info("queryresolve: resolved upcoming_events", zap.Int("items", len(events)), zap.Int("skipped_unparseable_date", skippedUnparseable))
	} else {
		logger.Info("queryresolve: resolved upcoming_events", zap.Int("items", len(events)))
	}
	return projectUpcomingEvents(events), nil
}

// projectUpcomingEvents HTML-escapes every string for text/template
// rendering (component_library.go's RenderTemplate does not auto-escape —
// same reasoning as projectNewsItems) and omits any field the fact did not
// carry, so the template's own {{if}} guards decide what a missing field
// shows rather than this resolver inventing a placeholder.
func projectUpcomingEvents(events []upcomingEvent) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(events))
	for _, e := range events {
		title := strings.Join(e.Participants, " vs ")
		if title == "" {
			title = e.Claim
		}
		m := map[string]interface{}{
			"fact_id": html.EscapeString(e.FactID),
			"title":   html.EscapeString(title),
			"date":    html.EscapeString(e.DateText),
		}
		if len(e.Participants) > 0 {
			participants := make([]string, 0, len(e.Participants))
			for _, p := range e.Participants {
				if p = strings.TrimSpace(p); p != "" {
					participants = append(participants, html.EscapeString(p))
				}
			}
			m["participants"] = participants
		}
		if e.Venue != "" {
			m["venue"] = html.EscapeString(e.Venue)
		}
		if e.Broadcaster != "" {
			m["broadcaster"] = html.EscapeString(e.Broadcaster)
		}
		if e.SourceTitle != "" {
			m["source_title"] = html.EscapeString(e.SourceTitle)
		}
		if e.SourceURL != "" {
			m["source_url"] = html.EscapeString(e.SourceURL)
		}
		out = append(out, m)
	}
	return out
}
