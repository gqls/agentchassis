// FILE: platform/orchestration/actions/query_row_cap.go
//
// A ROW-COUNT `LIMIT` FEEDING AN LLM PROMPT IS A SILENT CAP (bugs_open/275).
//
// `tool-suggester`'s `load_library_tools` ended `ORDER BY display_name LIMIT 30`
// against a library of 74 masters, so the model choosing which tools a site
// should have was shown **30 of 74** — and which 44 were hidden was an accident
// of the alphabet, not a judgement. Nothing ever looked wrong, because an LLM
// returns plausible suggestions whether it saw 30 rows or 74. That is what makes
// it a *silent* cap rather than a small one: there is no output to inspect.
//
// It is not one query. Censused 2026-08-16 across live `agent_definitions`,
// there are 26 literal `LIMIT`s in query-shaped step configs, of which seven are
// multi-row caps that can bite this way (30, 15, 12, 10, 5, 5, 2). Every one of
// them runs through `QueryDatabaseAction`, which is why the detection belongs
// here and not in any single agent's config: one function, and the whole class
// becomes visible at once, including caps nobody has written yet.
//
// DELIBERATELY OBSERVATIONAL. This file cannot change a query, a result, or a
// row. It emits a log line. That is the entire scope, and it is chosen: this is
// a shared action serving 26+ live steps, and the owner ruling of 2026-08-02 §2
// (RFC_010) says new authority on a shared seam must arrive opt-in with the
// unsafe default OFF. A warning takes no authority, so there is nothing to
// switch off.
//
// WHY EQUALITY AND NOT A REAL COUNT. "30 of 74" would need either a second
// COUNT(*) per step — doubling database work fleet-wide — or rewriting the SQL
// to `LIMIT n+1` and trimming the extra row, which changes what a shared action
// RETURNS and is a far larger blast radius than this warrants. Equality is a
// cheap and honest *suspicion*: when a result set is exactly as large as its own
// ceiling, it may have been truncated, and naming the step lets a human settle it
// with one query. The `n+1` probe is the recorded follow-up, not smuggled in here.
package actions

import (
	"regexp"
	"strconv"
)

// queryLimitRe matches a trailing literal LIMIT. Anchored to the END of the
// statement (allowing trailing whitespace and a semicolon) because that is the
// only position where the number bounds the whole result set — a LIMIT inside a
// subquery or a CTE bounds something else, and warning on it would be wrong.
//
// A parameterised limit (`LIMIT $2`) deliberately does not match: there is no
// literal to compare against, and guessing would produce a warning nobody can
// act on.
// A trailing SQL comment is tolerated. Without this, `... LIMIT 30 -- widened 08-14`
// is a FALSE NEGATIVE — a genuinely capped query going undetected by the very
// mechanism built to end silent caps, which is the worst possible place for a
// blind spot. Raised by the council's editquality seat (corr b684a399 round 2,
// advisory low) and worth taking: an operator annotating a cap is exactly the
// operator whose cap most deserves a second look.
var queryLimitRe = regexp.MustCompile(`(?is)\bLIMIT\s+(\d+)\s*;?\s*(?:(?:--[^\n]*|/\*.*?\*/)\s*)*$`)

// queryRowCap returns the statement's trailing literal LIMIT and whether one was
// found.
//
// ⚠ `LIMIT 1` IS EXCLUDED, and this is the signal-to-noise decision the whole
// check rests on. 19 of the 26 live hits are `LIMIT 1` — the fetch-one and
// claim-one idiom (`load_brief`, `claim_item`, `load_tool`, `load_diagnosis`, …).
// Those return exactly one row *by design*, on every single execution, so
// including them would emit a warning per run for ever and train every reader to
// ignore the channel. A cap that "bites" at one row is not a truncation, it is a
// lookup. `LIMIT 0` is likewise not a cap worth reporting.
//
// The cost of the exclusion is stated rather than hidden: a genuine one-row
// truncation is invisible to this check. That is accepted — the class this exists
// for is "the model was shown a fraction of a corpus", and a corpus of one is not
// that.
func queryRowCap(query string) (int, bool) {
	m := queryLimitRe.FindStringSubmatch(query)
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n < 2 {
		return 0, false
	}
	return n, true
}

// resultHitItsRowCap reports whether a result set is exactly as large as its own
// declared ceiling — i.e. the query may have returned a truncated view and the
// caller cannot tell.
//
// KNOWN FALSE POSITIVE, stated here rather than discovered later: a population
// that happens to equal the cap exactly will warn while nothing is actually
// hidden. That is the correct trade for an observational signal. A false "go and
// check this" costs one query; a false silence cost 44 of 74 tools for as long as
// nobody looked.
func resultHitItsRowCap(query string, rowCount int) (limit int, hit bool) {
	n, ok := queryRowCap(query)
	if !ok {
		return 0, false
	}
	// `>=`, not `==`, and the difference is deliberate rather than incidental.
	//
	// A database cannot return more rows than the statement's own LIMIT, so for
	// every input reachable from QueryDatabaseAction the two are identical —
	// mutation-tested 2026-08-16, and `==` -> `>=` SURVIVED, which is what an
	// equivalent mutant looks like. The reason to prefer `>=` anyway is the
	// SECOND caller this helper does not have yet: anyone passing a row count
	// from a different source (a merged set, a retried page, a hand-built slice)
	// would silently lose the signal under `==` at exactly the moment the count
	// stopped being the raw driver result. `>=` degrades to the same answer and
	// cannot go quiet.
	return n, rowCount >= n
}
