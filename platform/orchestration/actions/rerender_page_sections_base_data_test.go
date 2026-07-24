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

			base := buildRerenderBaseData(context.Background(), db, siteID, "example.com", zap.NewNop())

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
