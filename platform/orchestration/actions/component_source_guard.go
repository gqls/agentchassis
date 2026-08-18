// FILE: platform/orchestration/actions/component_source_guard.go
//
// A generated component's input_schema must not declare a data source the
// platform cannot resolve. This guard is the birth gate for that rule
// (bugs_open/309).
//
// Why it exists: component generation is an LLM emitting a schema, and the
// schema's `source` strings are executed verbatim by plan_sections'
// sourceResolver. Nothing between generation and render ever checked that a
// declared source is inside the resolver's vocabulary, or that a
// `site_specs.<aspect>` path names an aspect any writer produces. The
// motivating case: `blog-listing_pre_037` (generated 2026-04-08) declared
// `post1_url…post6_url` as required fields sourced from
// `site_specs.blog.postN_url` — an aspect that has NEVER existed on any site —
// so every URL resolved nowhere, on_missing defaulted to skip_field, and the
// template's `{{if .postN_url}}` gate silently swallowed every anchor:
// fundamentallyai.com's Platform Log index served six complete-looking article
// cards with zero links, indistinguishable from success at every stage.
//
// Calibration (2026-08-18, against every active component's schema — census
// queries in docs024_key_docs_latest/bugfix_309_unclickable_index_cards/
// RUNBOOK): the check flags 58 fields across 11 components declaring 10
// site_specs aspects that exist on no site, 7 `query.*` names the resolver
// does not register, and 3 fields with prefixes outside the vocabulary
// entirely (`nav.*`, `site.*`). Every flagged field currently resolves
// NOWHERE — the phantom aspects have zero site_specs rows in all history, and
// an unknown query name errors at plan time — so the guard refuses nothing
// that works today. Zero flags on every field whose source is resolvable.
//
// NOTE for whoever hits a refusal on a REGENERATION of one of those 11: the
// field-contract check (stranded-fields, store_generated_component_action.go)
// requires keeping the existing field NAMES, and this guard refuses their
// phantom SOURCES — so the regeneration must repoint the source (for listings,
// a `query.*` source is almost always the right one), or the component must be
// repaired by migration. That is deliberate, and it is the moment the phantom
// source becomes visible instead of shipping again — the same shape as the
// fabricated-fallback guard's regeneration note.
//
// Scope: top-level `fields` only. Nested `items` sub-schemas use `field.<key>`
// references into each query-result item and are validated by the query
// machinery that consumes them, not here.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
)

// bareSourceValues are the source strings that need no resolution at plan
// time (plan_sections resolve() returns found=true for them untouched). An
// absent/empty source is treated as one of these — the field loop tolerates
// it today and this guard must not tighten that behaviour.
var bareSourceValues = map[string]bool{
	"":         true,
	"llm":      true,
	"renderer": true,
	"static":   true,
}

// prefixedSourceKinds are the `<prefix>.<path>` families plan_sections
// resolve() dispatches on. A prefix outside this set falls to resolve()'s
// default arm — (nil, false), silently — which is exactly the outcome this
// guard exists to refuse at birth.
var prefixedSourceKinds = map[string]bool{
	"renderer":    true, // "renderer.x" — injected at render time
	"static":      true, // "static.x" — same
	"site_specs":  true, // per-site spec aspects; aspect vocabulary checked below
	"site_assets": true, // per-site asset keys + image-role alias, per-site at plan time
	"pages":       true, // per-site page names, resolved per-site at plan time
	"config":      true, // per-site content_data config
	"query":       true, // named queries; name checked against queryresolve
}

// sourceVocabularyIssues walks a component input_schema (house dialect,
// {"fields": {...}}) and returns one issue string per field whose declared
// source cannot resolve for ANY site. Pure: the caller supplies the live
// site_specs aspect vocabulary. knownSpecAspects == nil means the aspect set
// could not be loaded — the site_specs aspect check is then SKIPPED (fail
// open: a transient read failure must not block all component generation),
// while the prefix and query-name checks, which need no DB, still run.
func sourceVocabularyIssues(inputSchemaJSON string, knownSpecAspects map[string]bool) []string {
	var schema struct {
		Fields map[string]json.RawMessage `json:"fields"`
	}
	if err := json.Unmarshal([]byte(inputSchemaJSON), &schema); err != nil || len(schema.Fields) == 0 {
		// Malformed / empty schemas are other checks' remit (Check 3/4).
		return nil
	}

	// Deterministic ordering: issues join into an error message and a
	// rejection record; their order must not vary run to run.
	names := make([]string, 0, len(schema.Fields))
	for name := range schema.Fields {
		names = append(names, name)
	}
	sort.Strings(names)

	var issues []string
	for _, name := range names {
		var def struct {
			Source string `json:"source"`
		}
		if err := json.Unmarshal(schema.Fields[name], &def); err != nil {
			continue
		}
		source := def.Source
		if bareSourceValues[source] {
			continue
		}

		prefix, path, hasDot := strings.Cut(source, ".")
		if !hasDot || !prefixedSourceKinds[prefix] {
			issues = append(issues, fmt.Sprintf(
				"field %q declares source %q, which is outside the resolver's vocabulary — it would resolve nowhere on every site and the field would be silently omitted (bugs_open/309); valid sources: llm, renderer, static, site_specs.*, site_assets.*, pages.*, config.*, query.*",
				name, source))
			continue
		}

		switch prefix {
		case "query":
			if !queryresolve.IsKnownQueryName(path) {
				issues = append(issues, fmt.Sprintf(
					"field %q declares source %q but no such query is registered — it would error at plan time and the field would be silently omitted (bugs_open/309); registered queries: %s",
					name, source, strings.Join(queryresolve.KnownQueryBases(), ", ")))
			}
		case "site_specs":
			if knownSpecAspects == nil {
				continue // aspect set unavailable; see function comment
			}
			aspect, _, _ := strings.Cut(path, ".")
			if !knownSpecAspects[aspect] {
				aspects := make([]string, 0, len(knownSpecAspects))
				for a := range knownSpecAspects {
					aspects = append(aspects, a)
				}
				sort.Strings(aspects)
				issues = append(issues, fmt.Sprintf(
					"field %q declares source %q but no site carries a site_specs aspect named %q — the value would resolve nowhere on every site and the field would be silently omitted (bugs_open/309); aspects that exist: %s; for listing data use a query.* source (e.g. query.blog_posts)",
					name, source, aspect, strings.Join(aspects, ", ")))
			}
		}
	}
	return issues
}

// loadKnownSpecAspects reads the live site_specs aspect vocabulary — every
// aspect ANY site has ever carried (superseded rows included, deliberately: an
// aspect with only historical rows still names a store a writer produces, and
// refusing it would flag components that worked). Callers treat an error as
// "check unavailable" (pass nil to sourceVocabularyIssues), never as a
// rejection.
func loadKnownSpecAspects(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT DISTINCT aspect FROM site_specs`)
	if err != nil {
		return nil, fmt.Errorf("loadKnownSpecAspects: %w", err)
	}
	defer rows.Close()

	aspects := make(map[string]bool)
	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			return nil, fmt.Errorf("loadKnownSpecAspects: %w", err)
		}
		aspects[a] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("loadKnownSpecAspects: %w", err)
	}
	return aspects, nil
}
