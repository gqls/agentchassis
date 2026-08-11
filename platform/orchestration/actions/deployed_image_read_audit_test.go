package actions

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// Covers deployedImageURL (bugs_open/236, commission item 2): the three readers
// of a deploy_image_asset result no longer fail silently when the result carries
// no URL.
//
// The two properties worth pinning are opposites, and only one of them is the
// obvious one:
//
//   - a present-but-unusable container is recorded DURABLY (agent_error_log is
//     the only sink that outlives the awaited step's 4-hour state row);
//   - an ABSENT container records NOTHING, because most pages never deploy a
//     hero or logo and an else on the outer guard would file a row per page
//     fleet-wide. That no-op case is tested rather than assumed, and it is the
//     one a mock's own bookkeeping cannot vouch for — see the mutation note in
//     the lane NOTES: flipping the presence guard must make
//     TestDeployedImageURL_AbsentContainerRecordsNothing fail, and it does.
//
// The context payload carries KEYS, NEVER VALUES, and
// TestDeployedImageURL_RecordsKeysNeverValues pins that: the response bodies
// behind these keys hold base64 image payloads.

// capturedString matches any driver value and records it, so a test can assert
// on the context JSON without enumerating all thirteen INSERT arguments.
type capturedString struct{ into *string }

func (c capturedString) Match(v driver.Value) bool {
	if s, ok := v.(string); ok {
		*c.into = s
	}
	return true
}

// the236Shape is the container as measured on the live decisive row
// (bugs_open/236 §1, orchestration 3e46be5b, cookly.uk): the awaited response
// merged in, and none of the three keys DeployImageAssetAction assigned.
func the236Shape() map[string]interface{} {
	return map[string]interface{}{
		"response": map[string]interface{}{
			"data": map[string]interface{}{
				"file_path":      "/assets/images/logo.png",
				"domain":         "cookly.uk",
				"success":        true,
				"commit_message": "Deploy logo image for cookly.uk",
			},
		},
		"response_status":      "complete",
		"response_received_at": "2026-08-09T13:54:37Z",
	}
}

func auditTestParams(t *testing.T, collected map[string]interface{}) (ActionParams, *observer.ObservedLogs, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	core, logs := observer.New(zap.DebugLevel)
	params := ActionParams{
		CollectedData: collected,
		DB:            db,
		Logger:        zap.New(core),
		AgentType:     "site-work-orchestrator",
		CurrentStep:   "build_render_context",
		ExecutionContext: &types.ExecutionContext{
			OrchestrationID: "3e46be5b-8788-447b-9643-e32ae33f601b",
			StepName:        "build_render_context",
			Action:          "execute",
		},
		StepConfig: models.Step{Config: map[string]interface{}{}},
	}
	return params, logs, mock, func() { db.Close() }
}

// missWarnings is the count of the Warn this change adds — the same literal the
// post-roll pod-grep looks for.
func missWarnings(logs *observer.ObservedLogs) int {
	return len(logs.FilterMessageSnippet("carries no usable URL").All())
}

