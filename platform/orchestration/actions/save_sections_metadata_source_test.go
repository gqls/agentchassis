package actions

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// Tests for the metadata-source seam on save_page_sections (bugs_open/194).
//
// EVERY FIXTURE HERE IS BUILT TO BE DISCRIMINATING, because the defect these
// guard against is one that passed six months of green builds: an action that
// silently persisted NULL content_data reported success every time. In
// particular the two paths in TestConfiguredFieldWinsOverTheDefault hold
// DIFFERENT arrays — with identical ones the test would pass whichever path won,
// which is the "two blind checks agree with each other" shape.
//
// The mutation that must break each test is named on the test itself. A test
// whose mutation still passes is not testing what its name says.

// resolveSectionsForTest runs the seam exactly as the action does: resolve the
// field, extract from collected_data, parse into SectionData. Deliberately not a
// helper that hides the resolution — the resolution IS what is under test.
func resolveSectionsForTest(config map[string]interface{}, collected map[string]interface{}) ([]SectionData, string, string) {
	field, origin := resolveSectionsMetadataField(config)
	if field == "" {
		return nil, field, origin
	}
	raw := datahelpers.ExtractNestedField(collected, field)
	if raw == nil {
		return nil, field, origin
	}
	return extractSectionsFromMetadata(raw, zap.NewNop()), field, origin
}

func metadataEntry(slot, html string, contentData map[string]interface{}) map[string]interface{} {
	e := map[string]interface{}{
		"rendered_html":      html,
		"component_function": slot,
	}
	if contentData != nil {
		e["content_data"] = contentData
	}
	return e
}

// writerReplyCollected is the shape every page_content-carrying caller produces:
// a call_agent step with output_field `page_content`, whose reply nests under
// `.response`. Measured on the live definitions of page-build-handler,
// page-rebuild, pageflow-builder and site-work-orchestrator, 2026-08-04.
func writerReplyCollected(entries ...interface{}) map[string]interface{} {
	return map[string]interface{}{
		"page_content": map[string]interface{}{
			"response": map[string]interface{}{
				"sections_metadata": entries,
			},
		},
	}
}

func TestResolveSectionsMetadataFieldStates(t *testing.T) {
	cases := []struct {
		name       string
		config     map[string]interface{}
		wantField  string
		wantOrigin string
	}{
		{
			name:       "configured field is used verbatim",
			config:     map[string]interface{}{sectionsMetadataFieldKey: "rerender_sections.sections_metadata"},
			wantField:  "rerender_sections.sections_metadata",
			wantOrigin: metadataOriginConfigured,
		},
		{
			// The bug: four of six live callers had exactly this config.
			name:       "nothing configured falls back to the shared default",
			config:     map[string]interface{}{"html_field": "assembled_page.html"},
			wantField:  defaultSectionsMetadataField,
			wantOrigin: metadataOriginDefault,
		},
		{
			name:       "an empty string is not a configuration",
			config:     map[string]interface{}{sectionsMetadataFieldKey: ""},
			wantField:  defaultSectionsMetadataField,
			wantOrigin: metadataOriginDefault,
		},
		{
			name:       "a caller may declare it has no structured content",
			config:     map[string]interface{}{expectsNoSectionsMetadataKey: true},
			wantField:  "",
			wantOrigin: metadataOriginDeclaredAbsent,
		},
		{
			// The declaration wins over a stale field left behind, so retiring a
			// caller's structured content is one key, not two edits.
			name: "the declaration outranks a leftover field",
			config: map[string]interface{}{
				expectsNoSectionsMetadataKey: true,
				sectionsMetadataFieldKey:     "page_content.response.sections_metadata",
			},
			wantField:  "",
			wantOrigin: metadataOriginDeclaredAbsent,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			field, origin := resolveSectionsMetadataField(tc.config)
			if field != tc.wantField || origin != tc.wantOrigin {
				t.Fatalf("got (%q, %q), want (%q, %q)", field, origin, tc.wantField, tc.wantOrigin)
			}
		})
	}
}

