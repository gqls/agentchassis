// FILE: platform/orchestration/datahelpers/jsonb_nul.go
//
// Shared jsonb NUL-escape sanitiser (bugs_open/056; council-approved as its own
// change, corr d8e844ac-309f-4576-9898-77034f871fbc, 2026-07-22). Postgres
// jsonb rejects exactly ONE Unicode escape —   (SQLSTATE 22P05) — and
// json.Marshal emits it (always lower-case) for any NUL byte in a string or
// map key, so a single stray NUL anywhere in a marshalled value fails the
// whole INSERT/UPDATE it rides in. Pre-fix that meant the run died: no code
// anywhere handles 22P05, and jsonb structurally cannot store the byte, so a
// hard failure preserves nothing. The policy (reviewed on its own terms per
// the council trail above) is: substitute U+FFFD, succeed, and let the CALLER
// make the substitution loud (see UpdateStateWithVersion's WARN).
//
// Lives here, not in state.go, so other jsonb writers (diagnosis_artifacts,
// site_work_items, agent_definitions persist paths) can reuse it instead of
// re-deriving it.
package datahelpers

import "bytes"

var (
	jsonbNulEscape  = []byte(`\u0000`)
	jsonbNulReplace = []byte(`\ufffd`) // U+FFFD keeps composite keys distinct where stripping would collapse a delimiter
)

// SanitiseJSONBNulEscapes replaces every GENUINE \u0000 escape in marshalled
// JSON with the replacement character's escape, so the value survives a jsonb
// column, and reports how many it replaced so the caller can log the event.
// Literal backslash-u0000 TEXT (an escaped backslash followed by "u0000" —
// e.g. a diagnosis quoting this very escape from a doc) is left untouched: a
// backslash starts an escape only when preceded by an even number of
// backslashes. Zero-allocation on the overwhelmingly common no-match path.
func SanitiseJSONBNulEscapes(b []byte) ([]byte, int) {
	if !bytes.Contains(b, jsonbNulEscape) {
		return b, 0
	}
	out := make([]byte, 0, len(b))
	replaced := 0
	for i := 0; i < len(b); {
		j := bytes.Index(b[i:], jsonbNulEscape)
		if j < 0 {
			out = append(out, b[i:]...)
			break
		}
		j += i
		preceding := 0
		for k := j - 1; k >= 0 && b[k] == '\\'; k-- {
			preceding++
		}
		out = append(out, b[i:j]...)
		if preceding%2 == 0 {
			out = append(out, jsonbNulReplace...)
			replaced++
		} else {
			out = append(out, jsonbNulEscape...)
		}
		i = j + len(jsonbNulEscape)
	}
	return out, replaced
}
