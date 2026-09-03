// FILE: platform/orchestration/actions/refresh_evidence_unverifiable_test.go
//
// The sweep must ACCOUNT for every fact, and say so when a register is
// broken — bugs_closed/161's residual, measured 2026-09-03.
//
// Two findings are pinned here:
//
//  1. A fact that nothing re-proves used to fall through the residue arm
//     UNCOUNTED — no counter, no entry, no nudge — unless it happened to use
//     `attested_by`. 27 facts on 5 sites were in that state: the register
//     asserted them, the writer was instructed by them, every gate vouched
//     for them, and the one daily mechanism that could have said "nothing has
//     ever checked this" could not even count them.
//
//  2. A register that fails the typed parse raised nothing at all, and the
//     FactsChecked==0 early return sat in front of every raise — so the site
//     most likely to be broken was the site least able to report it.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// bareArtifactFact is the shape of the 27: a real artefact named in prose,
// with nothing that can re-prove it.
func bareArtifactFact(id, verifiedAt, artifact string) map[string]interface{} {
	return map[string]interface{}{
		"id":          id,
		"claim":       "configurable inputs in the drop-rate tuner",
		"value":       float64(4),
		"kind":        "count",
		"source":      map[string]interface{}{"artifact": artifact},
		"verified_at": verifiedAt,
	}
}

// TestUnverifiableArtifactFactIsNudgedWhenOld is the residue arm's whole point.
//
// MUTATION THAT MUST GO RED: restore the `if _, has := src["attested_by"]`
// guard around checkAttestationStaleness — this fact has no attested_by, so
// it is dropped uncounted and both assertions fail.
func TestUnverifiableArtifactFactIsNudgedWhenOld(t *testing.T) {
	fact := bareArtifactFact("gd-tuner-inputs", "2026-01-10", "drop-rate tuner UI: drop chance, kills per hour")
	src := fact["source"].(map[string]interface{})

	entry := checkAttestationStaleness(fact, "2026-09-03")
	if entry == nil {
		t.Fatal("a fact last dated 2026-01-10 is well past the 180-day cadence and must be due")
	}
	entry.Detail = unverifiableFactDetail(src, entry.Detail)

	if entry.Outcome != "attestation_due" {
		t.Fatalf("outcome = %q, want attestation_due", entry.Outcome)
	}
	// The nudge must never look like a verdict on the claim.
	if entry.VerifiedAt != "" {
		t.Fatal("a nudge must not bump verified_at — nothing was re-proved")
	}
	// It must say WHICH kind of unverifiable this is, or the human cannot act.
	if !strings.Contains(entry.Detail, "artifact_check") {
		t.Fatalf("an artifact fact's nudge must name the remedy, got %q", entry.Detail)
	}
	// NOT a Contains check on "human's word": the artifact remedy legitimately
	// ends "...retype it as attested_by if a human's word is the honest
	// source", so that needle cannot tell the two wordings apart. Stage 1's
	// message is identifiable by how it OPENS, which is what to assert.
	if strings.HasPrefix(entry.Detail, "attested_by fact") {
		t.Fatalf("an artifact fact got the attested_by wording: %q", entry.Detail)
	}
}

// TestUnverifiableFactNotYetDueIsStillCounted separates the two questions the
// result JSON must answer: how many need a human TODAY, and how big is the
// population nothing will ever check.
//
// MUTATION: move res.FactsUnverifiable++ inside the `if entry != nil` — a
// young fact then vanishes from the count and a run showing 0 becomes
// indistinguishable from the arm never having looked.
func TestUnverifiableFactNotYetDueIsStillCounted(t *testing.T) {
	fact := bareArtifactFact("rh-parameters", "2026-08-20", "MatchMatrix scoring implementation")
	if entry := checkAttestationStaleness(fact, "2026-09-03"); entry != nil {
		t.Fatalf("a fact dated 14 days ago is not due: %+v", entry)
	}
	// The counter is incremented by the caller for EVERY fact reaching the
	// arm, which the real-loop test below asserts end to end.
}

// TestUnverifiableDetailRoutesAUrlToCitation: 12 of the 27 name an external
// URL in `artifact`, and the mechanism they want already exists.
func TestUnverifiableDetailRoutesAUrlToCitation(t *testing.T) {
	got := unverifiableFactDetail(map[string]interface{}{"artifact": "https://trmagazine.es/panerai"}, "")
	if !strings.Contains(got, "citation") {
		t.Fatalf("a URL-sourced artifact fact must be pointed at citation, got %q", got)
	}
	// And a non-URL artefact must NOT be, or the advice is wrong.
	got = unverifiableFactDetail(map[string]interface{}{"artifact": "platform/x/y.go"}, "")
	if strings.Contains(got, "re-fetches that URL") {
		t.Fatalf("a repo-path fact must not be told to use citation, got %q", got)
	}
	// attested_by keeps stage 1's own wording verbatim.
	if got := unverifiableFactDetail(map[string]interface{}{"attested_by": "owner"}, "ORIGINAL"); got != "ORIGINAL" {
		t.Fatalf("an attested_by fact must keep its existing detail, got %q", got)
	}
	// A source with no recognised key is named as such, not silently ignored.
	if got := unverifiableFactDetail(map[string]interface{}{"nonsense": "x"}, ""); !strings.Contains(got, "no key") {
		t.Fatalf("an unrecognised source must say so, got %q", got)
	}
}

