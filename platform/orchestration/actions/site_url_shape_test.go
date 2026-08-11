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
// The round-2 REVISE (corr 70256656) widened the pin to every URL-consuming
// caller that operates on an existing site: apply_gap_plan's new_page,
// create_tool_component, deploy_tool's resolveToolPageIdentity. Deliberately
// NOT pinned, with the reason: create_blog_posts_action.go and deploy_tool's
// companionGuideIdentity synthesise role=blog-post identities, which already
// emit the flat shape — FlatURLs is a no-op for them by construction
// (nestedOrFlatURL is only consulted for tool/guide/game); and
// apply_adoption_plan_action.go's recreation path canonicalises to nested by
// design while the structure spec is being written in the same transaction.
//
// The descriptor literal is extracted ANCHORED at the CanonicalisePage call,
// not grepped bare, so a comment or another literal cannot satisfy it.
func TestFlatURLFlagReachesBothCanonicalisationSurfaces(t *testing.T) {
	descriptorCall := regexp.MustCompile(`(?s)CanonicalisePage\(datahelpers\.PageDescriptor\{([^}]*)\}`)

	// file -> how many CanonicalisePage descriptor literals must carry
	// FlatURLs. deploy_tool_action.go has two descriptors and exactly one
	// (resolveToolPageIdentity's) must carry it — companionGuideIdentity's
	// blog-post descriptor must NOT be forced to, see above.
	type pin struct {
		file    string
		flagged int
		total   int
	}
	for _, p := range []pin{
		{"write_site_plan_action.go", 1, 1},
		{"site_db_actions.go", 1, 1},
		{"apply_gap_plan_action.go", 1, 1},
		{"create_tool_component_action.go", 1, 1},
		{"deploy_tool_action.go", 1, 2},
	} {
		src, err := os.ReadFile(p.file)
		if err != nil {
			t.Fatalf("cannot read %s: %v", p.file, err)
		}
		if !regexp.MustCompile(`siteUsesFlatURLs\(`).Match(src) {
			t.Errorf("%s no longer calls siteUsesFlatURLs — the URL shape must come from the one shared reader", p.file)
		}
		matches := descriptorCall.FindAllSubmatch(src, -1)
		if len(matches) != p.total {
			t.Fatalf("%s: found %d CanonicalisePage descriptor literals, expected %d — a caller was added or restructured; re-pin this test deliberately",
				p.file, len(matches), p.total)
		}
		flagged := 0
		for _, m := range matches {
			if regexp.MustCompile(`FlatURLs:`).Match(m[1]) {
				flagged++
			}
		}
		if flagged != p.flagged {
			t.Errorf("%s: %d of %d descriptors carry FlatURLs, expected %d — an unflagged surface emits nested URLs on a flat site",
				p.file, flagged, p.total, p.flagged)
		}
	}
}
