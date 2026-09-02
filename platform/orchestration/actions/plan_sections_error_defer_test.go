// FILE: platform/orchestration/actions/plan_sections_error_defer_test.go
//
// bugs_open/444 — an ERRORED resolve of a REQUIRED (or min_items-declaring)
// query field must DEFER the section, not leave it ready. Before this fix the
// error branch logged one Warn and proceeded, so directory-listing built with
// only its LLM headline on every site whose exporter config is missing —
// resolveBusinessDirectory's DESIGNED loud failure (bugs_open/206) evaporating
// into a log line. The mutation half of these tests is the first one: it
// asserts "deferred" on exactly the fixture that used to produce "ready" (the
// stored orchestration d0a858be-… artefact shape).
//
// on_missing is still never consulted for errors (bugs_open/054's rule): the
// component below declares on_missing=skip_section, and the test asserts the
// section is DEFERRED, not skipped — error and no-data stay distinguishable.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func directoryListingComponent() componentInfo {
	return componentInfo{
		ID:       uuid.New().String(),
		Function: "directory-listing",
		InputSchema: map[string]interface{}{
			"fields": map[string]interface{}{
				"entries": map[string]interface{}{
					"type":       "array",
					"source":     "query.business_directory",
					"required":   true,
					"min_items":  float64(1),
					"on_missing": "skip_section",
				},
				"headline": map[string]interface{}{
					"type":     "text",
					"source":   "llm",
					"required": true,
				},
			},
		},
		Raw: map[string]interface{}{"component_level": "section"},
	}
}

// The 444 fixture: no directory-json-exporter config row, so the resolver
// returns its designed misconfiguration ERROR (not an empty list). The section
// must defer — before the fix it came back "ready" with only the LLM headline.
func TestPlanSection_RequiredQuerySourceErrors_Defers(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// Config lookup finds nothing → resolveBusinessDirectory errors by design.
	mock.ExpectQuery("directory-json-exporter").
		WillReturnRows(sqlmock.NewRows([]string{"vertical", "business_type_ilike"}))
	// The regeneration carry then looks for a stored value; none exists (first
	// build). Any query it makes is answered empty-handed by the mock erroring,
	// which storedFieldValue treats as not-found — either way no carry.

	resolver := newSourceResolver(uuid.New(), db, zap.NewNop(), "channels-directory-index")
	item := planSection(context.Background(), "directory-listing", sectionRef{}, directoryListingComponent(), resolver, zap.NewNop())

	if item.Status != "deferred" {
		t.Fatalf("required query source errored: status = %q (reason %q), want deferred — a hollow ready section is the bugs_open/444 artefact", item.Status, item.Reason)
	}
	if item.Status == "skipped" {
		t.Fatal("error must not be routed into on_missing=skip_section (bugs_open/054)")
	}
	if len(item.Missing) == 0 {
		t.Fatal("the deferred item must carry the missing-field record naming the errored source")
	}
}

// Scope pin: an OPTIONAL field whose source errors keeps today's behaviour
// (field left unresolved / fallback; section stays ready) — the defer applies
// only where the schema says the data is load-bearing.
func TestPlanSection_OptionalQuerySourceErrors_StaysReady(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("directory-json-exporter").
		WillReturnRows(sqlmock.NewRows([]string{"vertical", "business_type_ilike"}))

	comp := directoryListingComponent()
	fields := comp.InputSchema["fields"].(map[string]interface{})
	entries := fields["entries"].(map[string]interface{})
	entries["required"] = false
	delete(entries, "min_items")

	resolver := newSourceResolver(uuid.New(), db, zap.NewNop(), "channels-directory-index")
	item := planSection(context.Background(), "directory-listing", sectionRef{}, comp, resolver, zap.NewNop())

	if item.Status != "ready" {
		t.Fatalf("optional errored source: status = %q, want ready (behaviour unchanged)", item.Status)
	}
	// The disposition is unchanged, but the error must no longer be SILENT:
	// a durable structural-miss record rides the plan item (bugs_open/238's
	// channel; council corr c0990eb3 round 2, bug_historian).
	found := false
	for _, m := range item.StructuralMisses {
		if m.Field == "entries" {
			found = true
		}
	}
	if !found {
		t.Fatal("optional errored source must leave a structural-miss record — a Warn line in a restarting pod is not evidence")
	}
}
