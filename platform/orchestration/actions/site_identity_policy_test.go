package actions

import (
	"context"
	"errors"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Tests for the site identity policy (bugs_open/215, quiet mode). The helper is
// the ONE reader of these two keys; the contract test at the foot pins that both
// canonicalisation surfaces actually consume the answer.

var siteIdentityPolicyQuery = regexp.QuoteMeta(`
		SELECT COALESCE((data->>'honour_realised_identity')::boolean, false),
		       COALESCE((data->>'stem_twin_snap')::boolean, false)
		FROM site_specs
		WHERE site_id = $1 AND aspect = 'structure' AND is_current = true
	`)

func TestSiteIdentityPolicyFor(t *testing.T) {
	siteID := uuid.New()

	cases := []struct {
		name       string
		rows       func(sqlmock.Sqlmock)
		wantHonour bool
		wantStem   bool
	}{
		{
			name: "both keys true",
			rows: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(siteIdentityPolicyQuery).WithArgs(siteID).
					WillReturnRows(sqlmock.NewRows([]string{"honour", "stem"}).AddRow(true, true))
			},
			wantHonour: true, wantStem: true,
		},
		{
			// The pilot shape: preserve realised identities without yet trusting
			// the weakest matching layer. These must be independently settable or
			// the safe half cannot ship ahead of the risky one.
			name: "honour without stem",
			rows: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(siteIdentityPolicyQuery).WithArgs(siteID).
					WillReturnRows(sqlmock.NewRows([]string{"honour", "stem"}).AddRow(true, false))
			},
			wantHonour: true, wantStem: false,
		},
		{
			name: "spec present, keys absent, defaults false",
			rows: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(siteIdentityPolicyQuery).WithArgs(siteID).
					WillReturnRows(sqlmock.NewRows([]string{"honour", "stem"}).AddRow(false, false))
			},
		},
		{
			name: "no current structure spec: today's behaviour",
			rows: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(siteIdentityPolicyQuery).WithArgs(siteID).
					WillReturnRows(sqlmock.NewRows([]string{"honour", "stem"}))
			},
		},
		{
			// A malformed value lands here too (::boolean rejects it). A plan
			// write must never be lost over a spec typo.
			name: "read error: today's behaviour, not a failure",
			rows: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(siteIdentityPolicyQuery).WithArgs(siteID).
					WillReturnError(errors.New("invalid input syntax for type boolean"))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()
			tc.rows(mock)

			got := siteIdentityPolicyFor(context.Background(), db, siteID, zap.NewNop())
			if got.HonourRealisedIdentity != tc.wantHonour {
				t.Errorf("HonourRealisedIdentity = %v, want %v", got.HonourRealisedIdentity, tc.wantHonour)
			}
			if got.StemTwinSnap != tc.wantStem {
				t.Errorf("StemTwinSnap = %v, want %v", got.StemTwinSnap, tc.wantStem)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

func TestSiteIdentityPolicyNilDB(t *testing.T) {
	got := siteIdentityPolicyFor(context.Background(), nil, uuid.New(), zap.NewNop())
	if got.HonourRealisedIdentity || got.StemTwinSnap {
		t.Errorf("nil db must mean today's behaviour, got %+v", got)
	}
}

func TestRealisedIdentityOf(t *testing.T) {
	cases := []struct {
		name     string
		page     map[string]interface{}
		wantOK   bool
		wantName string
	}{
		{
			name: "marked page yields its stored identity",
			page: map[string]interface{}{
				"identity_authority": "realised",
				"name":               "llm-cost-calculator",
				"url":                "/tools/llm-cost-calculator.html",
				"page_type":          "tool",
			},
			wantOK: true, wantName: "llm-cost-calculator",
		},
		{
			// The ordinary case: an LLM-proposed page has no authority over its
			// own identity and must be canonicalised as it always was.
			name:   "unmarked page falls through",
			page:   map[string]interface{}{"name": "pricing", "url": "/pricing.html", "page_type": "content"},
			wantOK: false,
		},
		{
			// Fails safe: a marked page missing any identity field must not write
			// a blank name or url over a real one.
			name: "marked but incomplete falls through",
			page: map[string]interface{}{
				"identity_authority": "realised",
				"name":               "llm-cost-calculator",
				"url":                "",
				"page_type":          "tool",
			},
			wantOK: false,
		},
		{
			// Only this exact value carries authority — a stray or forged marker
			// value is not a licence.
			name: "unrecognised authority value falls through",
			page: map[string]interface{}{
				"identity_authority": "llm",
				"name":               "x", "url": "/x.html", "page_type": "content",
			},
			wantOK: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			name, _, _, ok := realisedIdentityOf(tc.page)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && name != tc.wantName {
				t.Errorf("name = %q, want %q", name, tc.wantName)
			}
		})
	}
}

// TestIdentityPolicyReachesBothCanonicalisationSurfaces pins the invariant the
// helper's comment states, and is the load-bearing test of this change: a
// reconciler that correctly recognises a twin is USELESS if either write surface
// re-derives the identity afterwards, and the two surfaces disagreeing is worse
// than the bug — it breaks the site_plan_pages.name = pages.name join that
// check_sectionless_pages and reconcile_site_plan both depend on.
//
// Mutation-checked: deleting the guard from either file fails this test while
// every unit test above still passes, which is exactly why it exists.
//
// The guard is matched ANCHORED at its call, not grepped bare, so a comment
// mentioning the helper cannot satisfy it (the source-scan trap: the first
// occurrence wins, and comments are the usual first occurrence).
func TestIdentityPolicyReachesBothCanonicalisationSurfaces(t *testing.T) {
	guardCall := regexp.MustCompile(`realisedIdentityOf\(\w+\)`)
	policyRead := regexp.MustCompile(`siteIdentityPolicyFor\(`)
	gated := regexp.MustCompile(`identityPolicy\.HonourRealisedIdentity`)

	for _, file := range []string{"write_site_plan_action.go", "site_db_actions.go"} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("cannot read %s: %v", file, err)
		}
		if !policyRead.Match(src) {
			t.Errorf("%s no longer calls siteIdentityPolicyFor — who owns a page identity must come from the one shared reader, or the two surfaces can disagree", file)
		}
		if !guardCall.Match(src) {
			t.Errorf("%s no longer calls realisedIdentityOf — this surface will re-derive a realised page's identity and re-mint the twin bugs_open/215 exists to prevent", file)
		}
		if !gated.Match(src) {
			t.Errorf("%s applies the identity guard without gating it on the site policy — the unsafe default must stay off (owner ruling 2026-08-02)", file)
		}
	}
}
