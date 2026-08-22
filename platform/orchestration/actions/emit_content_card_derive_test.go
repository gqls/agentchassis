// FILE: platform/orchestration/actions/emit_content_card_derive_test.go
//
// bugs_open/114. Two things are pinned here, and they are different in kind.
//
// 1. THE CONTRACT. The event-driven emitter and the discovery sweep both file
//    a needs_content_image item for the same page, and derive_card_asset reads
//    the spec's field names directly. If the two filings ever disagree about
//    the item key, idx_swi_dedup stops collapsing them and one page gets two
//    asset-deployer runs upserting the same card; if they disagree about the
//    spec, the handler reads a field that is not there. Both helpers are now
//    exported from discovery_checks precisely so there is one spelling — this
//    asserts the emitter actually uses it rather than reimplementing it.
//
// 2. THE BRANCHES. Every skip disposition is a decision not to spend an
//    asset-deployer run, and a skip that fires when it should not is silent by
//    construction — the page simply never gets a card, which is the whole of
//    bugs_open/114. So each arm is driven against sqlmock and asserted, and the
//    "raise" arm asserts the INSERT actually happens.

package actions

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"github.com/gqls/agentchassis/platform/orchestration/imageryplan"
	"go.uber.org/zap"
)

// TestContentCardDeriveUsesTheDiscoveryChecksOwnContract is the drift guard. It
// does not re-state what the key and spec look like — restating them here would
// create the third spelling. It asserts the emitter's inputs resolve through the
// SAME exported helpers the sweep calls, for a page name with the hyphen/underscore
// shape that makes key conventions go wrong.
func TestContentCardDeriveUsesTheDiscoveryChecksOwnContract(t *testing.T) {
	const pageName = "tool-rate-forecaster"
	const pageID = "3fbb0f5e-1111-2222-3333-444444444444"
	siteID, batchID := uuid.New(), uuid.New()

	// Build the item the emitter actually files. Asserting the helpers'
	// outputs alone does NOT pin this: an earlier cut of this test did exactly
	// that, and a mutation replacing the emitter's ContentImageItemKey call with
	// a hand-spelled "content-image:"+page passed it unnoticed.
	item, err := contentCardDeriveItem(siteID, pageID, pageName, batchID)
	if err != nil {
		t.Fatalf("item build failed: %v", err)
	}

	if want := discovery_checks.ContentImageItemKey(pageName); item.itemKey != want {
		t.Fatalf("item key = %q, want %q — the sweep and the event must produce the same key or idx_swi_dedup stops collapsing them into one asset-deployer run",
			item.itemKey, want)
	}
	if item.itemKey != "content_image:"+pageName {
		t.Fatalf("item key = %q — the contract is content_image:<page> with an UNDERSCORE; a hyphen here silently doubles the work", item.itemKey)
	}

	// derive_card_asset reads these field names off the spec. Asserting them on
	// the emitter's own item means a change to the shared helper's shape breaks
	// here rather than surfacing as a card that never appears.
	for _, needle := range []string{
		`"mode":"content_card"`,
		`"entity_type":"page"`,
		`"entity_id":"` + pageID + `"`,
		`"page_name":"` + pageName + `"`,
		`"purpose":"card"`,
	} {
		if !strings.Contains(item.spec, needle) {
			t.Fatalf("spec %s is missing %s — derive_card_asset reads this field", item.spec, needle)
		}
	}

	// Routing and dispatchability. 'detected' is what the SWEEP files, and it
	// waits for a promoter — the very dependency this emitter exists to remove,
	// since claim_work_item takes ('triaged','approved') only.
	if item.status != "triaged" {
		t.Fatalf("status = %q, want \"triaged\" — a 'detected' item filed at the event still waits for a promoter that may never run", item.status)
	}
	if item.handlerAgent != "asset-deployer" || item.itemType != "needs_content_image" {
		t.Fatalf("routing = %s/%s, want needs_content_image/asset-deployer", item.itemType, item.handlerAgent)
	}

	// The hero existence probe must use the resolver's own convention, or the
	// emitter would request a card for a hero plan_sections cannot find.
	if got, want := imageryplan.ContentHeroKey(pageName), "content_hero_tool_rate_forecaster"; got != want {
		t.Fatalf("ContentHeroKey(%q) = %q, want %q", pageName, got, want)
	}
}

// TestEmitContentCardDeriveBranches drives each disposition. The skips matter as
// much as the raise: a skip that fires wrongly produces no card and no error,
// which is indistinguishable from success until someone looks at a listing page.
func TestEmitContentCardDeriveBranches(t *testing.T) {
	const pageName = "tool-rate-forecaster"
	const pageID = "3fbb0f5e-1111-2222-3333-444444444444"
	siteID := uuid.New()
	batchID := uuid.New()

	cases := []struct {
		name        string
		hasPage     bool
		hasHero     bool
		hasCard     bool
		expectInsrt bool
		want        string
		why         string
	}{
		{
			name: "hero landed and no card yet", hasPage: true, hasHero: true, hasCard: false,
			expectInsrt: true, want: "raised",
			why: "this is the case the whole change exists for — the sweep that used to do this has not run since 2026-08-11",
		},
		{
			name: "card already exists", hasPage: true, hasHero: true, hasCard: true,
			expectInsrt: false, want: "skipped_card_exists",
			why: "the listing joins already resolve; re-deriving costs an asset-deployer run for no pixel change",
		},
		{
			name: "no content hero for this page", hasPage: true, hasHero: false, hasCard: false,
			expectInsrt: false, want: "skipped_no_content_hero",
			why: "a logo/favicon/sprite landing on a page-scoped item has no hero to derive a card from",
		},
		{
			name: "no page row of that name", hasPage: false,
			expectInsrt: false, want: "skipped_no_page",
			why: "nothing to link a card to; must not fabricate an entity_id",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()

			mock.ExpectBegin()
			q := mock.ExpectQuery("FROM pages").
				WithArgs(siteID, pageName, imageryplan.ContentHeroKey(pageName))
			if tc.hasPage {
				q.WillReturnRows(sqlmock.NewRows([]string{"id", "has_hero", "has_card"}).
					AddRow(pageID, tc.hasHero, tc.hasCard))
			} else {
				q.WillReturnError(sql.ErrNoRows)
			}
			if tc.expectInsrt {
				mock.ExpectExec("INSERT INTO site_work_items").
					WillReturnResult(sqlmock.NewResult(1, 1))
			}

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("begin: %v", err)
			}
			got := emitContentCardDerive(context.Background(), tx, siteID, pageName, batchID, zap.NewNop())
			_ = tx.Rollback()

			if got != tc.want {
				t.Fatalf("disposition = %q, want %q — %s", got, tc.want, tc.why)
			}
			// The negative half: with no INSERT expectation registered, an
			// unexpected insert leaves ExpectationsWereMet unsatisfied or errors
			// the exec, so this catches a skip arm that files anyway.
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet/unexpected DB expectations (%s): %v", tc.why, err)
			}
		})
	}
}
