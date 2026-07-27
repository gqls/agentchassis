// FILE: platform/orchestration/datahelpers/meta_description.go
//
// PublicMetaDescription is the one gate between a string we hold internally and
// the text a search engine prints under a result.
//
// It exists because of bugs_open/103: every deployed tool page published its
// internal BUILD BRIEF as its public meta description, verbatim. The worst live
// case was 1,206 characters of generator instructions on vonc.com's Arena page
// — "no fetch calls, no backend", "embed 5 sample provocations in JS and pick
// one by day-of-date" — telling Google the page has no backend.
//
// The mechanism was mundane and is the reason a shared helper is the right
// shape: `content_components.description` for a component_level='tool' row is
// the brief, not marketing copy, and two separate page-creating call sites bound
// it straight into pages.meta_description. Neither was wrong on its face; the
// field is a string and a brief is a string. Nothing anywhere could tell that
// what crossed into a public column was internal.
//
// So the check is here rather than at either call site. A caller cannot publish
// a brief by forgetting something — it has to pass the candidate through this
// function, and the function refuses on the caller's behalf, falling back to
// composed copy the caller supplies.
package datahelpers

import (
	"regexp"
	"strings"
)

// maxPublicMetaDescription is the length above which a string is treated as
// internal rather than public. Google truncates a meta description at roughly
// 155 characters, and real copy is written to that; the shortest observed brief
// in the bugs_open/103 census was 449 characters and the longest 1,206. 320 sits
// well clear of both — twice a usable description, and far under any brief
// measured — so it separates the two populations without judging borderline
// human copy.
const maxPublicMetaDescription = 320

// briefMarkers are phrases measured in the live bugs_open/103 census that appear
// in build briefs and effectively never in visitor-facing copy. They are the
// second signal because length alone would miss a SHORT brief, which is the
// failure this guard would otherwise still allow through.
//
// Deliberately narrow: each is an instruction to a generator, not a topic a page
// might legitimately be about. A page CAN be about "client-side experience"
// design, so the phrase is anchored to the brief's own construction where it can
// be.
var briefMarkers = regexp.MustCompile(`(?i)no fetch calls|elements, in order|embed [0-9]+ sample|fully self-contained client-side|no backend\)|:\s*\(1\)`)

// MetaDescriptionLooksInternal reports whether s reads as something written for
// a generator rather than for a visitor.
//
// It is exported because the same question is worth asking in a discovery check
// or a backfill, and a second copy of the rule is how the two would drift.
func MetaDescriptionLooksInternal(s string) bool {
	t := strings.TrimSpace(s)
	if t == "" {
		return false // empty is a different problem, not this one
	}
	if len([]rune(t)) > maxPublicMetaDescription {
		return true
	}
	return briefMarkers.MatchString(t)
}

// PublicMetaDescription returns the text that may be published as a page's meta
// description: candidate if it reads as public copy, otherwise composed.
//
// Both arguments are checked. A composed fallback that is itself brief-shaped is
// a bug in the caller, and returning it anyway would quietly reintroduce exactly
// what this function exists to stop, so the empty string is returned instead —
// a page with no meta description is a worse SEO outcome than a good one and a
// better one than a published spec.
//
// The second return value says whether the candidate was REJECTED, and exists
// because the council's bug_historian seat objected that the first version
// dropped a candidate with no log, no warning and no work item — a brand-new
// silent-drop path introduced by the fix for a silent-content problem. That was
// right. The reason it is a return value rather than a log call is that this
// package takes no logger dependency; the caller has one and knows the page.
//
// A caller that ignores `replaced` reintroduces the silence, so callers log it.
func PublicMetaDescription(candidate, composed string) (text string, replaced bool) {
	c := strings.TrimSpace(candidate)
	if c != "" && !MetaDescriptionLooksInternal(c) {
		return c, false
	}
	// Everything below this point means the candidate did not survive.
	if f := strings.TrimSpace(composed); f != "" && !MetaDescriptionLooksInternal(f) {
		return f, c != ""
	}
	return "", c != ""
}
