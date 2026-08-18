package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// bugs_open/300 — page_component_status_drift keys its finding on
// page_components.id, which the estate's own rule (016b, and two live code
// comments) says is not stable across re-renders. A re-render between filing and
// dispatch turned a TRUE finding into sql.ErrNoRows, which this action turned
// into a hard error and the item into `failed`.
//
// WHY THAT COSTS MORE THAN ONE LOST REPAIR. detected-item-promoter's 25% floor
// (migration 444/454) cannot tell an artefact failure from an incompetent
// handler, so enough of them switch the whole item_type off — including the
// findings that are still true. The type disables itself by ageing.
//
// [MEASURED 2026-08-18, all 82 lifetime rows] spec.page_component_id resolves for
// 70; (page_id, slot_name) resolves for 82 of 82. And the ageing is observable
// rather than theoretical: 16 of 16 deferred rows resolved by id on 2026-08-17,
// 11 do today.
//
// THE AMBIGUITY CASES ARE NOT HYPOTHETICAL EITHER. (page_id, slot_name) is NOT
// unique fleet-wide — [MEASURED 2026-08-18] 17 pairs carry more than one
// component, worst case 4. Zero are drift rows today, which is exactly why a
// test has to hold the line: resolving by the pair alone would be correct on
// every row that exists now and silently wrong on the first one that does not.

const stableKeyQueryRe = `JOIN site_work_items wi ON wi\.page_id = pc\.page_id`

// resolverHarness drives resolveStatusRepairComponent against sqlmock.
type resolverHarness struct {
	t    *testing.T
	mock sqlmock.Sqlmock
	p    ActionParams
}

func newResolverHarness(t *testing.T) *resolverHarness {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &resolverHarness{
		t:    t,
		mock: mock,
		p: ActionParams{
			Context:          context.Background(),
			DB:               db,
			Logger:           zap.NewNop(),
			ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		},
	}
}

// expectStableKey queues the (page_id, slot_name) lookup, returning the given ids.
func (h *resolverHarness) expectStableKey(ids ...uuid.UUID) {
	rows := sqlmock.NewRows([]string{"id"})
	for _, id := range ids {
		rows.AddRow(id)
	}
	h.mock.ExpectQuery(stableKeyQueryRe).WillReturnRows(rows)
}

func TestResolveStatusRepairComponent_StaleStoredIdLosesToTheStableKey(t *testing.T) {
	// The measured instance from bugs_open/300: item bc041cfb… named component
	// 0f02ca76… (0 rows in page_components) while slot prose-0 on the same page
	// held a9550607…, deployed, after a re-render five days later.
	dead := uuid.New()
	live := uuid.New()

	h := newResolverHarness(t)
	h.expectStableKey(live)

	got, by, reason, ok := resolveStatusRepairComponent(
		context.Background(), h.p, dead.String(), "prose-0", uuid.New().String(), zap.NewNop())

	if !ok {
		t.Fatalf("resolution failed with %q — under the old lookup this row was a hard error and a vote against the pair", reason)
	}
	if got != live {
		t.Errorf("resolved to %s, want the live component %s", got, live)
	}
	if by != "page_id+slot_name" {
		t.Errorf("resolved_by = %q, want page_id+slot_name — the census needs to tell the two keys apart", by)
	}
}

func TestResolveStatusRepairComponent_AmbiguousPairIsBrokenByTheStoredId(t *testing.T) {
	// 17 (page_id, slot_name) pairs on the estate carry more than one component.
	// When the stored id is one of the matches it is the finding's own subject
	// and must win — this is the "id as tiebreak" half of the owner's decision.
	other := uuid.New()
	mine := uuid.New()

	h := newResolverHarness(t)
	h.expectStableKey(other, mine)

	got, by, _, ok := resolveStatusRepairComponent(
		context.Background(), h.p, mine.String(), "prose-0", uuid.New().String(), zap.NewNop())

	if !ok {
		t.Fatal("an ambiguous pair whose stored id is among the matches is resolvable, not a refusal")
	}
	if got != mine {
		t.Errorf("resolved to %s, want the stored subject %s — picking either match is how a fix for one bug files another", got, mine)
	}
	if !strings.Contains(by, "tiebreak") {
		t.Errorf("resolved_by = %q, want it to record that the tiebreak fired", by)
	}
}

