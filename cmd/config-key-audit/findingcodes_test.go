// FILE: cmd/config-key-audit/findingcodes_test.go
//
// bugs_open/358. NON-VACUITY IS THE POINT OF THIS FILE, not a nicety.
//
// This mode's own subject is checks that cannot fail — codes written as if they
// were findings, which nothing ever reads. A version of THIS check that always
// passes would be the bug reproducing itself one level up, and it would look
// exactly like a healthy estate. So every case below is stated as a mutation
// WITH the control that must come out the other way: it is not enough to show
// the check fires, it must be shown that the same input with the defect removed
// does not.
package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

var testNow = time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)

// srcContaining fakes the tree: a reader reference resolves to whatever body
// the map gives it. Injected rather than read from disk so these tests pin the
// LOGIC, and the real read is exercised by repoSourceReader in the live run.
func srcContaining(bodies map[string]string) sourceReader {
	return func(fileLine string) (string, error) {
		if b, ok := bodies[fileLine]; ok {
			return b, nil
		}
		return "", fmt.Errorf("no such file")
	}
}

func kinds(rep findingCodeReport) []string {
	out := make([]string, 0, len(rep.Findings))
	for _, f := range rep.Findings {
		out = append(out, f.Kind+"/"+f.Code)
	}
	return out
}

// ---------------------------------------------------------------------------
// The ratchet, proved in both directions
// ---------------------------------------------------------------------------

func TestUndeclaredObservedCodeIsAFinding(t *testing.T) {
	reg := map[string]findingCodeEntry{"KNOWN_CODE": {Disposition: "operational"}}

	got := auditFindingCodes([]string{"KNOWN_CODE", "TEST_UNREGISTERED_X"}, reg, srcContaining(nil), testNow)
	if len(got.Findings) != 1 || got.Findings[0].Kind != "undeclared" ||
		got.Findings[0].Code != "TEST_UNREGISTERED_X" {
		t.Fatalf("an observed code absent from the registry must be THE finding; got %v", kinds(got))
	}

	// THE CONTROL. Without this the test above passes for a check that flags
	// everything, which is the same uselessness in the other direction.
	reg["TEST_UNREGISTERED_X"] = findingCodeEntry{Disposition: "unruled"}
	clean := auditFindingCodes([]string{"KNOWN_CODE", "TEST_UNREGISTERED_X"}, reg, srcContaining(nil), testNow)
	if len(clean.Findings) != 0 {
		t.Fatalf("declaring the code must clear the finding — otherwise the check cannot come out "+
			"clean and is not discriminating; got %v", kinds(clean))
	}
	if len(clean.Unruled) != 1 || clean.Unruled[0] != "TEST_UNREGISTERED_X" {
		t.Errorf("an unruled code must be COUNTED, not silently accepted; got %v", clean.Unruled)
	}
}

// An unruled entry is a visible backlog, never an exemption — and never a
// finding, because a check that fails from day one over a pre-existing backlog
// is a check that gets ignored.
func TestUnruledIsCountedButNotAFinding(t *testing.T) {
	reg := map[string]findingCodeEntry{
		"A_CODE": {Disposition: "unruled"},
		"B_CODE": {Disposition: "unruled"},
	}
	got := auditFindingCodes([]string{"A_CODE", "B_CODE"}, reg, srcContaining(nil), testNow)
	if len(got.Findings) != 0 {
		t.Fatalf("unruled must not be a finding; got %v", kinds(got))
	}
	if len(got.Unruled) != 2 {
		t.Fatalf("unruled must be counted so the backlog is visible; got %v", got.Unruled)
	}
}

// ---------------------------------------------------------------------------
// The `consumed` reader claim is VERIFIED, not trusted
// ---------------------------------------------------------------------------
//
// This is the property that makes the field unsatisfiable by typing — the
// lesson optional_explicit_wire_acks.json paid for (RFC_029 §10.15: "an ack
// satisfiable by typing the key is no ack"). A reader reference pointed at the
// wrong file is WORSE than none, because it reads as a closed loop.

