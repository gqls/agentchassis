// FILE: platform/orchestration/actions/discovery_checks/check_build_prerequisites_test.go
//
// RFC_056 prerequisites seat. Two things are worth pinning here and a sqlmock
// happy path cannot pin either on its own:
//
//   - the contract that every failure is an ABSENCE observed by SQL, so the
//     emptiness guard on evidence_base ('{}' counts as absent) lives in the
//     predicate TEXT — sqlmock returns whatever it is handed regardless of the
//     WHERE clause, so TestBuildPrerequisitesPredicateText reads the SQL, and
//     the '{}' case in the verdict table below only shows the check honours a
//     false answer (check_missing_structure_test.go's header, same reasoning);
//   - the RFC_010 safety property: a retraction fires only on a positive
//     observation, and an errored run retracts nothing and files nothing.
package discovery_checks

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const buildPrerequisitesTestSite = "5c2b0a1e-9d3f-4b7a-8e61-2f0c4d9a7b13"

func newBuildPrerequisitesCtx(t *testing.T) (DiscoveryCheckContext, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return DiscoveryCheckContext{
		Ctx:       context.Background(),
		DB:        db,
		SiteID:    uuid.MustParse(buildPrerequisitesTestSite),
		Pipeline:  "build",
		AgentType: "completeness-discovery-agent",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	}, mock
}

// buildPrerequisitesMatchers is one regex per kind, in the ORDER the check
// issues them — sqlmock enforces order by default, so a reordering of
// buildPrerequisiteKinds shows up here as an unmatched expectation rather
// than a silently wrong verdict.
var buildPrerequisitesMatchers = []struct{ kind, sql string }{
	{"vertical_landscape", `aspect = 'vertical_landscape'`},
	{"page_research", `FROM research_results`},
	{"evidence_base", `aspect = 'evidence_base'`},
	{"feed_sources", `FROM content_sources`},
}

// expectBuildPrerequisites queues the four EXISTS answers in issue order.
func expectBuildPrerequisites(mock sqlmock.Sqlmock, siteID uuid.UUID, present map[string]bool) {
	for _, m := range buildPrerequisitesMatchers {
		mock.ExpectQuery(m.sql).WithArgs(siteID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(present[m.kind]))
	}
}

func prerequisiteKey(kind string) string {
	return "prerequisite_missing:" + kind + ":" + buildPrerequisitesTestSite
}

func TestBuildPrerequisitesIsRegistered(t *testing.T) {
	c := Get("build_prerequisites")
	if c == nil {
		t.Fatal("build_prerequisites is not registered — init() did not run or Name() changed")
	}
	if c.Name() != "build_prerequisites" {
		t.Errorf("Name() = %q, want build_prerequisites", c.Name())
	}
}

