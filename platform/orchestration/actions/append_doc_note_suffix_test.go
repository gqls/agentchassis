package actions

import (
	"strings"
	"testing"
)

// bugs_open/223 — the landmine-verifier's product is a doc_notes row of PROSE,
// read months later by sessions and by council seats, and nothing parses its
// status. So when a verdict rests on checks the code index could not answer, the
// only way a later reader can tell is if the row itself says so.
//
// Asking the model to include that caveat is not a fix: on identical 0-row input,
// three of four recorded runs already hedged correctly and the entry was degraded
// by the fourth. A suffix composed in Go cannot be softened, contradicted or
// dropped by the model whose verdict it qualifies — that is the whole of why this
// is a function.
func TestApplyBodySuffix(t *testing.T) {
	const verdict = "**last verified (landmine-verifier): STALE.** None of the three scripts exist."
	const evidence = "[code-lookup evidence: 8 check(s) ran; 0 matched indexed code; 5 NOT ANSWERABLE by this index.]"

	// OPT-IN: no field configured ⇒ byte-identical to before this change existed.
	// This is the assertion that lets the key ship to a shared action without
	// touching any of its other consumers.
	if got := applyBodySuffix(verdict, evidence, ""); got != verdict {
		t.Errorf("with no field configured the body must be untouched, got:\n%s", got)
	}

	// Configured and resolved: the qualifier lands in the persisted row.
	got := applyBodySuffix(verdict, evidence, "lookup.evidence_line")
	if !strings.HasPrefix(got, verdict) {
		t.Error("the model's verdict must survive verbatim — this qualifies it, it does not replace it")
	}
	if !strings.Contains(got, evidence) {
		t.Errorf("the mechanical qualifier must be present, got:\n%s", got)
	}

	// Configured and EMPTY: stated in the row, not silently skipped. A missing
	// qualifier is precisely the condition the key exists to make visible, so it
	// must be loud in the artefact and not only in a log nobody reads. (The
	// estate's own recorded failure: a guard that fails silent reads as a guard
	// that passed.)
	empty := applyBodySuffix(verdict, "", "lookup.evidence_line")
	if !strings.Contains(empty, "resolved EMPTY") || !strings.Contains(empty, "lookup.evidence_line") {
		t.Errorf("an empty resolution must say so, naming the field, got:\n%s", empty)
	}
	if !strings.Contains(empty, "treat this note as unqualified") {
		t.Errorf("the empty case must tell the reader what to do about it, got:\n%s", empty)
	}
}