func TestConsumedEntryNeedsAReaderThatActuallyNamesTheCode(t *testing.T) {
	entry := findingCodeEntry{Disposition: "consumed", Reader: "some/reader.go:12",
		ReaderSink: "agent_error_log"}
	reg := map[string]findingCodeEntry{"MY_CODE": entry}

	wrong := auditFindingCodes([]string{"MY_CODE"}, reg,
		srcContaining(map[string]string{"some/reader.go:12": "func unrelated() { db.Query(`FROM agent_error_log`) }"}), testNow)
	if len(wrong.Findings) != 1 || wrong.Findings[0].Kind != "reader-does-not-name-code" {
		t.Fatalf("a reader that never mentions the code must be rejected; got %v", kinds(wrong))
	}

	// THE CONTROL — the same entry against a body that DOES name it.
	right := auditFindingCodes([]string{"MY_CODE"}, reg,
		srcContaining(map[string]string{
			"some/reader.go:12": "rows, _ := db.Query(`SELECT context FROM agent_error_log WHERE error_code = 'MY_CODE'`)",
		}), testNow)
	if len(right.Findings) != 0 {
		t.Fatalf("a real reader must pass, or the check rejects every entry equally; got %v", kinds(right))
	}
}

// The reader may be reached through a Go CONSTANT rather than a literal — the
// trap 358 §3.2 records, which made the bug file's own first census verdict
// DEPLOY_STAMP_REFUSED_ON_SKIP unread when its reader was right there binding a
// const to $1. Verifying at file granularity is what keeps that case passing.
func TestConsumedReaderMayReachTheCodeThroughAConstant(t *testing.T) {
	reg := map[string]findingCodeEntry{
		"DEPLOY_STAMP_REFUSED_ON_SKIP": {Disposition: "consumed", Reader: "guard.go:131",
			ReaderSink: "agent_error_log"},
	}
	body := "const deployStampRefusedErrorCode = \"DEPLOY_STAMP_REFUSED_ON_SKIP\"\n" +
		"db.QueryRow(`SELECT count(*) FROM agent_error_log WHERE error_code = $1`, deployStampRefusedErrorCode)"
	got := auditFindingCodes([]string{"DEPLOY_STAMP_REFUSED_ON_SKIP"}, reg,
		srcContaining(map[string]string{"guard.go:131": body}), testNow)
	if len(got.Findings) != 0 {
		t.Fatalf("a reader binding the code through a const must count as a reader; got %v", kinds(got))
	}
}

func TestConsumedWithoutAnyReaderIsAFinding(t *testing.T) {
	reg := map[string]findingCodeEntry{"MY_CODE": {Disposition: "consumed"}}
	got := auditFindingCodes([]string{"MY_CODE"}, reg, srcContaining(nil), testNow)
	if len(got.Findings) != 1 || got.Findings[0].Kind != "consumed-without-reader" {
		t.Fatalf("claiming 'consumed' with no reader is the exact dishonesty this guards; got %v", kinds(got))
	}
}

// ---------------------------------------------------------------------------
// `instrumented` is TIME-BOXED — the date is what stops it being a permanent
// exemption dressed up as a decision
// ---------------------------------------------------------------------------

func TestInstrumentedExpires(t *testing.T) {
	reg := map[string]findingCodeEntry{
		"RESOLVER_X": {Disposition: "instrumented", Owner: "RFC_029.md", ReviewBy: "2026-08-01"},
	}
	expired := auditFindingCodes([]string{"RESOLVER_X"}, reg, srcContaining(nil), testNow)
	if len(expired.Findings) != 1 || expired.Findings[0].Kind != "instrumentation-expired" {
		t.Fatalf("a past review_by must fire; got %v", kinds(expired))
	}

	// THE CONTROL — a future date, everything else identical.
	reg["RESOLVER_X"] = findingCodeEntry{Disposition: "instrumented", Owner: "RFC_029.md", ReviewBy: "2026-12-01"}
	live := auditFindingCodes([]string{"RESOLVER_X"}, reg, srcContaining(nil), testNow)
	if len(live.Findings) != 0 {
		t.Fatalf("an in-date instrument must pass; got %v", kinds(live))
	}
}

