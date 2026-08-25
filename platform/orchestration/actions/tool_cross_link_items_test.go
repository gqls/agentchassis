// FILE: platform/orchestration/actions/tool_cross_link_items_test.go
//
// bugs_open/029. The defect these tests stand against: a tool page's URL was
// CONSTRUCTED from the tool's function name at suggestion time, and matched no
// page on any of the three shapes this platform actually produces.
//
// The emitter itself needs a live DB (it reads pages/site_work_items), so what
// is unit-testable is the boundary around it: the shapes related_pages arrives
// in, and the build_status predicate that decides whether a link is safe to
// write now or must wait behind the page's build.

package actions

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

func TestRelatedPagesFromSpec(t *testing.T) {
	cases := []struct {
		name string
		in   interface{}
		want []string
	}{
		{
			// The shape it actually arrives in: spec jsonb decoded into
			// map[string]interface{} by the work item loader.
			name: "decoded jsonb array",
			in:   []interface{}{"services", "capabilities"},
			want: []string{"services", "capabilities"},
		},
		{
			name: "already a string slice",
			in:   []string{"services"},
			want: []string{"services"},
		},
		{
			name: "json-encoded string",
			in:   `["services","about"]`,
			want: []string{"services", "about"},
		},
		{
			// A suggestion with no related_pages is normal, not an error:
			// the emitter logs and does nothing.
			name: "absent",
			in:   nil,
			want: nil,
		},
		{
			name: "wrong type is not a panic",
			in:   42,
			want: nil,
		},
		{
			// Non-string members are dropped rather than stringified — a
			// number here would resolve against no page anyway.
			name: "mixed members",
			in:   []interface{}{"services", 7, "", "about"},
			want: []string{"services", "about"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := relatedPagesFromSpec(tc.in)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v (len %d), want %v (len %d)", got, len(got), tc.want, len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("index %d: got %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestToolPageLive(t *testing.T) {
	// The vocabulary is exactly these three (checked against pages.build_status
	// fleet-wide 2026-07-25: deployed 363, needs_rebuild 31, planned 26).
	// needs_rebuild counts as live: the page was deployed and is queued for a
	// refresh, so the link resolves today. planned does NOT: linking to it is
	// the bug.
	if !toolPageLive("deployed") {
		t.Error("deployed must count as live")
	}
	if !toolPageLive("needs_rebuild") {
		t.Error("needs_rebuild must count as live — the page is served while it waits")
	}
	if toolPageLive("planned") {
		t.Error("planned must NOT count as live — that is the 404 this bug is about")
	}
	if toolPageLive("") {
		t.Error("an unreadable build_status must not be treated as live")
	}
}

// TestCrossLinkEmitDecision pins Guard 2's decision table — bugs_open/353.
//
// This exists because the defect it stands against lived in a branch NO unit
// test could reach: the decision was inline in a DB-dependent function, so the
// only tests possible were of its inputs (`toolPageLive`, `relatedPagesFromSpec`
// above), and those passed throughout the 19 days the guard was silently
// withholding every new tool's cross-links. Pinning inputs is not pinning a
// guard; the decision is extracted so it can be CALLED.
//
// The two rows that matter are the last two: identical except for the caller's
// promise, and they must differ. If they ever agree, either the opt-in has been
// made the default (unsafe — the owner's 2026-08-02 ruling says the unsafe side
// defaults OFF) or it has become inert (the 353 defect is back).
func TestCrossLinkEmitDecision(t *testing.T) {
	cases := []struct {
		name                  string
		pageLive              bool
		gateItemFound         bool
		buildEnqueuedByCaller bool
		want                  crossLinkDecision
	}{
		{"a served page needs no gate at all", true, false, false, crossLinkEmitUngated},
		{"served, and the promise is irrelevant", true, false, true, crossLinkEmitUngated},
		{"not live but a build item exists: gate on it", false, true, false, crossLinkEmitGated},
		{"a gate item outranks the caller's promise — depends_on is stricter", false, true, true, crossLinkEmitGated},
		{"THE 353 CASE: not live, no gate item, no promise -> withhold", false, false, false, crossLinkWithhold},
		{"THE 353 FIX: not live, no gate item, caller owns the build -> emit", false, false, true, crossLinkEmitUngated},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := crossLinkEmitDecision(tc.pageLive, tc.gateItemFound, tc.buildEnqueuedByCaller); got != tc.want {
				t.Errorf("crossLinkEmitDecision(live=%v, gate=%v, promise=%v) = %v, want %v",
					tc.pageLive, tc.gateItemFound, tc.buildEnqueuedByCaller, got, tc.want)
			}
		})
	}
}

// TestCrossLinkOptInDefaultsToWithhold is the one assertion that cannot be made
// by the table above: that a caller which says NOTHING gets the safe branch. A
// zero-valued request is exactly what a new caller written by someone who never
// read this file produces, and it must not be granted the permissive arm.
func TestCrossLinkOptInDefaultsToWithhold(t *testing.T) {
	var req toolCrossLinkRequest // zero value: the forgetful caller
	if req.pageBuildIsEnqueuedByThisWorkflow {
		t.Fatal("the opt-in must default to false — the unsafe side is the default per the 2026-08-02 shared-seam ruling")
	}
	if got := crossLinkEmitDecision(false, false, req.pageBuildIsEnqueuedByThisWorkflow); got != crossLinkWithhold {
		t.Errorf("a zero-valued request on an unbuilt page must withhold, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// The CALL SITE. bugs_open/353, council round 2 (corr 642ecc3c, editquality).
// ---------------------------------------------------------------------------
//
// TestCrossLinkCallSitePassesTheRealPageLive exists because of an objection the
// two tests above CANNOT answer, and the reviewer was right to raise it: pinning
// crossLinkEmitDecision's table proves the function is correct, not that the
// production caller hands it real values. An earlier cut of this fix called it
// as `crossLinkEmitDecision(false, ...)` — a literal — which made the pageLive
// branch dead in production while every test above stayed green. A pure
// function's table can never see that; only an assertion through the DB-facing
// caller can.
//
// The setup is the discriminating one: the tool page is SERVED, and the opt-in
// is OFF. Correct wiring reads build_status='deployed' → pageLive TRUE → the
// first branch → emit. The literal-false wiring reaches the third branch with
// no gate item and no promise → withhold, and creates NOTHING.
//
// The assertion is the EFFECT — one item created — not the absence of a query.
// A test that asserted "the gate query was not issued" would pass vacuously the
// moment the call errored for any other reason (LANDMINES: a test asserting a
// query is NOT issued passes vacuously). Here the return value is the claim: 1
// on correct wiring, 0 on the defect this fix removes.
func TestCrossLinkCallSitePassesTheRealPageLive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	toolPageID := uuid.New()
	relatedPageID := uuid.New()

	// Guard 2's read: the tool page is already deployed.
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(build_status, '') FROM pages WHERE id = $1`)).
		WithArgs(toolPageID).
		WillReturnRows(sqlmock.NewRows([]string{"build_status"}).AddRow("deployed"))

	// The gate-item lookup is deliberately UNREGISTERED. It is wasted work on a
	// served page, so correct code never issues it. This is not the assertion —
	// it is what makes the defect fail loudly rather than silently: the
	// literal-false path issues it, gets an error, and falls to withhold.

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT name, id FROM pages`)).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"name", "id"}).AddRow("services", relatedPageID))

	// insertWorkItem's two-strike history check. Registered so it SUCCEEDS with
	// a history that suppresses nothing (0 priors), rather than erroring and
	// being swallowed into a false negative.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age_hours"}).AddRow(0, 0.0))

	mock.ExpectBegin()
	expectWorkItemDoorGenericPage(mock) // bugs_open/333: writeWorkItem consults the policy door here
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO site_work_items")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	params := ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		CollectedData:    map[string]interface{}{},
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "save_tool"},
	}

	created := emitToolCrossLinkItems(context.Background(), params, zap.NewNop(), toolCrossLinkRequest{
		siteID:       siteID,
		toolFunction: "tool-worked-case",
		toolName:     "Worked Case",
		toolDesc:     "a tool",
		toolPageID:   toolPageID,
		toolPageURL:  "/tools/worked-case/",
		relatedPages: []string{"services"},
		emittedBy:    "tool-generator",
		// THE OPT-IN IS OFF. The emit below is licensed by the page being
		// served, and by nothing else — so this test fails if the call site
		// stops passing the real pageLive, whatever the opt-in is doing.
		pageBuildIsEnqueuedByThisWorkflow: false,
	})

	if created != 1 {
		t.Fatalf("emitToolCrossLinkItems created %d items for a SERVED tool page with the opt-in OFF, want 1. "+
			"That is the literal-false defect: the call site is not passing the real pageLive to "+
			"crossLinkEmitDecision, so a deployed page falls through to the withhold arm and every "+
			"cross-link for an already-live tool page is dropped", created)
	}
}

