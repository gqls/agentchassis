// FILE: platform/orchestration/actions/rerender_page_sections_base_data_test.go
//
// Guards the light re-render path's contact-address source (bugs_open/006 §B).
// buildRerenderBaseData must prefer the canonical sites.email COLUMN over
// content_data.email, so a section re-render converts a dead contact form to a
// mailto to the SAME address the full-writer path (loadSiteDataFull) would use.
// Before this, the light path read only content_data.email, which is empty on
// most sites and stale on idea.uk (idea-uk@leopardess.uk).

package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestBuildRerenderBaseDataPrefersSitesEmailColumn(t *testing.T) {
	siteID := uuid.New()

	cases := []struct {
		name        string
		contentData string
		colEmail    string
		wantEmail   string
	}{
		{
			name:        "sites.email column wins over a stale content_data.email",
			contentData: `{"email":"idea-uk@leopardess.uk","company_name":"Acme"}`,
			colEmail:    "idea.uk@contactforsales.com",
			wantEmail:   "idea.uk@contactforsales.com",
		},
		{
			name:        "empty column falls back to content_data.email",
			contentData: `{"email":"cd@example.com"}`,
			colEmail:    "",
			wantEmail:   "cd@example.com",
		},
		{
			name:        "empty content_data still gets the column address",
			contentData: `{}`,
			colEmail:    "col@contactforsales.com",
			wantEmail:   "col@contactforsales.com",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()

			mock.ExpectQuery("SELECT content_data").
				WithArgs(sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"content_data", "email"}).
					AddRow([]byte(tc.contentData), tc.colEmail))

			base := buildRerenderBaseData(context.Background(), db, siteID, "example.com", "index", zap.NewNop())

			if got, _ := base["email"].(string); got != tc.wantEmail {
				t.Errorf("base[email] = %q, want %q", got, tc.wantEmail)
			}
			// contact_email is set from the column only when the column is non-empty.
			if tc.colEmail != "" {
				if got, _ := base["contact_email"].(string); got != tc.colEmail {
					t.Errorf("base[contact_email] = %q, want %q", got, tc.colEmail)
				}
			}
		})
	}
}

// TestBuildRerenderBaseDataCarriesPageIdentity guards the SECOND instance of
// bugs_open/085, on the scoped section-re-render path.
//
// 085 was fixed on the page-BUILD path (BuildRenderContextAction) and shipped in
// v1.0.1173. This path was still broken, and reading the code was not what found
// it: firing a scoped re-render on the FIXED binary re-rendered fundamentallyai's
// index at 14:08 and it still carried all three charts, two of which are assigned
// to a different page. A fix applied to one branch of a two-branch router reads as
// done while the other branch keeps the bug (016b §9).
//
// The page name was already in scope at the call site — it is passed to
// newSourceResolver on the line above — so the identity was available here all
// along. Setting the map key is sufficient because mergeIntoRenderContext restores
// it into RenderContext.CurrentPage.
func TestBuildRerenderBaseDataCarriesPageIdentity(t *testing.T) {
	siteID := uuid.New()

	cases := []struct {
		name     string
		pageName string
		want     string
		wantKey  bool
	}{
		{name: "plain page name", pageName: "capabilities", want: "capabilities", wantKey: true},
		{name: "the .html suffix is stripped to match buildHeaderConfig", pageName: "index.html", want: "index", wantKey: true},
		{name: "no page name: the key is absent, not an empty string", pageName: "", wantKey: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()

			mock.ExpectQuery("SELECT content_data").
				WithArgs(sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"content_data", "email"}).
					AddRow([]byte(`{}`), ""))

			base := buildRerenderBaseData(context.Background(), db, siteID, "example.com", tc.pageName, zap.NewNop())

			got, present := base["current_page"]
			if present != tc.wantKey {
				t.Fatalf("base has current_page = %v, want %v", present, tc.wantKey)
			}
			if tc.wantKey {
				if s, _ := got.(string); s != tc.want {
					t.Errorf("base[current_page] = %q, want %q", s, tc.want)
				}
				// The value is only useful if it survives into the struct field —
				// the regex-fallback renderer reads that, not ContentData.
				rc := &RenderContext{}
				mergeIntoRenderContext(rc, base)
				if rc.CurrentPage != tc.want {
					t.Errorf("after mergeIntoRenderContext, CurrentPage = %q, want %q", rc.CurrentPage, tc.want)
				}
			}
		})
	}
}