func TestResolveStatusRepairComponent_AmbiguousPairWithNoTiebreakRefusesToGuess(t *testing.T) {
	// Two matches, stored id is neither. There is no honest answer, so the
	// posture is the same as the action's two existing guards: refuse rather
	// than guess. It must NOT return one of them, and it must NOT be a hard
	// error either — an unguessable subject is not a handler that failed.
	h := newResolverHarness(t)
	h.expectStableKey(uuid.New(), uuid.New())

	_, _, reason, ok := resolveStatusRepairComponent(
		context.Background(), h.p, uuid.New().String(), "prose-0", uuid.New().String(), zap.NewNop())

	if ok {
		t.Fatal("guessed a component from an ambiguous slot — this is the failure mode the tiebreak exists to prevent")
	}
	if !strings.Contains(reason, "refusing to guess") {
		t.Errorf("reason %q does not say why it refused; the row is the only place a human will read it", reason)
	}
}

func TestResolveStatusRepairComponent_FallsBackToTheStoredId(t *testing.T) {
	// Backwards compatibility, and the negative control for the whole change: a
	// caller with no work_item_id or no slot_name must behave EXACTLY as before.
	// Without this, "the stable key is used" is equally consistent with having
	// broken every other caller of this fix_type.
	stored := uuid.New()

	cases := []struct {
		name       string
		slot, item string
	}{
		{"no slot_name in the spec", "", uuid.New().String()},
		{"no work_item_id in the payload", "prose-0", ""},
		{"neither", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newResolverHarness(t) // no query queued: none must be issued
			got, by, _, ok := resolveStatusRepairComponent(
				context.Background(), h.p, stored.String(), tc.slot, tc.item, zap.NewNop())
			if !ok || got != stored {
				t.Errorf("got (%s, ok=%v), want the stored id %s", got, ok, stored)
			}
			if by != "spec.page_component_id" {
				t.Errorf("resolved_by = %q, want spec.page_component_id", by)
			}
			if err := h.mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("expectations: %v", err)
			}
		})
	}
}

func TestResolveStatusRepairComponent_StableKeyMissFallsBackRatherThanRefusing(t *testing.T) {
	// Zero matches on the pair is not a refusal — the stored id may still be
	// good (this is every row filed before the page was ever re-rendered).
	stored := uuid.New()
	h := newResolverHarness(t)
	h.expectStableKey() // no rows

	got, by, _, ok := resolveStatusRepairComponent(
		context.Background(), h.p, stored.String(), "prose-0", uuid.New().String(), zap.NewNop())

	if !ok || got != stored {
		t.Errorf("got (%s, ok=%v), want the stored id %s", got, ok, stored)
	}
	if by != "spec.page_component_id" {
		t.Errorf("resolved_by = %q, want spec.page_component_id", by)
	}
}

func TestResolveStatusRepairComponent_NoSubjectAtAll(t *testing.T) {
	h := newResolverHarness(t)
	h.expectStableKey()

	_, _, reason, ok := resolveStatusRepairComponent(
		context.Background(), h.p, "", "prose-0", uuid.New().String(), zap.NewNop())

	if ok {
		t.Fatal("resolved a subject from nothing")
	}
	if !strings.Contains(reason, "no usable subject") {
		t.Errorf("reason %q should say the subject could not be identified", reason)
	}
}

// End to end through the action, on the shape that motivated the bug: a stale
// stored id, a live component in the slot, already deployed because a re-render
// repaired it. The old code produced a hard error here.
func TestFixPageComponentStatus_StaleIdResolvesAndReportsAlreadyDeployed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	dead := uuid.New()
	live := uuid.New()
	item := uuid.New()

	mock.ExpectQuery(stableKeyQueryRe).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(live))
	mock.ExpectQuery(`FROM page_components pc`).
		WithArgs(live).
		WillReturnRows(sqlmock.NewRows([]string{"build_status", "slot_name", "page_build_status", "has_html"}).
			AddRow("deployed", "prose-0", "deployed", true))

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{
				"work_item_id": item.String(),
				"spec": map[string]interface{}{
					"page_component_id": dead.String(),
					"slot_name":         "prose-0",
				},
			},
		},
	}

	out, err := fixPageComponentStatus(context.Background(), params, zap.NewNop())
	if err != nil {
		t.Fatalf("hard error on a resolvable finding: %v — this is exactly bugs_open/300", err)
	}
	res, _ := out.(map[string]interface{})
	if res["reason"] != "already deployed" {
		t.Errorf("reason = %v, want %q", res["reason"], "already deployed")
	}
	if res["page_component_id"] != live.String() {
		t.Errorf("page_component_id = %v, want the LIVE component %s — the result must name what was actually inspected", res["page_component_id"], live)
	}
	if res["resolved_by"] != "page_id+slot_name" {
		t.Errorf("resolved_by = %v, want page_id+slot_name", res["resolved_by"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
