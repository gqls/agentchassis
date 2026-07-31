package actions

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// bugs_open/092 — the page writer is never told which pages exist.
//
// `prepare_link_context` looked for its page list only in collected_data, at
// four paths none of which has ever existed on page-content-writer's
// orchestration. It returned an empty list on 26 of 26 runs, produced an empty
// constraint text, and the prompt template's {{if}} guard then removed the whole
// "## Internal Linking" block — so the model wrote links with no idea what
// existed. Every layer was silent.
//
// WHAT THESE TESTS PIN, and why each one is here rather than a happy-path smoke
// test: the failure mode of this action is producing NOTHING and returning nil.
// A test that only asserts "no error" passes against the bug. So every case
// below asserts a positive artefact — the constraint text's content, the source
// label, the degraded flag, or the agent_error_log INSERT's own arguments.
//
// Mutation-checked 2026-07-31 (each assertion was confirmed to fail against the
// pre-fix behaviour it guards; the specific mutation is named on each test).

// writerRunCollectedData reproduces the SHAPE of a real page-content-writer
// orchestration, taken from the live rows: input_data.site_id present, and
// db_sync / site_record / top-level site_id all ABSENT. Using the real shape is
// the point — a test that helpfully supplies site_record would pass while the
// live path still resolved nothing.
func writerRunCollectedData(siteID uuid.UUID) map[string]interface{} {
	return map[string]interface{}{
		"input_data": map[string]interface{}{
			"site_id": siteID.String(),
			"domain":  "robot-hands.com",
		},
		"render_context": map[string]interface{}{
			"company_name":  "Robot Hands",
			"primary_color": "#123456",
		},
	}
}

// ---------------------------------------------------------------------------
// 1. The fix itself: the list comes from the database
// ---------------------------------------------------------------------------

// MUTATION: delete the database branch in PrepareLinkContextAction and this
// fails on page_count 0 and source "none".
func TestPrepareLinkContextLoadsPagesFromTheDatabase(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta("FROM pages")).
		WithArgs(siteID, maxLinkablePagesInPrompt+1).
		WillReturnRows(sqlmock.NewRows([]string{"name", "url", "title", "description"}).
			AddRow("index", "/index.html", "Home", "").
			AddRow("contact", "/contact.html", "", "Get in touch").
			AddRow("learning-center", "/learning-center.html", "", ""))

	params := ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process", StepName: "prepare_link_context"},
		StepConfig:       models.Step{Config: map[string]interface{}{"enabled": true, "pages_field": "db_sync.pages", "max_links_per_section": float64(3)}},
		CollectedData:    writerRunCollectedData(siteID),
	}

	out, err := PrepareLinkContextAction(context.Background(), params)
	if err != nil {
		t.Fatalf("action returned error: %v", err)
	}
	res := out.(map[string]interface{})

	if got := res["page_count"].(int); got != 3 {
		t.Errorf("page_count = %d, want 3 — the DB list did not reach the output", got)
	}
	if got := res["source"].(string); got != linkSourceDatabase {
		t.Errorf("source = %q, want %q", got, linkSourceDatabase)
	}
	if !res["db_consulted"].(bool) {
		t.Error("db_consulted = false — the authority was not read")
	}
	if res["degraded"].(bool) {
		t.Error("degraded = true on a successful load")
	}

	text := res["link_constraint_text"].(string)
	for _, want := range []string{"/index.html", "/contact.html", "/learning-center.html"} {
		if !strings.Contains(text, want) {
			t.Errorf("constraint text is missing %q — the writer is not being told about it:\n%s", want, text)
		}
	}
	// The title fallback must humanise the name, not print the slug.
	if !strings.Contains(text, "(Learning Center)") {
		t.Errorf("expected the humanised title fallback, got:\n%s", text)
	}
	// The template already emits "## Internal Linking"; this text must not add
	// a second heading one line below it.
	if strings.Contains(text, "## Internal Links") {
		t.Errorf("constraint text re-emits a heading the prompt template already writes:\n%s", text)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expected the pages query to be issued: %v", err)
	}
}

// The site id lives at input_data.site_id on every real writer run, and the
// shared extractSiteID does not look there. This is the single point on which
// the whole fix turns.
//
// MUTATION: drop "input_data.site_id" from resolveLinkContextSiteID's list and
// this fails — no query is issued and the run is degraded.
func TestPrepareLinkContextResolvesSiteIDFromInputData(t *testing.T) {
	siteID := uuid.New()
	params := ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData:    writerRunCollectedData(siteID),
	}
	if got := resolveLinkContextSiteID(params); got != siteID.String() {
		t.Errorf("resolveLinkContextSiteID = %q, want %q — the live writer shape does not resolve", got, siteID)
	}
}

// ---------------------------------------------------------------------------
// 2. The fail-open that made the bug invisible: an empty list said NOTHING
// ---------------------------------------------------------------------------