// TestBuildPrerequisitesVerdicts is the table: which kinds are observed
// present, and exactly which keys are filed and which are retracted.
func TestBuildPrerequisitesVerdicts(t *testing.T) {
	all := func(v bool) map[string]bool {
		return map[string]bool{"vertical_landscape": v, "page_research": v, "evidence_base": v, "feed_sources": v}
	}
	cases := []struct {
		name         string
		present      map[string]bool
		wantMissing  []string
		wantResolved []string
	}{
		{
			name:         "all four present: nothing filed, four retractions",
			present:      all(true),
			wantMissing:  nil,
			wantResolved: []string{"vertical_landscape", "page_research", "evidence_base", "feed_sources"},
		},
		{
			name:         "all four absent: four verdicts, no retraction",
			present:      all(false),
			wantMissing:  []string{"vertical_landscape", "page_research", "evidence_base", "feed_sources"},
			wantResolved: nil,
		},
		{
			// The DB answers false for a current evidence_base row whose data is
			// '{}' — that is the predicate's job (pinned by the text test below);
			// this row shows the check files on that answer and retracts the rest.
			name:         "evidence_base present but empty ('{}') counts as absent",
			present:      map[string]bool{"vertical_landscape": true, "page_research": true, "evidence_base": false, "feed_sources": true},
			wantMissing:  []string{"evidence_base"},
			wantResolved: []string{"vertical_landscape", "page_research", "feed_sources"},
		},
		{
			// The two route defects the header names: on the greenfield route
			// nothing requests research and nothing enrols feeds.
			name:         "greenfield route shape: research and feeds absent",
			present:      map[string]bool{"vertical_landscape": true, "page_research": false, "evidence_base": true, "feed_sources": false},
			wantMissing:  []string{"page_research", "feed_sources"},
			wantResolved: []string{"vertical_landscape", "evidence_base"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dctx, mock := newBuildPrerequisitesCtx(t)
			expectBuildPrerequisites(mock, dctx.SiteID, tc.present)

			res, err := (&BuildPrerequisitesCheck{}).Run(dctx)
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}

			// One finding per kind, every run, carrying the observation either way.
			if len(res.Findings) != len(buildPrerequisiteKinds) {
				t.Fatalf("want %d findings (one per kind), got %d", len(buildPrerequisiteKinds), len(res.Findings))
			}
			for _, f := range res.Findings {
				kind, _ := f["kind"].(string)
				got, ok := f["present"].(bool)
				if !ok {
					t.Errorf("finding %q has no boolean present field: %v", kind, f)
					continue
				}
				if got != tc.present[kind] {
					t.Errorf("finding %q present=%v, want %v", kind, got, tc.present[kind])
				}
			}

			// Work items: exact keys, flag-only shape.
			var gotMissing []string
			for _, wi := range res.WorkItems {
				parts := strings.Split(wi.ItemKey, ":")
				if len(parts) != 3 {
					t.Errorf("ItemKey %q is not prerequisite_missing:<kind>:<site>", wi.ItemKey)
					continue
				}
				gotMissing = append(gotMissing, parts[1])
				if wi.ItemKey != prerequisiteKey(parts[1]) {
					t.Errorf("ItemKey = %q, want %q", wi.ItemKey, prerequisiteKey(parts[1]))
				}
				if wi.ItemType != "prerequisite_missing" {
					t.Errorf("ItemType = %q, want prerequisite_missing", wi.ItemType)
				}
				if wi.HandlerAgent != "" {
					t.Errorf("HandlerAgent = %q, want \"\" — this is a verdict row, nothing may be routed at it", wi.HandlerAgent)
				}
				if wi.Status != "detected" {
					t.Errorf("Status = %q, want detected", wi.Status)
				}
				if wi.Severity != "medium" || wi.Priority != 120 {
					t.Errorf("Severity/Priority = %q/%d, want medium/120", wi.Severity, wi.Priority)
				}
				if wi.Pipeline != dctx.Pipeline || wi.SiteID != dctx.SiteID || wi.BatchID != dctx.BatchID || wi.CreatedBy != dctx.AgentType || wi.Source != "discovery" {
					t.Errorf("item does not carry the run's context: %+v", wi)
				}
				if !strings.Contains(wi.Summary, parts[1]) {
					t.Errorf("Summary must name the kind %q; got %q", parts[1], wi.Summary)
				}

				var spec map[string]interface{}
				if err := json.Unmarshal([]byte(wi.SpecJSON), &spec); err != nil {
					t.Fatalf("SpecJSON is not JSON: %v\n%s", err, wi.SpecJSON)
				}
				for _, key := range []string{"kind", "predicate", "supplier", "seat", "rfc", "not_dispatchable"} {
					if s, _ := spec[key].(string); s == "" {
						t.Errorf("spec for %q lacks %q: %s", parts[1], key, wi.SpecJSON)
					}
				}
				if spec["kind"] != parts[1] || spec["seat"] != "prerequisites" || spec["rfc"] != "RFC_056" {
					t.Errorf("spec identity wrong for %q: kind=%v seat=%v rfc=%v", parts[1], spec["kind"], spec["seat"], spec["rfc"])
				}
				nd, _ := spec["not_dispatchable"].(string)
				if !strings.Contains(nd, "never promote") || !strings.Contains(nd, "380") {
					t.Errorf("not_dispatchable must say never-promote and cite bugs_open/380 D1; got %q", nd)
				}
				if !strings.Contains(wi.Summary, spec["supplier"].(string)) {
					t.Errorf("Summary must name the supplier %q; got %q", spec["supplier"], wi.Summary)
				}
			}
			if strings.Join(gotMissing, ",") != strings.Join(tc.wantMissing, ",") {
				t.Errorf("filed kinds = %v, want %v", gotMissing, tc.wantMissing)
			}

			// Retractions: only what was positively observed, under the same key.
			var gotResolved []string
			for _, r := range res.Resolved {
				parts := strings.Split(r.ItemKey, ":")
				if len(parts) != 3 {
					t.Errorf("Resolved ItemKey %q is not prerequisite_missing:<kind>:<site>", r.ItemKey)
					continue
				}
				gotResolved = append(gotResolved, parts[1])
				if r.ItemType != "prerequisite_missing" || r.ItemKey != prerequisiteKey(parts[1]) {
					t.Errorf("retraction targets %q/%q, want prerequisite_missing/%q", r.ItemType, r.ItemKey, prerequisiteKey(parts[1]))
				}
				if r.AllOfType {
					t.Errorf("retraction for %q must be narrow (ItemKey), not AllOfType — the other kinds may be absent on the same run", parts[1])
				}
				if !strings.Contains(r.Reason, parts[1]) || !strings.Contains(r.Reason, "positively observed present") {
					t.Errorf("Reason must state the positive observation of %q; got %q", parts[1], r.Reason)
				}
			}
			if strings.Join(gotResolved, ",") != strings.Join(tc.wantResolved, ",") {
				t.Errorf("retracted kinds = %v, want %v", gotResolved, tc.wantResolved)
			}
		})
	}
}