// TestUnverifiableFactsCountedInsideTheRealLoop drives refreshOneSiteEvidence
// itself: two bare artifact facts, one ancient and one fresh.
//
// MUTATION: delete the res.FactsUnverifiable++ line — FactsUnverifiable
// reads 0 and the population becomes invisible again, exactly as before.
func TestUnverifiableFactsCountedInsideTheRealLoop(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	site := uuid.New()
	eb, _ := json.Marshal(map[string]interface{}{
		"facts": []interface{}{
			bareArtifactFact("old-one", "2026-01-10", "some shipped component"),
			bareArtifactFact("young-one", "2026-08-20", "another shipped component"),
		},
	})

	mock.ExpectQuery("SELECT id, data, pinned FROM site_specs").WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"id", "data", "pinned"}).AddRow(uuid.New(), eb, false))
	mock.ExpectQuery("SELECT domain FROM sites").WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("gamesdesign.co.uk"))
	mock.ExpectQuery("SELECT to_char").
		WillReturnRows(sqlmock.NewRows([]string{"today"}).AddRow("2026-09-03"))
	mock.ExpectQuery(regexp.QuoteMeta(factDriftIndexQuery)).WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "build_status", "subject_key", "body", "fork_component_id"}))

	params := ActionParams{DB: db, Logger: zap.NewNop()}
	res, err := refreshOneSiteEvidence(context.Background(), params, site, true, zap.NewNop())
	if err != nil {
		t.Fatalf("refreshOneSiteEvidence: %v", err)
	}

	if res.FactsUnverifiable != 2 {
		t.Fatalf("BOTH facts are unverifiable whether or not they are due: got %d, want 2", res.FactsUnverifiable)
	}
	if res.AttestationsDue != 1 {
		t.Fatalf("only the ancient one is due today: got %d, want 1", res.AttestationsDue)
	}
	// The due one is reported, and it is the OLD one.
	found := ""
	for _, f := range res.Facts {
		if f.Outcome == "attestation_due" {
			found = f.FactID
		}
	}
	if found != "old-one" {
		t.Fatalf("the due entry must be the ancient fact, got %q", found)
	}
}

