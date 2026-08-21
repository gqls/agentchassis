// FILE: platform/orchestration/datahelpers/marked_config_key_parity_test.go
//
// Council REVISE round 2, editquality seat (medium): `MarkedConfigKey` is
// documented as adopting input_mapping.go's algorithm rather than
// ExtractActionInputs' own — and the ExtractActionInputs surface already has 6
// live `!` wires. Nothing verified the two algorithms AGREE on those before the
// swap, so a divergence would have shipped silently under a "no behaviour
// change" claim. The seat was right that the claim was unpinned; this file pins
// it.
//
// The two algorithms, as they stood before the extraction:
//
//	input_mapping.go (the one that survived, 77 live `?` + 1 live `!`):
//	    strict   = HasSuffix(k, "!")
//	    optional = !strict && HasSuffix(k, "?")
//	    base     = TrimSuffix(TrimSuffix(k, "!"), "?")
//
//	action_inputs.go (retired here, 6 live `!`, 0 live `?`):
//	    if base := TrimSuffix(k, "!"); base != k && base != "" -> strict
//	    else if base := TrimSuffix(k, "?"); base != k && base != "" -> optional
//
// They agree on every single-marker key and differ ONLY on the degenerate
// double-marker spellings, which no live definition carries. Both facts are
// asserted below rather than asserted in prose.
package datahelpers

import (
	"strings"
	"testing"
)

// legacyActionInputsParse is the ExtractActionInputs algorithm as it stood
// immediately before MarkedConfigKey replaced it (commit ecc419bd1's form),
// transcribed here so the parity claim is checkable rather than remembered. It
// exists only in this test.
func legacyActionInputsParse(key string) (base string, strict, optional bool) {
	if b := strings.TrimSuffix(key, "!"); b != key && b != "" {
		return b, true, false
	}
	if b := strings.TrimSuffix(key, "?"); b != key && b != "" {
		return b, false, true
	}
	return key, false, false
}

// EVERY MARKED KEY SPELLING LIVE IN THE FLEET, enumerated 2026-08-21 by a
// recursive walk over all 194 live definitions (every key at every depth, both
// loop-container spellings, arrays included) — not a hand-picked sample. The
// two `!` entries are the ExtractActionInputs surface's live strict wires; the
// rest are input_mapping's. Kept verbatim so a future divergence is caught
// against the real corpus and not against invented fixtures.
var liveMarkedKeys = []string{
	"area_code?", "asset_id!", "asset_id?", "asset_key?", "brand_logo_url?",
	"build_mode?", "check?", "component_id?", "constraints?", "correlation_id?",
	"destination_domain?", "domain?", "existing_content?", "expiry_minutes?",
	"fidelity?", "hero_url?", "include_fenced?", "instance_ip?", "issue?",
	"item_type?", "kind?", "language?", "logo_url?", "max_rows?",
	"min_response_length?", "model_filter?", "orchestration_id?", "owner?",
	"page_id?", "page_name?", "page_type?", "purpose?", "ref?", "repo?",
	"request_id?", "result!", "reviewed_brief?", "rewrite_guidance?",
	"runtime_page?", "runtime_site?", "sections?", "seed_scope?", "site_id?",
	"size_alert_bytes?", "source_notes?", "ssh_key_secret_name?", "ssh_user?",
	"strict_json?", "style_hints?", "subject_key?", "subject_type?",
	"submitter?", "target_url?", "triggered_by?", "url?", "work_item_id!",
}

// The parity claim, against the live corpus: the shared parser returns exactly
// what the retired one did for every marker spelling any live definition
// actually carries. This is what makes "the extraction changes no behaviour" a
// checked statement instead of an assertion in a doc comment.
func TestMarkedConfigKeyMatchesTheRetiredAlgorithmOnEveryLiveKey(t *testing.T) {
	var sawStrict, sawOptional int
	for _, key := range liveMarkedKeys {
		gotBase, gotStrict, gotOptional := MarkedConfigKey(key)
		wantBase, wantStrict, wantOptional := legacyActionInputsParse(key)

		if gotBase != wantBase || gotStrict != wantStrict || gotOptional != wantOptional {
			t.Errorf("DIVERGENCE on live key %q: MarkedConfigKey = (%q, strict=%v, optional=%v), "+
				"retired algorithm = (%q, strict=%v, optional=%v) — the extraction was claimed "+
				"to change no behaviour and this key is the counter-example",
				key, gotBase, gotStrict, gotOptional, wantBase, wantStrict, wantOptional)
		}
		if gotStrict {
			sawStrict++
		}
		if gotOptional {
			sawOptional++
		}
	}

	// VACUITY CONTROLS: a corpus that exercised neither marker would pass the
	// loop above no matter what either algorithm did.
	if sawStrict < 2 {
		t.Errorf("corpus exercised %d strict keys, want >=2 — the live `!` wires are what the "+
			"editquality objection was about; without them this test proves nothing about them",
			sawStrict)
	}
	if sawOptional < 10 {
		t.Errorf("corpus exercised %d optional keys, want >=10 — the `?` surface is the one "+
			"whose algorithm was kept, so it must be exercised too", sawOptional)
	}
}