// TestBuildPrerequisitesQueryErrorFilesNothing — RFC_010's safety property. A
// predicate that cannot be judged aborts the whole run: no partial verdicts
// (the kinds already observed present are NOT retracted, the ones already
// observed absent are NOT filed), and the error is returned so the runner skips
// Resolved on its side as well.
func TestBuildPrerequisitesQueryErrorFilesNothing(t *testing.T) {
	dctx, mock := newBuildPrerequisitesCtx(t)
	mock.ExpectQuery(`aspect = 'vertical_landscape'`).WithArgs(dctx.SiteID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(`FROM research_results`).WithArgs(dctx.SiteID).
		WillReturnError(errors.New("relation \"research_results\" does not exist"))

	res, err := (&BuildPrerequisitesCheck{}).Run(dctx)
	if err == nil {
		t.Fatal("want the predicate error propagated, got nil")
	}
	if !strings.Contains(err.Error(), "page_research") {
		t.Errorf("error must name the kind that could not be judged; got %v", err)
	}
	if res != nil {
		t.Errorf("an errored run must return no result at all (no items, no retractions); got %+v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestBuildPrerequisitesItemKeyPrefixIsItemType pins the "{item_type}:{target}"
// contract: the literal item type IS the key's first segment, for filed items
// and retractions alike, so the dedup index and the RFC_010 retraction meet on
// one string.
func TestBuildPrerequisitesItemKeyPrefixIsItemType(t *testing.T) {
	dctx, mock := newBuildPrerequisitesCtx(t)
	expectBuildPrerequisites(mock, dctx.SiteID, map[string]bool{
		"vertical_landscape": false, "page_research": true, "evidence_base": false, "feed_sources": true,
	})
	res, err := (&BuildPrerequisitesCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 2 || len(res.Resolved) != 2 {
		t.Fatalf("want 2 items + 2 retractions, got %d + %d", len(res.WorkItems), len(res.Resolved))
	}
	for _, wi := range res.WorkItems {
		if !strings.HasPrefix(wi.ItemKey, wi.ItemType+":") {
			t.Errorf("ItemKey %q does not start with ItemType %q + ':'", wi.ItemKey, wi.ItemType)
		}
		if !strings.HasSuffix(wi.ItemKey, ":"+dctx.SiteID.String()) {
			t.Errorf("ItemKey %q does not end with the site id", wi.ItemKey)
		}
	}
	for _, r := range res.Resolved {
		if !strings.HasPrefix(r.ItemKey, r.ItemType+":") {
			t.Errorf("Resolved ItemKey %q does not start with ItemType %q + ':'", r.ItemKey, r.ItemType)
		}
	}
}

// TestBuildPrerequisitesPredicateText is the anti-regression half: the SQL
// issued is the whole check, and sqlmock cannot see it. Every predicate must be
// site-scoped and return one EXISTS boolean; evidence_base must refuse an
// empty register; nothing here may read pages (this seat is about what the
// site was built FROM, not what it serves).
func TestBuildPrerequisitesPredicateText(t *testing.T) {
	if len(buildPrerequisiteKinds) != 4 {
		t.Fatalf("want exactly 4 kinds, got %d — update the header baseline and this test together", len(buildPrerequisiteKinds))
	}
	wantTable := map[string]string{
		"vertical_landscape": `FROM site_specs WHERE site_id = \$1 AND aspect = 'vertical_landscape' AND is_current`,
		"page_research":      `FROM research_results WHERE site_id = \$1`,
		"evidence_base":      `FROM site_specs WHERE site_id = \$1 AND aspect = 'evidence_base' AND is_current`,
		"feed_sources":       `FROM content_sources WHERE site_id = \$1 AND is_active`,
	}
	for _, k := range buildPrerequisiteKinds {
		if k.Supplier == "" || k.Absence == "" {
			t.Errorf("%s: supplier and absence text are both required", k.Kind)
		}
		if !strings.HasPrefix(k.Predicate, "SELECT EXISTS (") {
			t.Errorf("%s: predicate must be a single EXISTS boolean; got %q", k.Kind, k.Predicate)
		}
		if !regexp.MustCompile(`site_id = \$1`).MatchString(k.Predicate) {
			t.Errorf("%s: predicate must be scoped to $1 = site_id; got %q", k.Kind, k.Predicate)
		}
		want, ok := wantTable[k.Kind]
		if !ok {
			t.Errorf("unexpected kind %q — the verifier coverage entry names four", k.Kind)
			continue
		}
		if !regexp.MustCompile(want).MatchString(k.Predicate) {
			t.Errorf("%s: predicate does not read the expected table/columns: /%s/ did not match %q", k.Kind, want, k.Predicate)
		}
		if regexp.MustCompile(`(?i)(FROM|JOIN)\s+pages\b`).MatchString(k.Predicate) {
			t.Errorf("%s: predicate must not read pages; got %q", k.Kind, k.Predicate)
		}
		if k.Kind == "evidence_base" {
			for _, guard := range []string{`jsonb_typeof\(data\) = 'object'`, `data <> '\{\}'::jsonb`} {
				if !regexp.MustCompile(guard).MatchString(k.Predicate) {
					t.Errorf("evidence_base: an empty '{}' register must count as ABSENT — guard /%s/ missing from %q", guard, k.Predicate)
				}
			}
		}
	}
}
