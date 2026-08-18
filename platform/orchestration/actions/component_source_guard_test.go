// FILE: platform/orchestration/actions/component_source_guard_test.go
//
// The motivating-case tests pin the EXACT schema shapes from bugs_open/309:
// blog-listing_pre_037's phantom `site_specs.blog.*` sources (must flag) and
// content-listing's `query.blog_posts` source (must pass). If the guard stops
// flagging the first or starts flagging the second, the class has reopened or
// the guard has gone rabid — either way these fail before production does.

package actions

import (
	"strings"
	"testing"
)

// liveAspects mirrors the fleet's real aspect vocabulary at calibration time
// (2026-08-18). The guard takes the set as input, so the test set only needs
// to be representative, not current.
func liveAspects() map[string]bool {
	return map[string]bool{
		"briefing": true, "classification": true, "contact": true,
		"content_direction": true, "cta": true, "design_intent": true,
		"evidence_base": true, "identity": true, "mission_brief": true,
		"portfolio": true, "site_config": true, "strategy": true,
		"structure": true, "tools": true,
	}
}

func TestSourceGuardFlagsThePhantomBlogAspect(t *testing.T) {
	// blog-listing_pre_037's actual declarations (bugs_open/309): required
	// URLs sourced from a site_specs aspect no site has ever carried.
	schema := `{"fields": {
		"post1_url": {"type": "url", "source": "site_specs.blog.post1_url", "required": true},
		"post2_url": {"type": "url", "source": "site_specs.blog.post2_url", "required": true},
		"cta_url":   {"type": "url", "source": "site_specs.blog.archive_url", "required": false},
		"post1_title": {"type": "text", "source": "llm", "required": true}
	}}`
	issues := sourceVocabularyIssues(schema, liveAspects())
	if len(issues) != 3 {
		t.Fatalf("want 3 issues (the three phantom-aspect URLs), got %d: %v", len(issues), issues)
	}
	for _, issue := range issues {
		if !strings.Contains(issue, `"blog"`) || !strings.Contains(issue, "bugs_open/309") {
			t.Errorf("issue should name the phantom aspect and the bug: %s", issue)
		}
	}
	// Deterministic order: fields are walked sorted, so cta_url first.
	if !strings.Contains(issues[0], `"cta_url"`) {
		t.Errorf("issues not in sorted field order, first was: %s", issues[0])
	}
}

func TestSourceGuardPassesContentListing(t *testing.T) {
	// content-listing's actual schema — the framework-native article listing.
	schema := `{"fields": {
		"articles": {"type": "array", "source": "query.blog_posts", "required": true, "on_missing": "skip_section"},
		"section_title": {"type": "text", "source": "llm", "required": false},
		"load_more_text": {"type": "text", "source": "static", "fallback": "Load More", "required": false},
		"show_load_more": {"type": "boolean", "source": "static", "fallback": false, "required": false}
	}}`
	if issues := sourceVocabularyIssues(schema, liveAspects()); len(issues) != 0 {
		t.Fatalf("content-listing schema must pass, got: %v", issues)
	}
}

func TestSourceGuardFlagsUnknownQueryNames(t *testing.T) {
	// Real declarations from the 2026-08-18 census: 7 query names the
	// resolver does not register. One representative here; a known name
	// with an argument alongside, which must pass.
	schema := `{"fields": {
		"featured": {"type": "object", "source": "query.featured_post", "required": false},
		"tools":    {"type": "array",  "source": "query.pages_where_type:tool", "required": false}
	}}`
	issues := sourceVocabularyIssues(schema, liveAspects())
	if len(issues) != 1 {
		t.Fatalf("want exactly the unknown query flagged, got %d: %v", len(issues), issues)
	}
	if !strings.Contains(issues[0], `"featured"`) || !strings.Contains(issues[0], "blog_posts") {
		t.Errorf("issue should name the field and list registered queries: %s", issues[0])
	}
}

func TestSourceGuardFlagsUnknownPrefixes(t *testing.T) {
	// The census's `nav.*` / `site.*` stragglers, plus a bare unknown word.
	schema := `{"fields": {
		"links":  {"type": "array", "source": "nav.primary_links"},
		"name":   {"type": "text",  "source": "site.title"},
		"orphan": {"type": "text",  "source": "mystery"}
	}}`
	issues := sourceVocabularyIssues(schema, liveAspects())
	if len(issues) != 3 {
		t.Fatalf("want 3 unknown-prefix issues, got %d: %v", len(issues), issues)
	}
}

func TestSourceGuardToleratesBareAndPerSiteSources(t *testing.T) {
	// Everything here needs no fleet-wide existence check: bare sources are
	// resolved at render time, and site_assets/pages/config leaves are
	// per-site facts resolved at plan time.
	schema := `{"fields": {
		"a": {"type": "text", "source": "llm"},
		"b": {"type": "text", "source": "static", "fallback": "x"},
		"c": {"type": "url",  "source": "renderer"},
		"d": {"type": "text"},
		"e": {"type": "url",  "source": "site_assets.hero"},
		"f": {"type": "url",  "source": "pages.contact"},
		"g": {"type": "text", "source": "config.theme.accent"},
		"h": {"type": "text", "source": "site_specs.identity.company_name"}
	}}`
	if issues := sourceVocabularyIssues(schema, liveAspects()); len(issues) != 0 {
		t.Fatalf("all sources are resolvable shapes, got: %v", issues)
	}
}

func TestSourceGuardSkipsAspectCheckWhenSetUnavailable(t *testing.T) {
	// nil aspect set = the loader failed. The site_specs half must go quiet
	// (fail open), the DB-free halves must still fire.
	schema := `{"fields": {
		"u": {"type": "url", "source": "site_specs.blog.post1_url", "required": true},
		"q": {"type": "array", "source": "query.no_such_query"}
	}}`
	issues := sourceVocabularyIssues(schema, nil)
	if len(issues) != 1 {
		t.Fatalf("want only the unknown query (aspect check skipped on nil set), got %d: %v", len(issues), issues)
	}
	if !strings.Contains(issues[0], `"q"`) {
		t.Errorf("surviving issue should be the query one: %s", issues[0])
	}
}

func TestSourceGuardIgnoresMalformedAndEmptySchemas(t *testing.T) {
	// Malformed / empty schemas are Check 3 / Check 4's remit — the guard
	// must contribute nothing there, not double-report.
	for _, schema := range []string{``, `{}`, `{"fields":{}}`, `not json`, `{"properties":{"x":{}}}`} {
		if issues := sourceVocabularyIssues(schema, liveAspects()); len(issues) != 0 {
			t.Errorf("schema %q: want no issues, got %v", schema, issues)
		}
	}
}
