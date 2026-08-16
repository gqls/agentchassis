package discovery_checks

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestOrderedListsEqual(t *testing.T) {
	cases := []struct {
		name string
		a, b []string
		want bool
	}{
		{"identical", []string{"hero", "cta"}, []string{"hero", "cta"}, true},
		{"different length", []string{"hero"}, []string{"hero", "cta"}, false},
		{"different order (layout matters)", []string{"hero", "cta"}, []string{"cta", "hero"}, false},
		{"different content", []string{"hero", "specs"}, []string{"hero", "cta"}, false},
		{"both empty", []string{}, []string{}, true},
		// The exact robot-hands product-detail drift that motivated this check:
		// authoritative table still held the old e-commerce layout while
		// pages.sections had been swapped to the spec sheet.
		{
			"product-detail drift",
			[]string{"product-hero", "product-specs", "call-to-action"},
			[]string{"gripper-spec-sheet", "call-to-action"},
			false,
		},
	}
	for _, tc := range cases {
		if got := orderedListsEqual(tc.a, tc.b); got != tc.want {
			t.Errorf("%s: orderedListsEqual(%v,%v)=%v want %v", tc.name, tc.a, tc.b, got, tc.want)
		}
	}
}

// bugs_open/285 — the loader now merges human-locked live rows into the list
// and into pages.sections, so a page whose plan omits a locked section is NOT
// drifted, whether its cache pre-dates the fix (raw plan) or post-dates it
// (plan + locked). A genuine one-store edit still is.
func TestSectionSourceDrift_LockedLiveRowIsNotDrift(t *testing.T) {
	siteID := uuid.New()
	run := func(t *testing.T, cache map[string]string, wantFindings int) {
		t.Helper()
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock: %v", err)
		}
		defer db.Close()
		mock.ExpectQuery("FROM site_plan_sections sps").WithArgs(siteID).
			WillReturnRows(sqlmock.NewRows([]string{"page_name", "component_name"}).
				AddRow("contact", "hero").AddRow("contact", "contact-info").
				AddRow("about", "hero-about").AddRow("about", "faq"))
		mock.ExpectQuery("SELECT data FROM site_specs").WithArgs(siteID).WillReturnError(sql.ErrNoRows)
		cacheRows := sqlmock.NewRows([]string{"name", "sections"})
		for name, secs := range cache {
			cacheRows.AddRow(name, secs)
		}
		mock.ExpectQuery("FROM pages").WithArgs(siteID).WillReturnRows(cacheRows)
		// contact carries a locked chat box at position 3 the plan does not name.
		mock.ExpectQuery("FROM page_components pc").WithArgs(siteID, "").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slot_name", "position", "component_id", "function", "cname", "lock_type", "locked_by"}).
				AddRow("r1", "contact", "chat-input-box", 3, "cid", "chat-input-box", "chat-input-box", "permanent", "lane"))

		res, err := (&SectionSourceDriftCheck{}).Run(DiscoveryCheckContext{
			Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
		})
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if got := len(res.Findings); got != wantFindings {
			t.Errorf("findings = %d (%v), want %d", got, res.Findings, wantFindings)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectations: %v", err)
		}
	}
	t.Run("cache pre-dates the fix (raw plan) → not drift", func(t *testing.T) {
		run(t, map[string]string{"contact": `["hero","contact-info"]`, "about": `["hero-about","faq"]`}, 0)
	})
	t.Run("cache written by the fixed loader (plan + locked) → not drift", func(t *testing.T) {
		run(t, map[string]string{"contact": `["hero","contact-info","chat-input-box"]`, "about": `["hero-about","faq"]`}, 0)
	})
	t.Run("a genuine one-store edit still flags", func(t *testing.T) {
		run(t, map[string]string{"contact": `["hero","contact-info","chat-input-box"]`, "about": `["gripper-spec-sheet","faq"]`}, 1)
	})
}