// bugs_open/330 §12 + the owner ruling of 2026-08-24. The defect this stands
// against is not a crash — it is a silence. A tool ordered by hand carries no
// `related_pages`, so no cross-mention is ever written, the build succeeds, the
// page deploys, and the only trace is an `info` row. Measured that day: 0 of 58
// hand-filed add_tool items had ever carried the key, against 11 of 11 from
// tool-suggester.
//
// The remedy is a SECOND declared source — what the workflow's picker step
// chose when the requester named nothing. The two properties that matter are
// therefore an ORDER and a DECLARATION, and both are tested here through the
// real extractor rather than by calling the resolver with hand-built inputs:
// an undeclared key is dropped by ExtractActionInputs before the resolver ever
// sees it, which is exactly how this class of fix silently does nothing
// (bugs_closed/336). Driving the production spec is what makes that visible.
func TestRelatedPagesPrecedenceAndSource(t *testing.T) {
	// Both keys as the live config wires them: optional-explicit (`?`), each
	// naming ONE path. Migration 516 — neither may fall through to the
	// whole-tree search, which is what handed nine tools one tool's pages.
	// The four Required keys are scaffolding only — the extractor refuses
	// without them, and refusing is correct. Only the two optional wires below
	// are under test.
	config := map[string]interface{}{
		"site_id":                 "input_data.spec.site_id",
		"html_content":            "input_data.spec.html",
		"function":                "input_data.spec.function",
		"display_name":            "input_data.spec.name",
		"related_pages?":          "input_data.spec.related_pages",
		"related_pages_fallback?": "suggest_related_pages.result",
	}

	specPages := []interface{}{"learn-bayes", "learn-p-values"}
	pickedPages := `["services", "about"]`

	// spec builds the add_tool item's spec object with the required fields
	// always present, plus whatever the case adds.
	spec := func(extra map[string]interface{}) map[string]interface{} {
		out := map[string]interface{}{
			"site_id":  "11111111-1111-1111-1111-111111111111",
			"html":     "<div></div>",
			"function": "tool-worked-case",
			"name":     "Worked Case",
		}
		for k, v := range extra {
			out[k] = v
		}
		return out
	}

	cases := []struct {
		name       string
		collected  map[string]interface{}
		wantPages  []string
		wantSource string
		why        string
	}{
		{
			name: "requester named pages and the picker also ran: the REQUESTER wins",
			collected: map[string]interface{}{
				"input_data":            map[string]interface{}{"spec": spec(map[string]interface{}{"related_pages": specPages})},
				"suggest_related_pages": map[string]interface{}{"result": pickedPages},
			},
			wantPages:  []string{"learn-bayes", "learn-p-values"},
			wantSource: relatedPagesSourceSpec,
			why: "a picker that can overrule an explicit choice is worse than no picker: it makes " +
				"the field the requester filled in silently advisory",
		},
		{
			name: "requester named nothing: the picker's choice is used, and is marked as picked",
			collected: map[string]interface{}{
				"input_data":            map[string]interface{}{"spec": spec(nil)},
				"suggest_related_pages": map[string]interface{}{"result": pickedPages},
			},
			wantPages:  []string{"services", "about"},
			wantSource: relatedPagesSourceSuggested,
			why:        "this is the whole point of the change — the 0-of-58 case",
		},
		{
			name: "picker never ran (step unwired, or it failed to error_step): NOT an answer",
			collected: map[string]interface{}{
				"input_data": map[string]interface{}{"spec": spec(nil)},
			},
			wantPages:  nil,
			wantSource: relatedPagesSourceNoPicker,
			why: "the council's objection to migration 602: this used to be indistinguishable from " +
				"the picker running and honestly declining, so a picker failing on EVERY build read " +
				"as a site with no matching pages",
		},
		{
			name: "picker ran and DECLINED (empty list): a real answer, and a correct one",
			collected: map[string]interface{}{
				"input_data":            map[string]interface{}{"spec": spec(nil)},
				"suggest_related_pages": map[string]interface{}{"result": `[]`},
			},
			wantPages:  nil,
			wantSource: relatedPagesSourcePickerDeclined,
			why: "prompt rule 5 makes an empty answer correct and preferred to a weak match — a site " +
				"may genuinely have no page a tool belongs on, and inventing one is bugs_open/330 again",
		},
		{
			name: "picker ran and returned PROSE: unusable, and not the same event as declining",
			collected: map[string]interface{}{
				"input_data":            map[string]interface{}{"spec": spec(nil)},
				"suggest_related_pages": map[string]interface{}{"result": "Sorry, none of these pages fit."},
			},
			wantPages:  nil,
			wantSource: relatedPagesSourcePickerUnusable,
			why: "a model that stops obeying the output format is a prompt or model fault and needs " +
				"chasing; an honest refusal needs nothing. Collapsing them hides the first for ever",
		},
		{
			name: "picker ran and returned an empty ARRAY value (not a string): still declined",
			collected: map[string]interface{}{
				"input_data":            map[string]interface{}{"spec": spec(nil)},
				"suggest_related_pages": map[string]interface{}{"result": []interface{}{}},
			},
			wantPages:  nil,
			wantSource: relatedPagesSourcePickerDeclined,
			why: "the value's Go type depends on how the step's output was stored; a list type IS a " +
				"list however it arrived, so both spellings must classify the same way",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inputs, err := datahelpers.ExtractActionInputs(
				tc.collected, config, CreateToolComponentInputSpec, zap.NewNop())
			if err != nil {
				t.Fatalf("ExtractActionInputs: %v", err)
			}

			pages, source := relatedPagesFromInputs(inputs, tc.collected)

			if len(pages) != len(tc.wantPages) {
				t.Fatalf("got %d pages %v, want %d %v — %s", len(pages), pages, len(tc.wantPages), tc.wantPages, tc.why)
			}
			for i := range pages {
				if pages[i] != tc.wantPages[i] {
					t.Fatalf("page[%d] = %q, want %q — %s", i, pages[i], tc.wantPages[i], tc.why)
				}
			}
			if source != tc.wantSource {
				t.Fatalf("source = %q, want %q — %s.\n"+
					"The source is written into agent_error_log and every emitted item, and it is the "+
					"ONLY thing that separates the five outcomes: pages from the requester, pages from "+
					"the picker, a picker that never ran, one that ran and declined, and one whose "+
					"answer could not be read. Collapse any two and the census that reads them stops "+
					"being able to tell a working mechanism from a quiet failure.",
					source, tc.wantSource, tc.why)
			}
		})
	}
}

// The deploy path carries the identical wire to the identical helper. Migration
// 516's own lesson, stated in its scope section: leaving one of two identical
// armed wires unmarked is how a class gets "fixed" and then rediscovered. This
// pins BOTH declarations so the deployer half cannot be dropped in a later edit
// while every other test stays green.
func TestBothToolActionsDeclareTheFallbackKey(t *testing.T) {
	for _, tc := range []struct {
		action string
		spec   datahelpers.ActionInputSpec
	}{
		{"create_tool_component", CreateToolComponentInputSpec},
		{"deploy_tool_to_site", DeployToolToSiteInputSpec},
	} {
		var hasPages, hasFallback bool
		for _, k := range tc.spec.Optional {
			switch k {
			case "related_pages":
				hasPages = true
			case "related_pages_fallback":
				hasFallback = true
			}
		}
		if !hasPages || !hasFallback {
			t.Errorf("%s Optional = %v; needs BOTH related_pages and related_pages_fallback. "+
				"An undeclared key is dropped by the extractor before the resolver sees it, so the "+
				"picker would run, cost a model call, and change nothing (bugs_closed/336)",
				tc.action, tc.spec.Optional)
		}
	}
}
