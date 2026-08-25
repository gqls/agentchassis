package actions

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// argJSONSpec captures the spec JSON an INSERT was called with, so a test can
// assert on its CONTENT rather than only on the presence of keys — the sibling
// of argJSONHasKeys, which can only say a key is there.
type argJSONSpec struct{ got *map[string]interface{} }

func (a argJSONSpec) Match(v driver.Value) bool {
	s, ok := v.(string)
	if !ok {
		b, okb := v.([]byte)
		if !okb {
			return false
		}
		s = string(b)
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return false
	}
	*a.got = m
	return true
}

// argCapture records a string argument (the item_key) so a test can compare the
// key across two runs rather than restate it.
type argCapture struct{ got *string }

func (a argCapture) Match(v driver.Value) bool {
	switch t := v.(type) {
	case string:
		*a.got = t
	case []byte:
		*a.got = string(t)
	default:
		return false
	}
	return true
}

// fileOneContrastFinding runs the action over a single firm contrast finding and
// returns the spec it filed. summaryExtra is merged into the payload's summary,
// which is how a test chooses whether the adapter "speaks" cascade/v1.
func fileOneContrastFinding(t *testing.T, contrast map[string]interface{}, summaryExtra map[string]interface{}) (map[string]interface{}, string) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()
	var spec map[string]interface{}
	var key string

	mock.ExpectQuery("locked_at IS NOT NULL").
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}))
	mock.ExpectQuery("FROM pages").WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}).AddRow(pageID.String(), "/index.html"))
	mock.ExpectBegin()
	expectWorkItemDoorStandsDown(mock)
	mock.ExpectQuery("INTERVAL '7 days'").
		WithArgs(siteID, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(siteID, "render-audit", "build", "contrast_failure", sqlmock.AnyArg(),
			sqlmock.AnyArg(), argJSONSpec{got: &spec}, pageID, sqlmock.AnyArg(),
			60, "css-patch-agent", "detected", sqlmock.AnyArg(),
			argCapture{got: &key}, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	payload := map[string]interface{}{
		"run_id":   "run-cascade",
		"contrast": []map[string]interface{}{contrast},
	}
	if len(summaryExtra) > 0 {
		payload["summary"] = summaryExtra
	}
	if _, err := WriteRenderAuditFindingsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData:    renderAuditCollected(siteID, payload),
	}); err != nil {
		t.Fatalf("action failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
	return spec, key
}

// theVoncFinding is the worked case from bugs_open/390 in payload form: a real
// verified selector whose winning declaration is a section-scoped rule in the
// page's own <style> block, with the theme linked before it.
func theVoncFinding(withWinner bool) map[string]interface{} {
	c := map[string]interface{}{
		"url": "https://vonc.com/index.html", "tag": "a",
		"class": "gauntlet-btn-primary", "selector": "A.gauntlet-btn-primary",
		"matches": 2, "selector_verified": true, "text": "Enter the gauntlet",
		"fg": "rgb(124, 60, 255)", "bg": "rgb(252,92,125)",
		"ratio": 1.76, "need": 4.5, "font_px": 16, "over_image": false,
	}
	if withWinner {
		c["fg_winner"] = map[string]interface{}{
			"property": "color", "surface": "style_block",
			"selector": ".gauntlet-cta-section .gauntlet-btn-primary",
			"decl":     "var(--color-primary, #1a1a2e)", "var_name": "--color-primary",
			"verified": true, "theme_after_winner": false, "candidates": 2,
		}
	}
	return c
}

// TestCascadeSkewGuardKeepsTheOldSpecByteIdentical is the version-skew contract.
//
// The adapter and this binary are separate images and roll independently. An
// adapter too old to attribute sends no `cascade_scheme`, and the spec it
// produces must then be EXACTLY what it was before this change existed —
// otherwise every consumer of a contrast spec has to learn a new shape on a
// timetable nobody controls.
//
// MUTATION: make the routing block unconditional (drop the
// `payload.Summary.CascadeScheme != ""` gate) and this test goes RED.
func TestCascadeSkewGuardKeepsTheOldSpecByteIdentical(t *testing.T) {
	// An old adapter: it still sends the finding, and even a winner would be
	// meaningless without the capability declaration.
	spec, _ := fileOneContrastFinding(t, theVoncFinding(false), map[string]interface{}{
		"selector_scheme": "verified/v1",
	})
	for _, k := range []string{"repair_surface", "cascade_scheme", "winning_rule", "override_requirement", "override_example"} {
		if _, present := spec[k]; present {
			t.Errorf("spec carries %q against an adapter that does not speak cascade/v1 — "+
				"an old adapter must produce the spec this platform produced before "+
				"attribution existed, not a half-populated new one", k)
		}
	}
	if spec["selector_scheme"] != "verified/v1" {
		t.Errorf("the pre-existing selector_scheme key was disturbed: %v", spec["selector_scheme"])
	}
}

// TestCascadeRoutingIsWrittenIntoTheSpec is the positive half: a modern adapter's
// attribution reaches the spec, in the shape the repair prompt renders.
func TestCascadeRoutingIsWrittenIntoTheSpec(t *testing.T) {
	spec, _ := fileOneContrastFinding(t, theVoncFinding(true), map[string]interface{}{
		"selector_scheme": "verified/v1",
		"cascade_scheme":  "cascade/v1",
	})
	if spec["repair_surface"] != repairSurfaceTheme {
		t.Fatalf("repair_surface = %v, want %q", spec["repair_surface"], repairSurfaceTheme)
	}
	req, ok := spec["override_requirement"].(map[string]interface{})
	if !ok {
		t.Fatalf("override_requirement missing or not an object: %#v", spec["override_requirement"])
	}
	if req["min_specificity_text"] != "0,2,0" {
		t.Errorf("min_specificity_text = %v, want 0,2,0 (the page rule that actually wins)",
			req["min_specificity_text"])
	}
	if req["strictly_greater"] != true {
		t.Error("strictly_greater is not set, but the theme is linked BEFORE the winner, " +
			"so an equal-specificity rule loses on source order — this is the bug")
	}
	// The example must be one the platform CHECKED, not one it hoped was right.
	ex, _ := spec["override_example"].(string)
	if ex == "" {
		t.Fatal("no override_example offered")
	}
	if !satisfiesRequirement(ex, &overrideRequirement{
		MinSpecificity:  [3]int{0, 2, 0},
		StrictlyGreater: true,
	}) {
		t.Errorf("override_example %q does not satisfy the requirement it is offered for — "+
			"a worked example that does not work is worse than none", ex)
	}
	if _, ok := spec["winning_rule"].(map[string]interface{}); !ok {
		t.Error("winning_rule absent; a parked row must name what beat it")
	}
}

// TestCascadeFieldsDoNotChangeTheItemKey. The key embeds the selector and is the
// join retraction uses (VIZ-016). Changing it silently re-keys every open row and
// has already forced two alias-key retrofits — so the new facts ride the spec
// and the key must be byte-identical with and without them.
//
// MUTATION: key the item on the winner's selector instead of the filed one and
// this test goes RED.
func TestCascadeFieldsDoNotChangeTheItemKey(t *testing.T) {
	_, without := fileOneContrastFinding(t, theVoncFinding(false), map[string]interface{}{
		"selector_scheme": "verified/v1",
	})
	_, with := fileOneContrastFinding(t, theVoncFinding(true), map[string]interface{}{
		"selector_scheme": "verified/v1",
		"cascade_scheme":  "cascade/v1",
	})
	if without != with {
		t.Fatalf("item_key changed when attribution was added:\n without = %q\n with    = %q\n"+
			"retraction joins on this key; changing it strands every open row", without, with)
	}
	if with != "contrast_failure:/index.html#A.gauntlet-btn-primary" {
		t.Errorf("item_key = %q, want the filed-selector form", with)
	}
}

// TestUnverifiedAttributionReachesTheSpecAsUnattributed. The probe reporting a
// guess it could not prove must land as "we do not know", never as a routable
// answer — and it must still say so out loud, so a blinded run is visible.
func TestUnverifiedAttributionReachesTheSpecAsUnattributed(t *testing.T) {
	c := theVoncFinding(true)
	w := c["fg_winner"].(map[string]interface{})
	w["verified"] = false
	spec, _ := fileOneContrastFinding(t, c, map[string]interface{}{
		"selector_scheme": "verified/v1",
		"cascade_scheme":  "cascade/v1",
	})
	if spec["repair_surface"] != repairSurfaceUnattributed {
		t.Errorf("repair_surface = %v, want %q for an unproven attribution",
			spec["repair_surface"], repairSurfaceUnattributed)
	}
	if _, present := spec["override_requirement"]; present {
		t.Error("an unproven attribution produced a requirement; a guess must not be routed on")
	}
	if _, present := spec["winning_rule"]; present {
		t.Error("an unproven winner was recorded as the winning rule")
	}
}
