// Tests for classifyThunderReconcile — the pure comparison at the heart of
// the orphan sweep. The table exercises every classification arm and, more
// importantly, the NEGATIVE space: the shapes that must NOT produce a
// finding (matched pairs, the provision-window grace, vendor DELETED
// entries), because a false orphan is what would train operators to
// ignore the sweep.

package actions

import (
	"strings"
	"testing"
	"time"
)

func TestClassifyThunderReconcile(t *testing.T) {
	now := time.Date(2026, 8, 8, 18, 0, 0, 0, time.UTC)
	grace := 30 * time.Minute
	old := now.Add(-2 * time.Hour)     // well past grace
	young := now.Add(-5 * time.Minute) // inside grace

	cases := []struct {
		name        string
		vendor      []thunderVendorInstance
		rows        []thunderDBRow
		wantKinds   []string // finding kinds, in order (orphans then ghosts)
		wantMatched int
		wantInGrace int
		wantBilling int
		wantDeleted int
	}{
		{
			name:        "old vendor instance with no row is an orphan",
			vendor:      []thunderVendorInstance{{ID: "3", Status: "RUNNING", GpuType: "A6000", CreatedAt: old}},
			wantKinds:   []string{"orphan_no_row"},
			wantBilling: 1,
		},
		{
			name:        "young vendor instance with no row is the provision window, not an orphan",
			vendor:      []thunderVendorInstance{{ID: "3", Status: "PENDING", CreatedAt: young}},
			wantInGrace: 1,
			wantBilling: 1,
		},
		{
			name:        "unknown age cannot hide behind the grace window",
			vendor:      []thunderVendorInstance{{ID: "3", Status: "RUNNING"}}, // zero CreatedAt
			wantKinds:   []string{"orphan_no_row"},
			wantBilling: 1,
		},
		{
			name:   "vendor instance whose only row is terminal is an orphan",
			vendor: []thunderVendorInstance{{ID: "1", Status: "RUNNING", CreatedAt: old}},
			rows: []thunderDBRow{
				{RowID: "r1", ThunderInstanceID: "1", Status: "decommissioned"},
			},
			wantKinds:   []string{"orphan_terminal_row"},
			wantBilling: 1,
		},
		{
			name:   "vendor instance with a live row is matched",
			vendor: []thunderVendorInstance{{ID: "1", Status: "RUNNING", CreatedAt: old}},
			rows: []thunderDBRow{
				{RowID: "r1", ThunderInstanceID: "1", Status: "running"},
			},
			wantMatched: 1,
			wantBilling: 1,
		},
		{
			name:   "a live row alongside a terminal row still matches — history does not orphan",
			vendor: []thunderVendorInstance{{ID: "1", Status: "RUNNING", CreatedAt: old}},
			rows: []thunderDBRow{
				{RowID: "r1", ThunderInstanceID: "1", Status: "decommissioned"},
				{RowID: "r2", ThunderInstanceID: "1", Status: "running"},
			},
			wantMatched: 1,
			wantBilling: 1,
		},
		{
			name:        "vendor DELETED instance is not billing and produces nothing",
			vendor:      []thunderVendorInstance{{ID: "9", Status: "DELETED", CreatedAt: old}},
			wantDeleted: 1,
		},
		{
			name:      "live row with no vendor instance is a ghost",
			rows:      []thunderDBRow{{RowID: "r1", ThunderInstanceID: "7", Status: "running"}},
			wantKinds: []string{"ghost_row"},
		},
		{
			name:        "live row whose vendor instance is DELETED is a ghost",
			vendor:      []thunderVendorInstance{{ID: "7", Status: "DELETED", CreatedAt: old}},
			rows:        []thunderDBRow{{RowID: "r1", ThunderInstanceID: "7", Status: "decommissioning"}},
			wantKinds:   []string{"ghost_row"},
			wantDeleted: 1,
		},
		{
			name: "terminal rows alone are silent — history is not a ghost",
			rows: []thunderDBRow{
				{RowID: "r1", ThunderInstanceID: "5", Status: "decommissioned"},
				{RowID: "r2", ThunderInstanceID: "6", Status: "failed"},
			},
		},
		{
			name:        "vendor status casing does not matter",
			vendor:      []thunderVendorInstance{{ID: "9", Status: "deleted", CreatedAt: old}},
			wantDeleted: 1,
		},
		{
			name: "mixed account: one orphan, one match, one ghost, one in grace",
			vendor: []thunderVendorInstance{
				{ID: "0", Status: "RUNNING", CreatedAt: old},                  // matched
				{ID: "1", Status: "RUNNING", GpuType: "H100", CreatedAt: old}, // orphan
				{ID: "2", Status: "PENDING", CreatedAt: young},                // in grace
			},
			rows: []thunderDBRow{
				{RowID: "r0", ThunderInstanceID: "0", Status: "running"},
				{RowID: "r9", ThunderInstanceID: "42", Status: "running"}, // ghost
			},
			wantKinds:   []string{"orphan_no_row", "ghost_row"},
			wantMatched: 1,
			wantInGrace: 1,
			wantBilling: 3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyThunderReconcile(tc.vendor, tc.rows, grace, now)

			var kinds []string
			for _, f := range got.Findings {
				kinds = append(kinds, f.Kind)
			}
			if len(kinds) != len(tc.wantKinds) {
				t.Fatalf("findings = %v, want kinds %v", got.Findings, tc.wantKinds)
			}
			for i := range kinds {
				if kinds[i] != tc.wantKinds[i] {
					t.Errorf("finding[%d].Kind = %q, want %q", i, kinds[i], tc.wantKinds[i])
				}
			}
			if got.Matched != tc.wantMatched {
				t.Errorf("Matched = %d, want %d", got.Matched, tc.wantMatched)
			}
			if got.InGrace != tc.wantInGrace {
				t.Errorf("InGrace = %d, want %d", got.InGrace, tc.wantInGrace)
			}
			if got.VendorBilling != tc.wantBilling {
				t.Errorf("VendorBilling = %d, want %d", got.VendorBilling, tc.wantBilling)
			}
			if got.VendorDeleted != tc.wantDeleted {
				t.Errorf("VendorDeleted = %d, want %d", got.VendorDeleted, tc.wantDeleted)
			}
		})
	}
}

// TestClassifyThunderReconcile_UnknownAgeIsMarked pins the AgeUnknown flag:
// a finding built from an instance with no createdAt must SAY so, because
// the work-item reader will otherwise trust an age of zero minutes.
func TestClassifyThunderReconcile_UnknownAgeIsMarked(t *testing.T) {
	now := time.Date(2026, 8, 8, 18, 0, 0, 0, time.UTC)
	got := classifyThunderReconcile(
		[]thunderVendorInstance{{ID: "3", Status: "RUNNING"}},
		nil, 30*time.Minute, now)

	if len(got.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %v", got.Findings)
	}
	f := got.Findings[0]
	if !f.AgeUnknown {
		t.Errorf("AgeUnknown = false, want true for a zero CreatedAt")
	}
	if !strings.Contains(f.Summary, "UNKNOWN") {
		t.Errorf("summary should surface the unknown age, got: %s", f.Summary)
	}
}