// TestUnsetFieldResolvesTheDefaultAndKeepsContentData is the test that pins
// bugs_open/194 itself: with no config key at all, the writer's structured
// sections must still reach the save carrying their content_data.
//
// MUTATION: delete the default branch of resolveSectionsMetadataField (i.e.
// restore the pre-fix `if f, ok := config[...]; ok && f != ""` gate) and this
// fails — no sections resolve, which is precisely the state that wrote NULL.
func TestUnsetFieldResolvesTheDefaultAndKeepsContentData(t *testing.T) {
	collected := writerReplyCollected(
		metadataEntry("hero", "<section>hero</section>", map[string]interface{}{"headline": "Fuel supply, simplified"}),
		metadataEntry("article-body", "<section>body</section>", map[string]interface{}{"content": "…"}),
	)

	sections, field, origin := resolveSectionsForTest(map[string]interface{}{"html_field": "assembled_page.html"}, collected)

	if origin != metadataOriginDefault || field != defaultSectionsMetadataField {
		t.Fatalf("expected the default path, got %q (%s)", field, origin)
	}
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections from the default path, got %d", len(sections))
	}
	// The INSERT writes content_data only when len(ContentData) > 0
	// (save_page_sections_action.go, "Marshal content_data to JSON if present"),
	// so that predicate — not merely non-nil — is what the test asserts.
	for _, s := range sections {
		if len(s.ContentData) == 0 {
			t.Fatalf("section %q reached the save with no content_data: this is the 194 defect", s.ComponentName)
		}
	}
}

// TestConfiguredFieldWinsOverTheDefault pins page-rerender's semantics: it names
// its own path and must keep getting it. 2,878 runs in nine days go through here.
//
// MUTATION: consult the default first (or probe both and prefer the default) and
// this fails — the two paths hold DIFFERENT slots on purpose, so the assertion
// can tell which one was read.
func TestConfiguredFieldWinsOverTheDefault(t *testing.T) {
	collected := writerReplyCollected(
		metadataEntry("wrong-path-hero", "<section>from page_content</section>", map[string]interface{}{"a": 1}),
	)
	collected["rerender_sections"] = map[string]interface{}{
		"sections_metadata": []interface{}{
			metadataEntry("rerendered-hero", "<section>from rerender</section>", map[string]interface{}{"b": 2}),
		},
	}

	sections, _, origin := resolveSectionsForTest(
		map[string]interface{}{sectionsMetadataFieldKey: "rerender_sections.sections_metadata"}, collected)

	if origin != metadataOriginConfigured {
		t.Fatalf("origin = %q, want %q", origin, metadataOriginConfigured)
	}
	if len(sections) != 1 || sections[0].ComponentName != "rerendered-hero" {
		t.Fatalf("the configured path did not win: got %+v", sections)
	}
}

// TestConfiguredButUnresolvingFieldDoesNotFallToTheDefault is the no-op case —
// the one a change like this is least likely to check. A caller that NAMES a
// field and finds nothing there must behave exactly as it did before the default
// existed: fall to HTML parsing, not quietly to somebody else's metadata.
//
// MUTATION: make the default a fallback for an unresolving configured field
// ("probe anyway") and this fails.
func TestConfiguredButUnresolvingFieldDoesNotFallToTheDefault(t *testing.T) {
	collected := writerReplyCollected(
		metadataEntry("hero", "<section>hero</section>", map[string]interface{}{"headline": "x"}),
	)

	sections, field, origin := resolveSectionsForTest(
		map[string]interface{}{sectionsMetadataFieldKey: "rerender_sections.sections_metadata"}, collected)

	if origin != metadataOriginConfigured || field != "rerender_sections.sections_metadata" {
		t.Fatalf("got (%q, %q), want the configured path", field, origin)
	}
	if len(sections) != 0 {
		t.Fatalf("a configured-but-absent field resolved %d sections from elsewhere: %+v", len(sections), sections)
	}
}