// TestMalformedRegisterRaisesAnItemAndOnlyWhenNotDryRun pins the second
// finding end to end. A single text-valued fact used to void the register.
//
// MUTATION: restore `if res.FactsChecked == 0 { return res, nil }` — this
// register has no checkable fact at all, so the function returns before the
// raise and the expected INSERT never happens.
func TestMalformedRegisterRaisesAnItemAndOnlyWhenNotDryRun(t *testing.T) {
	// noted.co.uk's real shape: a text value, plus the bans it took down.
	ebJSON, _ := json.Marshal(map[string]interface{}{
		"facts": []interface{}{map[string]interface{}{
			"id": "noted-backup", "claim": "deleted text persists at most 30 days", "value": "30 days",
		}},
		"banned_claims": []interface{}{
			map[string]interface{}{"pattern": "end[- ]?to[- ]?end encrypt", "reason": "Not built."},
			map[string]interface{}{"pattern": "gdpr[- ]compliant", "reason": "No auditor."},
		},
	})

	expect := func(m sqlmock.Sqlmock, site uuid.UUID) {
		m.ExpectQuery("SELECT id, data, pinned FROM site_specs").WithArgs(site).
			WillReturnRows(sqlmock.NewRows([]string{"id", "data", "pinned"}).AddRow(uuid.New(), ebJSON, false))
		m.ExpectQuery("SELECT domain FROM sites").WithArgs(site).
			WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("noted.co.uk"))
		m.ExpectQuery("SELECT to_char").
			WillReturnRows(sqlmock.NewRows([]string{"today"}).AddRow("2026-09-03"))
		m.ExpectQuery(regexp.QuoteMeta(factDriftIndexQuery)).WithArgs(site).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "build_status", "subject_key", "body", "fork_component_id"}))
	}

	t.Run("dry run reports but writes nothing", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		site := uuid.New()
		expect(mock, site)

		res, err := refreshOneSiteEvidence(context.Background(),
			ActionParams{DB: db, Logger: zap.NewNop()}, site, true, zap.NewNop())
		if err != nil {
			t.Fatalf("refreshOneSiteEvidence: %v", err)
		}
		if len(res.MalformedFacts) != 1 {
			t.Fatalf("the undecodable fact must be reported: %+v", res.MalformedFacts)
		}
		if res.MalformedFacts[0].ID != "noted-backup" {
			t.Fatalf("the fact must be NAMED, got %+v", res.MalformedFacts[0])
		}
		if !strings.Contains(res.MalformedFacts[0].Err, "value") {
			t.Fatalf("the decode error must name the field, got %q", res.MalformedFacts[0].Err)
		}
		if res.MalformedFactWorkItemsCreated != 0 {
			t.Fatal("a dry run must write nothing")
		}
		// And must not even OPEN a transaction — no ExpectBegin was set, so
		// sqlmock fails the run if one is attempted.
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unexpected queries: %v", err)
		}
	})

	t.Run("a real run raises one item naming the fact", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		site := uuid.New()
		expect(mock, site)
		// THE ASSERTION IS THAT THE RAISE IS ATTEMPTED, not that a row lands.
		// writeWorkItem is not one statement — it runs several probe queries
		// and an ExecContext — so mocking it through would couple this test to
		// that helper's internals and break whenever it gains a probe. What
		// this change is responsible for is REACHING it, and nothing else on
		// this path opens a transaction: the Begin IS the signal. The write
		// then fails against the bare mock, which the code handles by logging
		// and rolling back, so the count stays 0 and must not be asserted as 1.
		mock.ExpectBegin()
		mock.ExpectRollback()

		res, err := refreshOneSiteEvidence(context.Background(),
			ActionParams{DB: db, Logger: zap.NewNop()}, site, false, zap.NewNop())
		if err != nil {
			t.Fatalf("refreshOneSiteEvidence: %v", err)
		}
		if len(res.MalformedFacts) != 1 {
			t.Fatalf("the undecodable fact must still be reported on a real run: %+v", res.MalformedFacts)
		}
		// ExpectationsWereMet is the whole test: the Begin proves control
		// reached createMalformedEvidenceFactItems. MUTATION: restore the bare
		// `if res.FactsChecked == 0 { return res, nil }` early return and this
		// fails with the Begin unmet, because the function returns first.
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("the raise was never attempted: %v", err)
		}
	})
}

// TestMalformedFactItemKeyIsPerFactNotPerSite: a second bad fact must not hide
// behind an item already open for the first — the lesson bugs_open/091 taught
// the sibling invalid-pattern writer, applied here before it can bite.
//
// MUTATION: key on siteID alone — both keys collide and this fails.
func TestMalformedFactItemKeyIsPerFactNotPerSite(t *testing.T) {
	site := uuid.New()
	a := malformedFactItemKey(site, datahelpers.MalformedFact{ID: "ft-licence-mit", Index: 0})
	b := malformedFactItemKey(site, datahelpers.MalformedFact{ID: "ft-licence-apache", Index: 1})
	if a == b {
		t.Fatal("two different malformed facts must not share one item key")
	}
	// Stable across passes, or every daily sweep raises a duplicate.
	if a != malformedFactItemKey(site, datahelpers.MalformedFact{ID: "ft-licence-mit", Index: 0}) {
		t.Fatal("the key must be stable for the same fact")
	}
	if !strings.HasPrefix(a, "malformed_evidence_fact:"+site.String()) {
		t.Fatalf("key must be site-scoped and typed, got %q", a)
	}
}

// TestEarlyReturnStillSilentOnACleanSite is the other side of the early-return
// change: a site with nothing to check AND nothing to report must still write
// nothing and query nothing further. Without this, the change trades one
// silence for a lot of noise.
func TestEarlyReturnStillSilentOnACleanSite(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	site := uuid.New()
	eb, _ := json.Marshal(map[string]interface{}{"facts": []interface{}{}, "banned_claims": []interface{}{}})

	mock.ExpectQuery("SELECT id, data, pinned FROM site_specs").WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"id", "data", "pinned"}).AddRow(uuid.New(), eb, false))
	mock.ExpectQuery("SELECT domain FROM sites").WithArgs(site).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("finetuning.uk"))
	mock.ExpectQuery("SELECT to_char").
		WillReturnRows(sqlmock.NewRows([]string{"today"}).AddRow("2026-09-03"))

	res, err := refreshOneSiteEvidence(context.Background(),
		ActionParams{DB: db, Logger: zap.NewNop()}, site, false, zap.NewNop())
	if err != nil {
		t.Fatalf("refreshOneSiteEvidence: %v", err)
	}
	if res.FactsChecked != 0 || res.MalformedFactWorkItemsCreated != 0 {
		t.Fatalf("an empty register must stay silent: %+v", res)
	}
	// No further queries — the early return still fires.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the early return did not fire on a clean empty site: %v", err)
	}
}

var _ = sql.ErrNoRows