// A site with no pages is a CORRECT empty list, and the writer must be told so
// explicitly. Returning "" let the consuming template's {{if}} guard elide the
// section, which is how a writer with no information came to be given no
// instruction.
//
// MUTATION: restore `if len(pages) == 0 { return "" }` and this fails on both
// the empty text and the missing instruction.
func TestPrepareLinkContextEmptySiteInstructsNoInternalLinks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta("FROM pages")).
		WithArgs(siteID, maxLinkablePagesInPrompt+1).
		WillReturnRows(sqlmock.NewRows([]string{"name", "url", "title", "description"}))

	params := ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData:    writerRunCollectedData(siteID),
	}

	out, err := PrepareLinkContextAction(context.Background(), params)
	if err != nil {
		t.Fatalf("action returned error: %v", err)
	}
	res := out.(map[string]interface{})

	text := res["link_constraint_text"].(string)
	if text == "" {
		t.Fatal("empty constraint text — this is the fail-open the bug is about")
	}
	if !strings.Contains(text, "Do NOT create any internal links") {
		t.Errorf("empty list must instruct explicitly, got:\n%s", text)
	}

	// The two causes of an empty list have opposite remedies and must not look
	// alike: this one is "the site has none", NOT "I could not find out".
	if res["degraded"].(bool) {
		t.Error("degraded = true, but the database WAS consulted and the site genuinely has no pages")
	}
	if !res["db_consulted"].(bool) {
		t.Error("db_consulted = false after a successful query")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("query expectations: %v", err)
	}
}

// ---------------------------------------------------------------------------
// 3. Unavailable ≠ empty, and it leaves a durable record
// ---------------------------------------------------------------------------

// MUTATION: delete the recordLinkContextUnavailable call and this fails on the
// unmet INSERT expectation; weaken the args and it fails on the pinned
// error_code/severity.
func TestPrepareLinkContextRecordsUnavailableWhenSiteIDIsUnresolvable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// The row is the point, so its two identifying columns are pinned by value:
	// error_code (what a dashboard filters on) and severity (which distinguishes
	// "degraded to no links at all" from "fell back to another source").
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_error_log")).
		WithArgs(
			sqlmock.AnyArg(), // site_id (empty — that IS the failure)
			sqlmock.AnyArg(), // agent_type
			sqlmock.AnyArg(), // step_name
			sqlmock.AnyArg(), // error_message
			linkContextUnavailableCode,
			"error",
			sqlmock.AnyArg(), // context json
			sqlmock.AnyArg(), // orchestration_id
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		AgentType:        "page-content-writer",
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process", StepName: "prepare_link_context"},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		// No site id anywhere — the shape that cannot be resolved.
		CollectedData: map[string]interface{}{"render_context": map[string]interface{}{}},
	}

	out, err := PrepareLinkContextAction(context.Background(), params)
	if err != nil {
		t.Fatalf("action returned error: %v", err)
	}
	res := out.(map[string]interface{})

	if !res["degraded"].(bool) {
		t.Error("degraded = false, but the page list could not be established")
	}
	if res["db_consulted"].(bool) {
		t.Error("db_consulted = true, but no query was possible")
	}
	if !strings.Contains(res["reason"].(string), "site_id") {
		t.Errorf("reason does not name the cause: %q", res["reason"])
	}
	if !strings.Contains(res["link_constraint_text"].(string), "Do NOT create any internal links") {
		t.Error("a run that knows nothing must still instruct explicitly")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expected a durable agent_error_log row: %v", err)
	}
}

// A query failure is NOT an empty site. Treating it as one would tell the writer
// "this site has no pages" on a site with a hundred of them.
//
// MUTATION: return (nil, 0, nil) instead of an error from loadLinkablePages and
// this fails — degraded goes false and no record is written.
func TestPrepareLinkContextTreatsAQueryFailureAsUnavailableNotEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta("FROM pages")).
		WillReturnError(errors.New("pq: canceling statement due to statement timeout"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_error_log")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData:    writerRunCollectedData(siteID),
	}

	out, err := PrepareLinkContextAction(context.Background(), params)
	if err != nil {
		t.Fatalf("action returned error: %v", err)
	}
	res := out.(map[string]interface{})

	if !res["degraded"].(bool) {
		t.Error("a failed page query must be degraded, not an empty site")
	}
	if res["db_consulted"].(bool) {
		t.Error("db_consulted must be false when the query errored")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expected the failure to be recorded: %v", err)
	}
}

// ---------------------------------------------------------------------------
// 4. Trap 2: a URL is never synthesised from a name
// ---------------------------------------------------------------------------