// TestToolRecreationShapeResolvesNothing covers the one live caller whose NULL
// content_data is CORRECT: tool-recreation-handler recreates a whole-page tool
// and has no writer step at all, so there is no page_content anywhere in its
// collected_data (measured off its live definition 2026-08-04).
//
// It is asserted twice over, because the two protections are independent: the
// declaration suppresses the lookup, AND the default cannot resolve on that shape
// anyway (ExtractNestedField walks top-level keys plus a `.response` unwrap; it
// has no way to reach a key that is not there).
//
// MUTATION: make the absence of sections an error rather than a fallback, and the
// first half fails; drop the declared-absent state, and the second half still
// holds — which is the point of checking both.
func TestToolRecreationShapeResolvesNothing(t *testing.T) {
	collected := map[string]interface{}{
		"validation_result": map[string]interface{}{"clean_html": "<section>a whole tool</section>"},
		"tool_recreation":   map[string]interface{}{"response": map[string]interface{}{"html": "…"}},
	}

	// declared
	sections, field, origin := resolveSectionsForTest(
		map[string]interface{}{expectsNoSectionsMetadataKey: true, "html_field": "validation_result.clean_html"}, collected)
	if origin != metadataOriginDeclaredAbsent || field != "" || len(sections) != 0 {
		t.Fatalf("declared-absent caller resolved %d sections at %q (%s)", len(sections), field, origin)
	}

	// undeclared: the default is consulted and finds nothing, which is the same
	// outcome by a different route
	sections, field, origin = resolveSectionsForTest(map[string]interface{}{"html_field": "validation_result.clean_html"}, collected)
	if origin != metadataOriginDefault || field != defaultSectionsMetadataField {
		t.Fatalf("got (%q, %q), want the default", field, origin)
	}
	if len(sections) != 0 {
		t.Fatalf("the default path resolved %d sections on a shape that has no page_content", len(sections))
	}
}

// TestShouldReportContentDataLoss pins the report's predicate. Each row names the
// mutation it catches.
func TestShouldReportContentDataLoss(t *testing.T) {
	withData := []SectionData{{ComponentName: "hero", ContentData: map[string]interface{}{"headline": "x"}}}
	withoutData := []SectionData{{ComponentName: "hero"}, {ComponentName: "article-body"}}

	cases := []struct {
		name     string
		origin   string
		existing int
		sections []SectionData
		want     bool
		mutation string
	}{
		{
			name:   "the 194 signature: the page had structured content and this save has none",
			origin: metadataOriginDefault, existing: 3, sections: withoutData, want: true,
			mutation: "invert the incoming==0 test and this stops firing on the only case it exists for",
		},
		{
			name: "silent when the incoming set carries content", origin: metadataOriginConfigured,
			existing: 3, sections: withData, want: false,
			mutation: "drop the countSectionsWithContentData test and every healthy save reports a loss",
		},
		{
			name: "silent when the page never had structured content", origin: metadataOriginDefault,
			existing: 0, sections: withoutData, want: false,
			mutation: "drop the existing>0 test and every first build of every page reports a loss",
		},
		{
			name: "silent for a caller that declares it has none by design", origin: metadataOriginDeclaredAbsent,
			existing: 3, sections: withoutData, want: false,
			mutation: "drop the declared-absent exemption and tool-recreation-handler reports on every run",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldReportContentDataLoss(tc.origin, tc.existing, tc.sections); got != tc.want {
				t.Fatalf("got %v, want %v — %s", got, tc.want, tc.mutation)
			}
		})
	}
}

// TestCountExistingRowsWithContentDataIsScopedToWritableDeployedRows pins the
// query the report's numerator comes from. A locked row is not this save's to
// replace, so counting it would report a loss that is not happening.
//
// MUTATION: drop the pageComponentAgentWritableSQL clause, the build_status
// filter or the IS NOT NULL and the expectation no longer matches, because
// sqlmock matches the SQL text.
func TestCountExistingRowsWithContentDataIsScopedToWritableDeployedRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT COUNT(*) FROM page_components
		WHERE page_id = $1
		  AND build_status = 'deployed'
		  AND content_data IS NOT NULL
		  AND ` + pageComponentAgentWritableSQL(""))).
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	p := ActionParams{DB: db, Logger: zap.NewNop()}
	if got := countExistingRowsWithContentData(context.Background(), p, pageID); got != 3 {
		t.Fatalf("count = %d, want 3", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestCountExistingRowsWithContentDataFailsQuiet: a counting failure must
// suppress the report, never invent one. The report is not a guard, so a broken
// count must not become a false alarm on a healthy save.
//
// MUTATION: return a non-zero sentinel on error and this fails.
func TestCountExistingRowsWithContentDataFailsQuiet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT").WillReturnError(errors.New("connection reset"))

	p := ActionParams{DB: db, Logger: zap.NewNop()}
	if got := countExistingRowsWithContentData(context.Background(), p, uuid.New()); got != 0 {
		t.Fatalf("a failed count returned %d, want 0 (the report must stay silent)", got)
	}
}
