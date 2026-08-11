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

// Tests for siteUsesFlatURLs (bugs_open/241 plumbing). The helper is the ONE
// reader of the structure spec's url_shape key; the contract test at the foot
// pins that both canonicalisation surfaces actually consume it.

var siteURLShapeQuery = regexp.QuoteMeta(`
		SELECT COALESCE(data->>'url_shape', '')
		FROM site_specs
		WHERE site_id = $1 AND aspect = 'structure' AND is_current = true
	`)

func TestSiteUsesFlatURLs(t *testing.T) {
	siteID := uuid.New()

	cases := []struct {
		name string
		rows func(sqlmock.Sqlmock)
		want bool
	}{
		{
			name: "url_shape flat returns true",
			rows: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(siteURLShapeQuery).WithArgs(siteID).
					WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow("flat"))
			},
			want: true,
		},
		{
			name: "url_shape nested returns false",
			rows: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(siteURLShapeQuery).WithArgs(siteID).
					WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow("nested"))
			},
			want: false,
		},
		{
			name: "structure spec without the key returns false (COALESCE '')",
			rows: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(siteURLShapeQuery).WithArgs(siteID).
					WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(""))
			},
			want: false,
		},
		{
			name: "no current structure spec returns false",
			rows: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(siteURLShapeQuery).WithArgs(siteID).
					WillReturnRows(sqlmock.NewRows([]string{"coalesce"}))
			},
			want: false,
		},
		{
			name: "read error returns false, not a failure",
			rows: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(siteURLShapeQuery).WithArgs(siteID).
					WillReturnError(errors.New("connection reset"))
			},
			want: false,
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

			got := siteUsesFlatURLs(context.Background(), db, siteID, zap.NewNop())
			if got != tc.want {
				t.Errorf("siteUsesFlatURLs = %v, want %v", got, tc.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

func TestSiteUsesFlatURLsNilDB(t *testing.T) {
	if siteUsesFlatURLs(context.Background(), nil, uuid.New(), zap.NewNop()) {
		t.Error("nil db must mean the nested default, got flat")
	}
}

// TestFlatURLFlagReachesBothCanonicalisationSurfaces pins the invariant the
// helper's comment states: WriteSitePlanAction and SyncPagesToDBAction both
// read the site URL shape through siteUsesFlatURLs and both thread it into
// their CanonicalisePage descriptor. These two surfaces write the SAME page
// identities to site_plan_pages and pages respectively; one consuming the
// flag without the other re-opens the divergence that already shipped a
// regression once (the ValidateRoles comment in SyncPagesToDBAction).
//
// The descriptor literal is extracted ANCHORED at the CanonicalisePage call,
// not grepped bare, so a comment or another literal cannot satisfy it.
func TestFlatURLFlagReachesBothCanonicalisationSurfaces(t *testing.T) {
	descriptorCall := regexp.MustCompile(`(?s)CanonicalisePage\(datahelpers\.PageDescriptor\{([^}]*)\}`)

	for _, file := range []string{"write_site_plan_action.go", "site_db_actions.go"} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("cannot read %s: %v", file, err)
		}
		if !regexp.MustCompile(`siteUsesFlatURLs\(`).Match(src) {
			t.Errorf("%s no longer calls siteUsesFlatURLs — the URL shape must come from the one shared reader", file)
		}
		matches := descriptorCall.FindAllSubmatch(src, -1)
		if len(matches) == 0 {
			t.Fatalf("%s: no CanonicalisePage(datahelpers.PageDescriptor{...}) literal found — if the call was restructured, re-pin this test", file)
		}
		for _, m := range matches {
			if !regexp.MustCompile(`FlatURLs:`).Match(m[1]) {
				t.Errorf("%s: a CanonicalisePage descriptor omits FlatURLs — this surface would emit nested URLs while the other emits flat (descriptor body: %q)", file, string(m[1]))
			}
		}
	}
}