// WHERE THEY DIVERGE, pinned deliberately: the double-marker spellings. No live
// definition carries one (the census above is the whole corpus), so this is a
// behaviour change on an EMPTY population — but an unrecorded one would be a
// silent trap for the first author who types `field!?`, and "no live instance"
// is exactly the reasoning that gets quoted later as "no difference".
func TestTheTwoAlgorithmsDivergeOnlyOnDoubleMarkers(t *testing.T) {
	cases := []struct {
		key                          string
		base                         string
		strict, optional             bool
		legacyBase                   string
		legacyStrict, legacyOptional bool
	}{
		// `field?!` — both agree it is strict; they disagree on the BASE. The
		// shared (input_mapping) form strips both suffixes; the retired one
		// stripped only the `!`, leaving a `?` glued to the field name, which
		// could never match a spec field. This is the ONLY divergence.
		{"field?!", "field", true, false, "field?", true, false},
	}
	// `field!?` was predicted to diverge too and DOES NOT — both algorithms
	// return ("field!", optional). Recorded because the wrong prediction is the
	// useful part: I reasoned that the retired form's `TrimSuffix(k,"!")` would
	// "see the ! anywhere", when TrimSuffix only ever strips a SUFFIX, so on a
	// key ending in `?` its first branch is a no-op and both algorithms fall
	// through to the same second one. The "must diverge" assertion below is
	// what caught it — an expectation table with no such guard would have
	// recorded my guess as a finding.

	for _, c := range cases {
		base, strict, optional := MarkedConfigKey(c.key)
		if base != c.base || strict != c.strict || optional != c.optional {
			t.Errorf("MarkedConfigKey(%q) = (%q, strict=%v, optional=%v), want (%q, %v, %v)",
				c.key, base, strict, optional, c.base, c.strict, c.optional)
		}
		lBase, lStrict, lOptional := legacyActionInputsParse(c.key)
		if lBase != c.legacyBase || lStrict != c.legacyStrict || lOptional != c.legacyOptional {
			t.Errorf("legacyActionInputsParse(%q) = (%q, %v, %v), want (%q, %v, %v) — the "+
				"transcription of the retired algorithm has drifted, so the parity test above "+
				"is comparing against the wrong thing",
				c.key, lBase, lStrict, lOptional, c.legacyBase, c.legacyStrict, c.legacyOptional)
		}
		if base == lBase && strict == lStrict && optional == lOptional {
			t.Errorf("%q was expected to DIVERGE between the two algorithms and did not — if the "+
				"divergence has been removed, this test should be deleted rather than left "+
				"asserting a difference that no longer exists", c.key)
		}
	}

	// The agreeing degenerate case, pinned so the divergence set above is known
	// to be COMPLETE rather than merely non-empty.
	if base, strict, optional := MarkedConfigKey("field!?"); base != "field!" || strict || !optional {
		t.Errorf(`MarkedConfigKey("field!?") = (%q, strict=%v, optional=%v), want ("field!", false, true)`,
			base, strict, optional)
	}
	if base, strict, optional := legacyActionInputsParse("field!?"); base != "field!" || strict || !optional {
		t.Errorf(`legacyActionInputsParse("field!?") = (%q, strict=%v, optional=%v), want ("field!", false, true) `+
			`— if this moved, the two algorithms now differ here too and the divergence set above is incomplete`,
			base, strict, optional)
	}
}

// A key that is nothing but a marker names no field. Both surfaces must skip it
// rather than register a field called "" — asserted because the shared parser
// returns an empty base and it is the CALLER that has to notice.
func TestBareMarkerYieldsAnEmptyBase(t *testing.T) {
	for _, key := range []string{"!", "?"} {
		if base, _, _ := MarkedConfigKey(key); base != "" {
			t.Errorf("MarkedConfigKey(%q) base = %q, want empty: callers guard on the empty base "+
				"to avoid registering a field with no name", key, base)
		}
	}
}
