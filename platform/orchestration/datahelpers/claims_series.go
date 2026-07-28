// Series facts — an evidence fact that holds MANY dated observations rather
// than one value.
//
// WHY THIS EXISTS. An EvidenceFact carries a single Value plus three dates, and
// none of them is the date the value applies to:
//
//	accessed     when we fetched the source
//	published    when the source was issued
//	verified_at  when we last checked the fact still holds
//
// All three are PROVENANCE. Asked for "Thames Water net debt at each year end"
// or "Ofwat determinations over time", the register has no shape for it. The
// options without this type are all bad: one fact per year (no series identity,
// so nothing can plot them and nothing knows one is missing), or a single fact
// whose claim string embeds the numbers (invisible to every scanner, which is
// how invented figures survive).
//
// The rule this preserves is the one the whole evidence layer exists for: a
// figure reaches a page only by being registered with a source. A chart is the
// worst place to relax it, because a plotted point carries no sentence a reader
// can interrogate and no scanner can read a bar's height. So:
//
//   - EVERY observation carries its OWN source. There is deliberately no
//     inheritance from the parent fact. A series where the first point is cited
//     and the rest are "continued from the same source" is exactly the artefact
//     this must make impossible — interpolation and extrapolation both arrive
//     looking like data, and a shared parent source is how they get in.
//   - as_of is REQUIRED and is distinct from verified_at. Conflating them makes
//     a stale series look freshly checked: re-verifying a 2021 figure in 2026
//     updates verified_at and must not move the point on the x-axis.
//
// Nothing here renders. Rendering resolves values through fact ids the way
// `evidence-chart` already does, so a chart definition still cannot carry a
// number of its own.
package datahelpers

import (
	"fmt"
	"sort"
	"strings"
)

// KindSeries is the EvidenceFact.Kind for a fact carrying observations.
// The existing kinds (metric | capability | entity | attestation) each describe
// one assertion; a series describes the same assertion measured repeatedly.
const KindSeries = "series"

// Observation is one dated point in a series fact.
//
// Source is the raw map rather than a typed field for the same reason
// EvidenceFact keeps its citation in jsonb: the provenance of a point may be a
// citation, a SQL query against our own tables, or an artifact, and ParseCitation
// already knows how to read that union. Keeping the shape identical to a fact's
// source means an observation is verifiable by exactly the machinery that
// verifies a fact, rather than by a second path that can drift.
type Observation struct {
	AsOf       string                 `json:"as_of"`      // YYYY-MM-DD or YYYY-MM or YYYY — the date the VALUE APPLIES TO
	Value      float64                `json:"value"`      //
	Source     map[string]interface{} `json:"source"`     // REQUIRED, never inherited from the parent fact
	VerifiedAt string                 `json:"verified_at"`// when we last checked THIS point
	Note       string                 `json:"note,omitempty"`
}

// SeriesObservations returns a fact's observations in as_of order.
//
// Sorting is lexical, which is correct for the accepted formats because they are
// all zero-padded and big-endian: "2024" < "2024-03" < "2024-03-31" < "2025".
// A mixed-granularity series therefore still orders correctly. ValidateSeries
// rejects anything that is not one of those three shapes, so this cannot be
// handed a format where lexical order and chronological order disagree.
func (f *EvidenceFact) SeriesObservations() []Observation {
	if f == nil || len(f.Observations) == 0 {
		return nil
	}
	out := make([]Observation, len(f.Observations))
	copy(out, f.Observations)
	sort.SliceStable(out, func(i, j int) bool { return out[i].AsOf < out[j].AsOf })
	return out
}

// IsSeries reports whether this fact carries observations.
func (f *EvidenceFact) IsSeries() bool {
	return f != nil && (f.Kind == KindSeries || len(f.Observations) > 0)
}

// SeriesProblem is one reason a series fact is not publishable.
type SeriesProblem struct {
	Index  int    // observation index, or -1 for a whole-fact problem
	AsOf   string //
	Reason string //
}

func (p SeriesProblem) String() string {
	if p.Index < 0 {
		return p.Reason
	}
	return fmt.Sprintf("observation %d (as_of %q): %s", p.Index, p.AsOf, p.Reason)
}

