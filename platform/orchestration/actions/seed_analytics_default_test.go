package actions

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// The guards live in the SQL, where Postgres enforces them; these needles assert the
// executed statement still carries each one. They match the SQL const the mock receives,
// not comments (each needle is a clause no comment contains verbatim).
var seedGuardNeedles = []string{
	`COALESCE(n.settings->'analytics'->>'gtm_container_id', '') <> ''`, // no network default, no seed
	`NOT EXISTS`, // never a second current row, never override mode=none
	`ss.aspect = 'site_config' AND ss.is_current`,
	`'mode', 'default'`, // seeded values are distinguishable from operator-set ones
}

func TestSeedAnalyticsDefaultGuardsAreInTheExecutedSQL(t *testing.T) {
	for _, needle := range seedGuardNeedles {
		if !strings.Contains(seedAnalyticsDefaultSQL, needle) {
			t.Errorf("guard clause missing from seedAnalyticsDefaultSQL: %q", needle)
		}
	}
}

func TestSeedAnalyticsDefaultReportsSeededOnlyWhenARowWasWritten(t *testing.T) {
	siteID, netID := uuid.New(), uuid.New()
	for _, tc := range []struct {
		name   string
		rows   int64
		seeded bool
	}{
		{"guards passed, row written", 1, true},
		{"a guard blocked it (existing row / no network default)", 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectExec(regexp.QuoteMeta("INSERT INTO site_specs")).
				WithArgs(siteID, netID).
				WillReturnResult(sqlmock.NewResult(0, tc.rows))
			seeded, err := seedAnalyticsDefault(context.Background(), db, siteID, netID, zap.NewNop())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if seeded != tc.seeded {
				t.Errorf("seeded=%v, want %v", seeded, tc.seeded)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSeedAnalyticsDefaultRefusesAnUnknownDBType(t *testing.T) {
	if _, err := seedAnalyticsDefault(context.Background(), struct{}{}, uuid.New(), uuid.New(), zap.NewNop()); err == nil {
		t.Fatal("want an error for an unsupported db type, got nil")
	}
}