func TestInstrumentedNeedsAnOwnerAndADate(t *testing.T) {
	for _, tc := range []struct {
		name  string
		entry findingCodeEntry
		want  string
	}{
		{"no owner", findingCodeEntry{Disposition: "instrumented", ReviewBy: "2026-12-01"},
			"instrumented-without-owner"},
		{"no date", findingCodeEntry{Disposition: "instrumented", Owner: "RFC_029.md"},
			"instrumented-without-review-date"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := auditFindingCodes([]string{"C"},
				map[string]findingCodeEntry{"C": tc.entry}, srcContaining(nil), testNow)
			if len(got.Findings) != 1 || got.Findings[0].Kind != tc.want {
				t.Fatalf("want %s; got %v", tc.want, kinds(got))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// `human-evidence` must ACCEPT the window, not merely decline to decide
// ---------------------------------------------------------------------------

func TestHumanEvidenceMustNameTheRetentionWindow(t *testing.T) {
	vague := auditFindingCodes([]string{"C"}, map[string]findingCodeEntry{
		"C": {Disposition: "human-evidence", Why: "we look at these by hand sometimes"},
	}, srcContaining(nil), testNow)
	if len(vague.Findings) != 1 || vague.Findings[0].Kind != "human-evidence-without-window" {
		t.Fatalf("a reason that never mentions the window has accepted nothing; got %v", kinds(vague))
	}

	// THE CONTROL.
	ok := auditFindingCodes([]string{"C"}, map[string]findingCodeEntry{
		"C": {Disposition: "human-evidence",
			Why: "forensics only; rows expire at 30 days unresolved and that is accepted"},
	}, srcContaining(nil), testNow)
	if len(ok.Findings) != 0 {
		t.Fatalf("naming the window must satisfy it; got %v", kinds(ok))
	}
}

// ---------------------------------------------------------------------------
// Normalisation, decided rather than accidental (358 §8)
// ---------------------------------------------------------------------------

func TestColonVariantsCollapseToOneFamilyKey(t *testing.T) {
	reg := map[string]findingCodeEntry{"tool_crosslink_not_emitted": {Disposition: "unruled"}}
	got := auditFindingCodes([]string{
		"tool_crosslink_not_emitted:tool_page_will_not_go_live",
		"tool_crosslink_not_emitted:no_related_pages",
		"tool_crosslink_not_emitted:a_reason_invented_tomorrow",
	}, reg, srcContaining(nil), testNow)

	if len(got.Findings) != 0 {
		t.Fatalf("the raw variants must resolve to the registered family key; got %v", kinds(got))
	}
	// AND the family must count ONCE. Counting three would let a family
	// double-count as compliance, which is the trap 358 §8 names by name.
	if got.ObservedCount != 1 {
		t.Fatalf("three raw variants of one family are ONE observed code; got %d", got.ObservedCount)
	}
}

// ---------------------------------------------------------------------------
// Registry-wide properties — these subsume the two hand-maintained `taken`
// lists that had gone stale in the actions package
// ---------------------------------------------------------------------------
//
// Prefix-disjointness is a REAL property and its justification is measured, not
// stylistic (save_sections_content_data_links_test.go:148): the estate has live
// LIKE queries on `tool_crosslink_not_emitted%` and `component_validation_%`, so
// a code sharing a prefix with another is silently swept into someone else's
// population.

func TestPrefixCollisionIsAFinding(t *testing.T) {
	got := auditFindingCodes([]string{"CONTENT_LINK"}, map[string]findingCodeEntry{
		"CONTENT_LINK":        {Disposition: "unruled"},
		"CONTENT_LINK_REPAIR": {Disposition: "unruled"},
	}, srcContaining(nil), testNow)
	found := false
	for _, f := range got.Findings {
		if f.Kind == "prefix-collision" {
			found = true
		}
	}
	if !found {
		t.Fatalf("a code that is a prefix of another must fire; got %v", kinds(got))
	}

	// THE CONTROL — codes that merely share an early segment are fine, and must
	// not fire, or the rule would forbid the estate's whole naming convention.
	clean := auditFindingCodes([]string{"CONTENT_LINK_REPAIR_DETAIL"}, map[string]findingCodeEntry{
		"CONTENT_LINK_REPAIR_DETAIL":  {Disposition: "unruled"},
		"CONTENT_LINK_REPAIR_SKIPPED": {Disposition: "unruled"},
	}, srcContaining(nil), testNow)
	for _, f := range clean.Findings {
		if f.Kind == "prefix-collision" {
			t.Fatalf("CONTENT_LINK_REPAIR_DETAIL and _SKIPPED share a stem but neither is a "+
				"prefix of the other — firing here would ban the naming convention; got %v", kinds(clean))
		}
	}
}

func TestBadDispositionIsAFinding(t *testing.T) {
	got := auditFindingCodes([]string{"C"}, map[string]findingCodeEntry{
		"C": {Disposition: "probably-fine"},
	}, srcContaining(nil), testNow)
	if len(got.Findings) != 1 || got.Findings[0].Kind != "bad-disposition" {
		t.Fatalf("an invented disposition must not pass as a declaration; got %v", kinds(got))
	}
}

// ---------------------------------------------------------------------------
// The asymmetry between the two directions
// ---------------------------------------------------------------------------
//
// Retention is 30 days, so an unobserved code proves 30 days of silence and
// never "never" (358 §8). It must be REPORTED — BUILD_DISPATCH_STALLED sits
// here because migration 214 was never applied, so a loop the bug file lists as
// closed does not exist live — but it must never be a finding, because a quiet
// code and a dead mechanism are indistinguishable from this side.

func TestRegisteredButUnobservedIsReportedNotAFinding(t *testing.T) {
	got := auditFindingCodes([]string{"SEEN"}, map[string]findingCodeEntry{
		"SEEN":   {Disposition: "operational"},
		"UNSEEN": {Disposition: "operational"},
	}, srcContaining(nil), testNow)
	if len(got.Findings) != 0 {
		t.Fatalf("an unobserved code must not fail the check — 30 days is not 'never'; got %v", kinds(got))
	}
	if len(got.NotObserved) != 1 || got.NotObserved[0] != "UNSEEN" {
		t.Fatalf("it must still be REPORTED, or a dead mechanism is invisible; got %v", got.NotObserved)
	}
}

// ---------------------------------------------------------------------------
// The shipped registry must satisfy its own rules
// ---------------------------------------------------------------------------
//
// Everything above tests the checker against fixtures. This tests the FILE the
// estate actually ships, which is the one that can rot. It deliberately does
// NOT assert a code count: the live population changes daily and a hard number
// here would be a second stale roster, which is the class this whole change
// retires.
func TestShippedRegistryIsSelfConsistent(t *testing.T) {
	reg, err := loadFindingCodeRegistry("../../" + findingCodeRegistryPath)
	if err != nil {
		t.Fatalf("the shipped registry must load: %v", err)
	}

	// Feed it its own codes so every entry is exercised, then assert the only
	// findings possible are ones about the entries themselves — never
	// "undeclared", which cannot arise when the input IS the registry.
	live := make([]string, 0, len(reg))
	for c := range reg {
		live = append(live, c)
	}
	got := auditFindingCodes(live, reg, repoSourceReader("../.."), testNow)
	if len(got.Findings) != 0 {
		var b strings.Builder
		for _, f := range got.Findings {
			fmt.Fprintf(&b, "\n  [%s] %s — %s", f.Kind, f.Code, f.Detail)
		}
		t.Fatalf("the shipped registry does not satisfy its own rules:%s", b.String())
	}
}

// ─── reader_sink (batch 1, 2026-08-23) ──────────────────────────────────────
//
// `consumed` was silently ambiguous between "an automated reader consumes this
// code" and "this table's row is read". component_validation_rejected is the
// live case where those differ: its reader is real, and it reads
// site_work_items.retry_feedback, not agent_error_log.

func TestConsumedWithoutAReaderSinkIsAFinding(t *testing.T) {
	reg := map[string]findingCodeEntry{
		"MY_CODE": {Disposition: "consumed", Reader: "r.go:1"}, // no ReaderSink
	}
	got := auditFindingCodes([]string{"MY_CODE"}, reg,
		srcContaining(map[string]string{"r.go:1": "FROM agent_error_log WHERE error_code='MY_CODE'"}), testNow)
	if len(got.Findings) != 1 || got.Findings[0].Kind != "consumed-without-reader-sink" {
		t.Fatalf("consumed must name the sink it reads from; got %v", kinds(got))
	}
}

// The sink claim is VERIFIED against the reader body, exactly as the code claim
// is. This is the check that would have caught the motivating case: 563's
// prompt template never mentions agent_error_log at all.
func TestReaderSinkMustAppearInTheReader(t *testing.T) {
	reg := map[string]findingCodeEntry{
		"MY_CODE": {Disposition: "consumed", Reader: "r.go:1", ReaderSink: "agent_error_log"},
	}
	got := auditFindingCodes([]string{"MY_CODE"}, reg,
		srcContaining(map[string]string{
			"r.go:1": "SELECT retry_feedback FROM site_work_items WHERE code='MY_CODE'",
		}), testNow)
	if len(got.Findings) != 1 || got.Findings[0].Kind != "reader-sink-not-in-reader" {
		t.Fatalf("a sink the reader never mentions must be rejected; got %v", kinds(got))
	}
}

// A FOREIGN sink is REPORTED, never failed. A parallel append-only record
// beside an overwritten column is often deliberate — what must not happen is
// the entry reading as a closed loop over agent_error_log when it is not one.
func TestForeignSinkIsReportedNotFailed(t *testing.T) {
	reg := map[string]findingCodeEntry{
		"component_validation_rejected": {
			Disposition: "consumed",
			Reader:      "563.sql:156",
			ReaderSink:  "site_work_items.retry_feedback",
		},
	}
	got := auditFindingCodes([]string{"component_validation_rejected"}, reg,
		srcContaining(map[string]string{
			"563.sql:156": "eq .last_error_code \"component_validation_rejected\" -- fed from site_work_items.retry_feedback",
		}), testNow)
	if len(got.Findings) != 0 {
		t.Fatalf("a foreign sink must not FAIL the check; got %v", kinds(got))
	}
	if len(got.ForeignSinks) != 1 {
		t.Fatalf("a foreign sink must be REPORTED; got %v", got.ForeignSinks)
	}
	if !strings.Contains(got.ForeignSinks[0], "still unread") {
		t.Errorf("the report must say the agent_error_log row is unread; got %q", got.ForeignSinks[0])
	}
}

// THE CONTROL for the test above: agent_error_log itself must NOT be reported,
// or the report would name every consumed entry and mean nothing.
func TestOwnSinkIsNotReported(t *testing.T) {
	reg := map[string]findingCodeEntry{
		"MY_CODE": {Disposition: "consumed", Reader: "r.go:1", ReaderSink: "agent_error_log"},
	}
	got := auditFindingCodes([]string{"MY_CODE"}, reg,
		srcContaining(map[string]string{"r.go:1": "FROM agent_error_log WHERE error_code='MY_CODE'"}), testNow)
	if len(got.Findings) != 0 || len(got.ForeignSinks) != 0 {
		t.Fatalf("the ordinary case must be silent on both channels; got %v / %v", kinds(got), got.ForeignSinks)
	}
}
