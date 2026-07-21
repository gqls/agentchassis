package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// Covers the bugs_open/040-partial-build deploy guard: pageSectionShortfall
// reports (planned, rendered) so UpdatePageStatusAction can refuse to stamp a
// page deployed when it is short of its plan. The guard's decision is
// rendered < planned.
//
// The row-count comparison is deliberate: pages.sections names do NOT reliably
// equal page_components.slot_name / content_components.function across templates
// (gaswholesalers services planned ["services-hero","call_to_action"] against
// live slots ["hero-services","call-to-action"]), so per-name matching would
// refuse healthy pages. These fixtures are shaped from the real live rows read
// 2026-07-21: dartsonline index = 6 planned / 5 rendered (refuse); gaswholesalers
// services = 4 planned / 4 rendered (allow, despite every name differing).

// shortfallRefuses mirrors the guard's decision in UpdatePageStatusAction.
func shortfallRefuses(planned, rendered int) bool { return rendered < planned }

func TestPageSectionShortfall(t *testing.T) {
	// query substring stable across the SELECT — avoids matching the
	// parenthesised sub-selects as regex metacharacters.
	const querySig = `FROM pages p\s+WHERE p.id =`

	cases := []struct {
		name       string
		planned    int
		rendered   int
		wantRefuse bool
	}{
		{
			// dartsonline index, 2026-07-20: testimonials never written.
			name:       "partial build — 5 of 6 sections, must refuse",
			planned:    6,
			rendered:   5,
			wantRefuse: true,
		},
		{
			// gaswholesalers services: complete despite every section name
			// diverging from its component slot_name. Row count catches this
			// where name matching would false-positive.
			name:       "complete build with divergent names — must allow",
			planned:    4,
			rendered:   4,
			wantRefuse: false,
		},
		{
			// A page rendered more rows than planned (duplicate/chrome rows):
			// never a shortfall.
			name:       "more rows than planned — must allow",
			planned:    3,
			rendered:   5,
			wantRefuse: false,
		},
		{
			// sections=[] (tool/blog-index page rendered by another subsystem,
			// bugs_open/050) — planned 0, never a shortfall regardless of rows.
			name:       "no planned sections — must allow",
			planned:    0,
			rendered:   0,
			wantRefuse: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()

			pageID := uuid.New()
			mock.ExpectQuery(querySig).
				WithArgs(pageID).
				WillReturnRows(sqlmock.NewRows([]string{"planned", "rendered"}).
					AddRow(tc.planned, tc.rendered))

			planned, rendered, err := pageSectionShortfall(context.Background(), db, pageID)
			if err != nil {
				t.Fatalf("pageSectionShortfall: %v", err)
			}
			if planned != tc.planned || rendered != tc.rendered {
				t.Fatalf("counts: got planned=%d rendered=%d, want planned=%d rendered=%d",
					planned, rendered, tc.planned, tc.rendered)
			}
			if got := shortfallRefuses(planned, rendered); got != tc.wantRefuse {
				t.Errorf("refuse decision: got %v, want %v (planned=%d rendered=%d)",
					got, tc.wantRefuse, planned, rendered)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("expectations: %v", err)
			}
		})
	}
}