// ValidateSeries returns every reason a series fact must not be published.
// An empty slice means publishable.
//
// This FAILS CLOSED and is deliberately strict: an unusable series must be
// unusable rather than degraded, because a chart silently missing its
// unverifiable points is more dangerous than one that refuses to render. The
// reader cannot see the gap, and the remaining shape still tells a story.
func (f *EvidenceFact) ValidateSeries() []SeriesProblem {
	if f == nil {
		return []SeriesProblem{{Index: -1, Reason: "nil fact"}}
	}
	if !f.IsSeries() {
		return []SeriesProblem{{Index: -1, Reason: "not a series fact"}}
	}
	var problems []SeriesProblem
	if len(f.Observations) < 2 {
		problems = append(problems, SeriesProblem{
			Index:  -1,
			Reason: "a series needs at least 2 observations; a single dated value is an ordinary metric fact",
		})
	}

	seen := make(map[string]int, len(f.Observations))
	for i, o := range f.Observations {
		if !validAsOf(o.AsOf) {
			problems = append(problems, SeriesProblem{Index: i, AsOf: o.AsOf,
				Reason: "as_of must be YYYY, YYYY-MM or YYYY-MM-DD"})
		}
		if prev, dup := seen[o.AsOf]; dup {
			problems = append(problems, SeriesProblem{Index: i, AsOf: o.AsOf,
				Reason: fmt.Sprintf("duplicate as_of, already used by observation %d", prev)})
		} else {
			seen[o.AsOf] = i
		}

		// The load-bearing check. No inheritance: an observation with no source
		// of its own is not evidence, whatever its parent fact carries. Uses the
		// SAME predicate as seriesSupports, so the validator and the gate cannot
		// disagree about what counts as sourced.
		if len(o.Source) == 0 {
			problems = append(problems, SeriesProblem{Index: i, AsOf: o.AsOf,
				Reason: "no source; an observation never inherits the parent fact's source"})
			continue
		}
		if !observationHasResolvableSource(o) {
			if _, err := ParseCitation(o.Source); err != nil {
				problems = append(problems, SeriesProblem{Index: i, AsOf: o.AsOf,
					Reason: "citation unusable: " + err.Error()})
			} else {
				problems = append(problems, SeriesProblem{Index: i, AsOf: o.AsOf,
					Reason: "source names no citation, sql or artifact"})
			}
		}
	}
	return problems
}

// observationHasResolvableSource is the single definition of "this point is
// evidence". It is deliberately shared by ValidateSeries and by seriesSupports.
//
// Council round 1 (corr da40ddf0) caught the hole this closes: ValidateSeries
// enforced the never-inherit rule, but numberSupported went straight from
// IsSeries() to seriesSupports() and never called the validator. An unsourced
// observation would therefore still have registered its value as supported, so
// an unsourced number could reach a page through a series nobody had validated
// — defeating the entire purpose of the type. A rule enforced only in the
// validator is not enforced; it has to hold at the gate that actually decides.
func observationHasResolvableSource(o Observation) bool {
	if len(o.Source) == 0 {
		return false
	}
	cit, err := ParseCitation(o.Source)
	if err != nil {
		return false // citation present but unusable (missing url/quote/publisher)
	}
	if cit != nil {
		return true
	}
	return hasNonCitationSource(o.Source)
}

// hasNonCitationSource reports whether a source map carries one of the
// non-citation provenance kinds (mirrors EvidenceSource's fields).
func hasNonCitationSource(src map[string]interface{}) bool {
	for _, k := range []string{"sql", "artifact", "attested_by"} {
		if s, ok := src[k].(string); ok && strings.TrimSpace(s) != "" {
			return true
		}
	}
	return false
}

// validAsOf accepts YYYY, YYYY-MM and YYYY-MM-DD, all zero-padded, and nothing
// else. The restriction is what lets SeriesObservations sort lexically, and it
// rejects the formats that sort wrongly ("3/2024", "March 2024").
func validAsOf(s string) bool {
	switch len(s) {
	case 4:
		return allDigits(s)
	case 7:
		return allDigits(s[:4]) && s[4] == '-' && allDigits(s[5:])
	case 10:
		return allDigits(s[:4]) && s[4] == '-' && allDigits(s[5:7]) &&
			s[7] == '-' && allDigits(s[8:])
	default:
		return false
	}
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// seriesSupports reports whether any observation in this series registers val.
//
// Context terms are applied at the SERIES level, not per observation: the terms
// scope which claim windows the series may support ("net debt", "determination"),
// and every observation is the same measurement at a different date, so a term
// that qualifies one qualifies all.
//
// Matching is exact, deliberately. The tolerances an ordinary fact may carry
// ("gte", "approx_pct") are wrong for a time series: a gte series would support
// every number below its maximum, which over a long enough series is very nearly
// every number on the page — the blanket-support failure the ContextTerms comment
// on EvidenceFact already warns about, made worse by having many values.
func (f *EvidenceFact) seriesSupports(val float64, lowerWindow string) bool {
	if f == nil || len(f.Observations) == 0 {
		return false
	}
	if len(f.ContextTerms) > 0 {
		matched := false
		for _, t := range f.ContextTerms {
			if strings.Contains(lowerWindow, strings.ToLower(t)) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	for _, o := range f.Observations {
		// An observation with no resolvable source of its own is NOT evidence and
		// must not register its value, even though it sits inside a series whose
		// other points are impeccable. Without this the type's central guarantee
		// would hold only for callers who remembered to run ValidateSeries first.
		if !observationHasResolvableSource(o) {
			continue
		}
		if o.Value == val || (o.Value-val < 1e-9 && val-o.Value < 1e-9) {
			return true
		}
	}
	return false
}