// The old code built "/" + name + ".html" for a page with no url. On a
// dir/index.html site, or a fleet that does not serve .html, that is a
// confidently wrong address handed to the writer as though it were real — the
// bugs_closed/029 failure mode one layer upstream.
//
// MUTATION: restore the synthesis in extractPagesForLinking and this fails on
// the presence of "/services.html".
func TestPrepareLinkContextNeverSynthesisesAURLFromAName(t *testing.T) {
	pages := extractPagesForLinking(
		map[string]interface{}{
			"db_sync": map[string]interface{}{
				"pages": []interface{}{
					map[string]interface{}{"name": "services", "title": "Services"}, // no url
					map[string]interface{}{"name": "about", "url": "/about/index.html"},
				},
			},
		},
		map[string]interface{}{"pages_field": "db_sync.pages"},
		zap.NewNop(),
	)

	if len(pages) != 1 {
		t.Fatalf("kept %d page(s), want 1 — the url-less entry must be dropped, not invented: %+v", len(pages), pages)
	}
	if pages[0].URL != "/about/index.html" {
		t.Errorf("url = %q, want the STORED url verbatim", pages[0].URL)
	}
	for _, p := range pages {
		if strings.Contains(p.URL, "services") {
			t.Errorf("a url was synthesised for a page that has none: %q", p.URL)
		}
	}
}

// ---------------------------------------------------------------------------
// 5. Precedence, and the one case where collected_data still wins
// ---------------------------------------------------------------------------

// The database is authoritative because it is the same table, under the same
// predicate, that validate_page_content uses to decide what is a phantom_link.
// If the writer were constrained from a different source the two could disagree,
// and a writer obeying its instructions would get its links flagged.
//
// MUTATION: try collected_data first and this fails — /stale.html appears.
func TestPrepareLinkContextPrefersTheDatabaseOverCollectedData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta("FROM pages")).
		WillReturnRows(sqlmock.NewRows([]string{"name", "url", "title", "description"}).
			AddRow("current", "/current.html", "Current", ""))

	collected := writerRunCollectedData(siteID)
	collected["db_sync"] = map[string]interface{}{
		"pages": []interface{}{map[string]interface{}{"name": "stale", "url": "/stale.html"}},
	}

	params := ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData:    collected,
	}

	out, err := PrepareLinkContextAction(context.Background(), params)
	if err != nil {
		t.Fatalf("action returned error: %v", err)
	}
	res := out.(map[string]interface{})
	text := res["link_constraint_text"].(string)

	if res["source"].(string) != linkSourceDatabase {
		t.Errorf("source = %q, want the database to win", res["source"])
	}
	if strings.Contains(text, "/stale.html") {
		t.Errorf("collected_data overrode the authoritative list:\n%s", text)
	}
	if !strings.Contains(text, "/current.html") {
		t.Errorf("the database list did not reach the prompt:\n%s", text)
	}
}

// When the database cannot be consulted at all, the configured field is still
// honoured — a workflow that declares a page list does not lose it.
//
// MUTATION: delete the fallback branch and this fails on page_count 0.
func TestPrepareLinkContextFallsBackToCollectedDataWithoutADatabase(t *testing.T) {
	params := ActionParams{
		DB:               nil, // no handle at all
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		StepConfig:       models.Step{Config: map[string]interface{}{"pages_field": "site_plan.pages"}},
		CollectedData: map[string]interface{}{
			"site_plan": map[string]interface{}{
				"pages": []interface{}{map[string]interface{}{"name": "planned", "url": "/planned.html"}},
			},
		},
	}

	out, err := PrepareLinkContextAction(context.Background(), params)
	if err != nil {
		t.Fatalf("action returned error: %v", err)
	}
	res := out.(map[string]interface{})

	if res["page_count"].(int) != 1 {
		t.Errorf("page_count = %v, want 1 from the declared source", res["page_count"])
	}
	if res["source"].(string) != linkSourceCollectedData {
		t.Errorf("source = %q, want %q", res["source"], linkSourceCollectedData)
	}
	if res["degraded"].(bool) {
		t.Error("degraded = true although a usable list was found")
	}
	// The reason must still say the authority was not read — a fallback that
	// reads as a clean load is how a stale list becomes invisible.
	if !strings.Contains(res["reason"].(string), "NOT consulted") {
		t.Errorf("reason hides that the database was never read: %q", res["reason"])
	}
}

// ---------------------------------------------------------------------------
// 6. The explicit opt-out still says nothing, which is the one correct silence
// ---------------------------------------------------------------------------

func TestPrepareLinkContextDisabledEmitsNothing(t *testing.T) {
	params := ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		StepConfig:       models.Step{Config: map[string]interface{}{"enabled": false}},
		CollectedData:    map[string]interface{}{},
	}
	out, err := PrepareLinkContextAction(context.Background(), params)
	if err != nil {
		t.Fatalf("action returned error: %v", err)
	}
	res := out.(map[string]interface{})
	if res["link_constraint_text"].(string) != "" {
		t.Error("an explicit opt-out must produce no prompt block at all")
	}
	if res["source"].(string) != linkSourceDisabled {
		t.Errorf("source = %q, want %q", res["source"], linkSourceDisabled)
	}
}