func TestDeployedImageURL_ReturnsURLFromHealthyContainer(t *testing.T) {
	params, logs, mock, done := auditTestParams(t, map[string]interface{}{
		"hero_deployed": map[string]interface{}{
			"image_url":   "/assets/images/hero.png",
			"output_path": "/assets/images/hero.png",
			"size_bytes":  1234,
		},
	})
	defer done()

	got := deployedImageURL(context.Background(), params, "hero_deployed", "image_url", "hero_url", "build_render_context")
	if got != "/assets/images/hero.png" {
		t.Fatalf("url = %q, want the container's image_url", got)
	}
	if n := missWarnings(logs); n != 0 {
		t.Errorf("healthy container warned %d times, want 0", n)
	}
	// No INSERT was expected; sqlmock fails an unexpected Exec, which this
	// helper reports as an Error log. Both are asserted.
	if n := len(logs.FilterMessageSnippet("failed to record").All()); n != 0 {
		t.Errorf("healthy container attempted a durable write")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// THE NO-OP CASE. Most pages deploy no hero and no logo; that must cost nothing.
func TestDeployedImageURL_AbsentContainerRecordsNothing(t *testing.T) {
	params, logs, mock, done := auditTestParams(t, map[string]interface{}{
		"site_record": map[string]interface{}{"domain": "cookly.uk"},
	})
	defer done()

	if got := deployedImageURL(context.Background(), params, "hero_deployed", "image_url", "hero_url", "build_render_context"); got != "" {
		t.Fatalf("url = %q, want empty", got)
	}
	if n := missWarnings(logs); n != 0 {
		t.Errorf("absent container warned %d times, want 0 — no demand, no finding", n)
	}
	// The discriminator: no Exec was attempted at all. sqlmock rejects any
	// unexpected call, so a write attempt would surface as the "failed to
	// record" Error below.
	if n := len(logs.FilterMessageSnippet("failed to record").All()); n != 0 {
		t.Errorf("absent container attempted a durable write — the demand gate is not holding")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

func TestDeployedImageURL_236ShapeIsRecordedDurably(t *testing.T) {
	params, logs, mock, done := auditTestParams(t, map[string]interface{}{
		"logo_deployed": the236Shape(),
	})
	defer done()

	var contextJSON string
	mock.ExpectExec("INSERT INTO agent_error_log").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), deployedImageMissCode, "warning", capturedString{&contextJSON},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if got := deployedImageURL(context.Background(), params, "logo_deployed", "image_url", "logo_url", "assemble_from_library"); got != "" {
		t.Fatalf("url = %q, want empty on the 236 shape", got)
	}
	if n := missWarnings(logs); n != 1 {
		t.Errorf("warned %d times, want 1", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the durable row was not written as expected: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(contextJSON), &payload); err != nil {
		t.Fatalf("context is not JSON: %v", err)
	}
	if payload["container_key"] != "logo_deployed" || payload["wanted_key"] != "image_url" {
		t.Errorf("container/wanted keys wrong: %v / %v", payload["container_key"], payload["wanted_key"])
	}
	// The keys are the diagnostic: this exact list IS the 236 signature.
	keys, _ := json.Marshal(payload["keys_present"])
	if want := `["response","response_received_at","response_status"]`; string(keys) != want {
		t.Errorf("keys_present = %s, want %s (sorted, so two rows of one shape compare equal)", keys, want)
	}
	if payload["fallback_sibling_present"] != false {
		t.Errorf("fallback_sibling_present = %v, want false — no logo_url in collected_data", payload["fallback_sibling_present"])
	}
}

// KEYS, NEVER VALUES. The response bodies behind these keys carry base64 image
// payloads and full adapter responses; a row that copied them would be both
// enormous and a second place for the data to leak.
func TestDeployedImageURL_RecordsKeysNeverValues(t *testing.T) {
	params, _, mock, done := auditTestParams(t, map[string]interface{}{
		"hero_deployed": the236Shape(),
	})
	defer done()

	var contextJSON string
	mock.ExpectExec("INSERT INTO agent_error_log").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), capturedString{&contextJSON},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	deployedImageURL(context.Background(), params, "hero_deployed", "image_url", "hero_url", "build_render_context")

	for _, leaked := range []string{"/assets/images/logo.png", "Deploy logo image", "cookly.uk"} {
		if strings.Contains(contextJSON, leaked) {
			t.Errorf("context payload leaked a VALUE (%q) — it must carry keys only:\n%s", leaked, contextJSON)
		}
	}
}

// The discriminator for bugs_open/236 §5: deploy_image_asset_action.go:404-415
// writes hero_url/logo_url as a sibling key, so "the container lost it" and "the
// URL is lost everywhere" are different failures and the row must say which.
func TestDeployedImageURL_RecordsWhetherTheFallbackSiblingSurvived(t *testing.T) {
	params, _, mock, done := auditTestParams(t, map[string]interface{}{
		"hero_deployed": the236Shape(),
		"hero_url":      "/assets/images/hero.png",
	})
	defer done()

	var contextJSON string
	mock.ExpectExec("INSERT INTO agent_error_log").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), capturedString{&contextJSON},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	deployedImageURL(context.Background(), params, "hero_deployed", "image_url", "hero_url", "build_render_context")

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(contextJSON), &payload); err != nil {
		t.Fatalf("context is not JSON: %v", err)
	}
	if payload["fallback_sibling_present"] != true {
		t.Errorf("fallback_sibling_present = %v, want true — collected_data carries hero_url, so the page is probably fine and the loss is confined to the container", payload["fallback_sibling_present"])
	}
}

func TestDeployedImageURL_NonMapContainerIsRecordedWithItsType(t *testing.T) {
	params, logs, mock, done := auditTestParams(t, map[string]interface{}{
		"hero_deployed": "not-a-map",
	})
	defer done()

	var contextJSON string
	mock.ExpectExec("INSERT INTO agent_error_log").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), capturedString{&contextJSON},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if got := deployedImageURL(context.Background(), params, "hero_deployed", "image_url", "hero_url", "build_render_context"); got != "" {
		t.Fatalf("url = %q, want empty", got)
	}
	if n := missWarnings(logs); n != 1 {
		t.Errorf("warned %d times, want 1 — a wrong-typed container is an anomaly worth a row", n)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(contextJSON), &payload); err != nil {
		t.Fatalf("context is not JSON: %v", err)
	}
	if payload["container_type"] != "string" {
		t.Errorf("container_type = %v, want string", payload["container_type"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// A lost row must never read as a recorded one (log_action_error.go's counted
// best-effort discipline). This is the only surviving copy once the
// orchestration state is pruned, so losing it earns an Error, not a shrug.
func TestDeployedImageURL_LostDurableRowIsLoggedAsAnError(t *testing.T) {
	params, logs, mock, done := auditTestParams(t, map[string]interface{}{
		"hero_deployed": the236Shape(),
	})
	defer done()

	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnError(context.DeadlineExceeded)

	deployedImageURL(context.Background(), params, "hero_deployed", "image_url", "hero_url", "build_render_context")

	if n := len(logs.FilterMessageSnippet("this finding now exists only in this log line").All()); n != 1 {
		t.Errorf("a lost durable row produced %d Error logs, want 1", n)
	}
}

// The third reader, end to end through its real ladder: the two cheaper rungs
// must win WITHOUT filing anything, and only a genuine fall-through to the
// deploy result records.
func TestExtractLogoURLFromParams_CheaperRungsWinAndFileNothing(t *testing.T) {
	for _, tc := range []struct {
		name string
		coll map[string]interface{}
		want string
	}{
		{
			name: "render_context wins",
			coll: map[string]interface{}{
				"render_context": map[string]interface{}{"logo_url": "/from/render_context.png"},
				"logo_deployed":  the236Shape(),
			},
			want: "/from/render_context.png",
		},
		{
			name: "site_record wins",
			coll: map[string]interface{}{
				"site_record":   map[string]interface{}{"logo_url": "/from/site_record.png"},
				"logo_deployed": the236Shape(),
			},
			want: "/from/site_record.png",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			params, logs, mock, done := auditTestParams(t, tc.coll)
			defer done()

			if got := extractLogoURLFromParams(context.Background(), params); got != tc.want {
				t.Fatalf("logo url = %q, want %q", got, tc.want)
			}
			if n := missWarnings(logs); n != 0 {
				t.Errorf("a satisfied ladder warned %d times, want 0", n)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("expectations: %v", err)
			}
		})
	}
}

func TestExtractLogoURLFromParams_FallThroughToBrokenDeployResultRecords(t *testing.T) {
	params, logs, mock, done := auditTestParams(t, map[string]interface{}{
		"logo_deployed": the236Shape(),
	})
	defer done()

	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(1, 1))

	if got := extractLogoURLFromParams(context.Background(), params); got != "" {
		t.Fatalf("logo url = %q, want empty", got)
	}
	if n := missWarnings(logs); n != 1 {
		t.Errorf("warned %d times, want 1", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the third reader did not record: %v", err)
	}
}

// The first two readers, end to end through the real action: a broken
// hero_deployed reaching BuildRenderContextAction files a row and leaves
// hero_url unset, rather than returning a clean-looking context.
func TestBuildRenderContextAction_BrokenHeroDeployedIsRecorded(t *testing.T) {
	params, logs, mock, done := auditTestParams(t, map[string]interface{}{
		"hero_deployed": the236Shape(),
		"input_data":    map[string]interface{}{"domain": "cookly.uk"},
	})
	defer done()
	mock.MatchExpectationsInOrder(false)
	mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(1, 1))

	out, err := BuildRenderContextAction(context.Background(), params)
	if err != nil {
		t.Fatalf("action failed: %v", err)
	}
	if n := missWarnings(logs); n != 1 {
		t.Errorf("warned %d times, want 1 (hero only — logo_deployed is absent, which is the no-op case)", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the real action did not record the miss: %v", err)
	}
	if result, ok := out.(map[string]interface{}); ok {
		if url, present := result["hero_url"]; present && url != "" {
			t.Errorf("hero_url = %v, want unset — the deploy result carried none", url)
		}
	}
}
