// FILE: platform/orchestration/actions/discovery_checks/check_contact_form_undeliverable_test.go
//
// Guards the auto-remediation gating added for bugs_open/006 §B
// (PLAN_2026-07-24_contact_form_hardening): a contact form on a site WITH a
// resolvable sites.email is auto-healed via a light page_rerender; a site with
// no honest address is parked for a human, unchanged. The resolvability guard is
// kept in lockstep with deliverableFormAction (component_library.go).

package discovery_checks

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// contactAddressResolvable is the branch predicate. It must agree with
// deliverableFormAction's guard: a real address the render seam will convert to
// a mailto, refusing the synthesised info@<own-domain> fallback.
func TestContactAddressResolvable(t *testing.T) {
	cases := []struct {
		name          string
		email, domain string
		want          bool
	}{
		{"a real address resolves", "vonc@contactforsales.com", "vonc.com", true},
		{"empty is not resolvable", "", "vonc.com", false},
		{"malformed (no @) is not resolvable", "not-an-address", "vonc.com", false},
		{"synthesised info@own-domain is refused", "info@robot-hands.com", "robot-hands.com", false},
		{"info@own-domain refused case-insensitively", "INFO@Robot-Hands.com", "robot-hands.com", false},
		{"a real info@ on a DIFFERENT domain is honoured", "info@leopardess.uk", "robot-hands.com", true},
		{"surrounding whitespace is trimmed", "  vonc@contactforsales.com  ", "vonc.com", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := contactAddressResolvable(tc.email, tc.domain); got != tc.want {
				t.Errorf("contactAddressResolvable(%q, %q) = %v, want %v", tc.email, tc.domain, got, tc.want)
			}
		})
	}
}

// TestContactFormUndeliverableRoutesByResolvability is the end-to-end guard:
// the SAME broken form routes to an auto-healing page_rerender when the site has
// a resolvable address, and to needs_human_review when it does not.
func TestContactFormUndeliverableRoutesByResolvability(t *testing.T) {
	siteID := uuid.New()
	pageID := uuid.New()

	cases := []struct {
		name          string
		email, domain string
		wantItemType  string
		wantHandler   string
		wantStatus    string
		wantKeyPrefix string
	}{
		{
			name:  "resolvable address → auto re-render",
			email: "vonc@contactforsales.com", domain: "vonc.com",
			wantItemType: "page_rerender", wantHandler: "page-rerender",
			wantStatus: "detected", wantKeyPrefix: "contact_form_undeliverable_rerender:",
		},
		{
			name:  "no address → parked for a human",
			email: "", domain: "robot-hands.com",
			wantItemType: "contact_form_undeliverable", wantHandler: "",
			wantStatus: "needs_human_review", wantKeyPrefix: "contact_form_undeliverable:",
		},
		{
			name:  "synthesised info@own-domain → parked for a human (not auto-healed to a fake inbox)",
			email: "info@robot-hands.com", domain: "robot-hands.com",
			wantItemType: "contact_form_undeliverable", wantHandler: "",
			wantStatus: "needs_human_review", wantKeyPrefix: "contact_form_undeliverable:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()

			// 1) the findings query returns one broken contact form
			mock.ExpectQuery("data-component").
				WithArgs(sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slot_name", "position", "action_target"}).
					AddRow(pageID.String(), "contact", "main", 0, "#contact"))
			// 2) resolveSiteContact reads the sites.email column + domain
			mock.ExpectQuery("FROM sites WHERE id").
				WithArgs(sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"email", "domain"}).
					AddRow(tc.email, tc.domain))

			c := &ContactFormUndeliverableCheck{}
			res, err := c.Run(DiscoveryCheckContext{
				Ctx: context.Background(), DB: db, SiteID: siteID,
				AgentType: "test-agent", BatchID: uuid.New(), Logger: zap.NewNop(),
			})
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if len(res.WorkItems) != 1 {
				t.Fatalf("want exactly 1 work item, got %d", len(res.WorkItems))
			}
			wi := res.WorkItems[0]
			if wi.ItemType != tc.wantItemType {
				t.Errorf("ItemType = %q, want %q", wi.ItemType, tc.wantItemType)
			}
			if wi.HandlerAgent != tc.wantHandler {
				t.Errorf("HandlerAgent = %q, want %q", wi.HandlerAgent, tc.wantHandler)
			}
			if wi.Status != tc.wantStatus {
				t.Errorf("Status = %q, want %q", wi.Status, tc.wantStatus)
			}
			if !strings.HasPrefix(wi.ItemKey, tc.wantKeyPrefix) {
				t.Errorf("ItemKey = %q, want prefix %q", wi.ItemKey, tc.wantKeyPrefix)
			}
		})
	}
}
