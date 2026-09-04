// FILE: platform/orchestration/datahelpers/negation_content.go
//
// The define-by-negation scanner applied to a WRITER'S CONTENT MAP (bugs_open/305).
//
// A section's generated content is a JSON object whose values are the fields the
// component template reads: strings, lists of strings, and lists of objects
// (`item_fields`). Prose lives in some of those and never in others — a `cta_url`
// or a `stat_1_value` is not a sentence — so this file is the one place that
// decides what counts as prose, and both consumers use it:
//
//	RenderComponentAction  — annotates every rendered LLM section (count only)
//	RewriteNegationsAction — repairs, by replacing whole sentences in place
//
// One walker for both, because a repair that scanned a different set of fields
// from the annotation would produce a count nobody could reconcile with what was
// actually fixed.
//
// TWO QUESTIONS, TWO PREDICATES — ONE WALK (bugs_open/420, 2026-09-04).
// "Is this string worth SCANNING?" and "is this string safe to OVERWRITE with a
// model's answer?" are different questions, and they were both being answered by
// one field-name list. They fail in opposite directions: a scan false-negative
// ships a defect (420's two heading tics), while an overwrite false-positive
// corrupts an identifier other rows key on, silently, on a page that still
// renders. `name` is where the two answers differ — a feature card's `name` IS
// its heading, and a listing item's `name` IS the page slug its own `url` is
// built from — so arguing `name` onto one side of a single list was always going
// to be wrong for the other. So:
//
//	isProseContentField  — fails toward SCANNING. Generous. A false positive
//	                       costs a log line on the annotation, or one proposal
//	                       that still has to clear the identity flag below, the
//	                       claim scan and AcceptNegationRewrite.
//	identityContentField — fails toward EXCLUSION. Strict. Flags a value the
//	                       walker still YIELDS (so both consumers count it) but
//	                       that no writer may be asked to rewrite.
//
// The estate has ruled this shape twice before and both times the answer was to
// split the predicate rather than to argue a member in or out — markup_spans.go
// ("a writer exclusion is not a detector exclusion") and
// resolve_internal_links_action.go ("'never newly SEND' is a sound default;
// 'never TRUST an existing' is a different and much stronger claim that happens
// to reuse the same set", bugs_open/248's clobber). runtime_fill.go is the
// mechanical form copied here: one vocabulary, two named predicates of
// deliberately different strictness, each stating which way it fails.
//
// The flag is carried on the yielded field rather than filtered out of the walk,
// so the count stays reconcilable: the repair records an identity field as an
// EXEMPTION (like "regulatory"), keeping total = exempt + withinBudget + targets.
// A filter would make the repair's population quietly smaller than the
// annotation's, which is the thing the paragraph above forbids.
//
// WHY A SETTER RATHER THAN A REBUILT MAP. The renderer reads the very map this
// walker was given (`extractContentWithFallbacks` returns the live map out of
// CollectedData), so a repair must mutate in place: rebuilding it would leave the
// renderer reading the original and the gate would report success over copy
// nobody changed. That is the silent-no-op shape this estate keeps being bitten
// by, so it is designed out rather than guarded.

package datahelpers

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// NegationTextField is one scannable string inside a content map, with the means
// to write a new value back to exactly where it came from.
type NegationTextField struct {
	Path string // "headline", "items[2].description"
	Text string
	Set  func(string)

	// Identity is "" when a writer may be asked to rewrite this value, and
	// otherwise names WHY it must never be sent to a model. The field is still
	// scanned and still counted; only the repair reads this.
	Identity string
}

// IdentityNameWithURLSibling: a `name` beside a `url` is a listing item's page
// slug, not display copy.
//
// MEASURED 2026-09-03 over every object at any depth in page_components.content_data
// carrying a `name` key (1,729 items; re-taken twice as the population grew, and
// reproduced independently by the components lane from a query sharing no code):
//
//	url sibling, non-empty : 908 items,   0 with a space in `name`  -> IDENTITY
//	no url key at all      : 825 items, 752 with a space in `name`  -> display
//	url key but empty/null :   0 items                              -> does not occur
//
// Zero crossover. The identity half is resolvePagesWhereType's projection: `name`
// is the real `pages.name` on the same site and the item's own `url` is built
// from it, so a rewrite desynchronises the item from BOTH in one stroke — and the
// page still renders, which is the expensive kind of wrong.
//
// ⚠ WHY THE SIBLING AND NOT THE VALUE. Those 908 names are hyphenated lowercase
// slugs, so the value tests in isProseContentField skip them TODAY BY LUCK, NOT BY
// PROTECTION — one tokeniser change away from silent and estate-wide. Two lexical
// signals (908/908 lowercase-slug, 825/825 containing uppercase) partition the same
// 1,729 items identically, and they are deliberately NOT used: case is a property
// of who the current producers are, the sibling is a property of the shape. A test
// asserts the lexical agreement so we find out when that assumption breaks.
//
// ⚠ AND NOT THE STEM TEST. `url` ending `/<name>.html` is the precise relationship
// but holds for only 188 of 908 (20.7%) — `guide-x` lives at `/guides/x/index.html`,
// the type prefix dropped from the path — so it would leave 720 identity names
// resting on the luck above. Key PRESENCE is the test: an empty `url` is still a
// listing item, and the guard's direction is exclusion.
const IdentityNameWithURLSibling = "identity_name_with_url"

// Field names that never hold prose. A false NEGATIVE here costs a missed tell;
// a false POSITIVE sends a URL to a model and asks it to rewrite the sentence,
// so the list is deliberately generous.
//
// ⚠ THIS LIST IS NOT THE IDENTITY GUARD, and it used to be doing that job by
// accident. Bare `name` LEFT it on 2026-09-04 (bugs_open/420) because a card's
// `name` is its rendered heading; identity is now identityContentField's job,
// backed by `dropped_name` in AcceptNegationRewrite so a caller that never sees
// this walker is still covered.
//
// The `_name` SUFFIX arm STAYS, and dropping it with the bare word would have been
// the quiet half of that mistake: it is what keeps `company_name` (84 values as of
// 2026-09-03), `current_page_name` (71), `cardN_client_name`, `*_author_name` and
// `tool_name` out of the model's hands. Zero of them carry a gate shape today, so
// nothing is lost by leaving them excluded.
var neverProseFieldRe = regexp.MustCompile(`(?i)(?:(^|_)(url|urls|href|src|id|ids|uuid|slug|key|class|classes|icon|image|img|colour|color|hex|type|kind|mode|status|position|order|count|value|amount|price|rate|date|time|email|phone|target|rel)|_name)$`)

// A value that is plainly not a sentence: a path, a URL, a bare token, a number.
var nonProseValueRe = regexp.MustCompile(`(?i)^\s*(https?://\S+|/[A-Za-z0-9._~/-]*|#[0-9a-f]{3,8}|[\d.,%£$+-]+|[A-Za-z0-9_-]+)\s*$`)

// WalkContentStrings returns every prose-bearing string in a writer's content
// map, in a stable order (map keys sorted, list order preserved) so two runs over
// the same content produce the same paths.
func WalkContentStrings(content map[string]interface{}) []NegationTextField {
	var out []NegationTextField
	walkContentMap(content, "", &out)
	return out
}

func walkContentMap(m map[string]interface{}, prefix string, out *[]NegationTextField) {
	for _, k := range sortedContentKeys(m) {
		key := k
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		switch v := m[key].(type) {
		case string:
			if isProseContentField(key, v) {
				*out = append(*out, NegationTextField{Path: path, Text: v,
					Identity: identityContentField(key, m),
					Set:      func(s string) { m[key] = s }})
			}
		case []interface{}:
			walkContentSlice(v, path, key, m, out)
		case map[string]interface{}:
			walkContentMap(v, path, out)
		}
	}
}

// parent is the nearest enclosing object, carried down so a string element of a
// list is judged against the same siblings its object-shaped neighbours are. A
// list nested inside a list keeps the object above it: there is no closer one.
func walkContentSlice(sl []interface{}, path, key string, parent map[string]interface{}, out *[]NegationTextField) {
	for i := range sl {
		idx := i
		p := path + "[" + strconv.Itoa(idx) + "]"
		switch v := sl[idx].(type) {
		case string:
			if isProseContentField(key, v) {
				*out = append(*out, NegationTextField{Path: p, Text: v,
					Identity: identityContentField(key, parent),
					Set:      func(s string) { sl[idx] = s }})
			}
		case map[string]interface{}:
			walkContentMap(v, p, out)
		case []interface{}:
			walkContentSlice(v, p, key, parent, out)
		}
	}
}

// sortedContentKeys exists because GetMapKeys returns Go map order, i.e. random
// per call. A walker with unstable order produces unstable field paths, and a
// path is what a finding, a rejection reason and a re-ask all key on — the same
// class of defect as bugs_open/327's randomly-ordered `formatted`, where a diff
// of two identical briefs reports 100% changed.
func sortedContentKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// isProseContentField — PREDICATE A, the SCAN question. Fails toward scanning:
// a false positive costs a log line, or one proposal that must still clear the
// identity flag, the claim scan and AcceptNegationRewrite; a false negative ships
// the defect unscanned, which is bugs_open/420 itself.
func isProseContentField(field, value string) bool {
	if neverProseFieldRe.MatchString(field) {
		return false
	}
	t := strings.TrimSpace(value)
	if len(t) < 12 || !strings.ContainsAny(t, " \t\n") {
		return false
	}
	return !nonProseValueRe.MatchString(t)
}

// identityContentField — PREDICATE B, the OVERWRITE question. Fails toward
// exclusion: a false positive costs one unrepaired heading that the annotation
// still reports (visible, recoverable); a false negative overwrites a value other
// rows key on, persisted, on a page that still renders (silent, and the estate
// finds out from a broken link months later).
//
// Returns "" when a writer may rewrite the value, else the reason why not. A
// string, not a bool, so the marker's exempt_reasons says WHICH rule fired and a
// second rule needs no signature change. See IdentityNameWithURLSibling for the
// census this rests on and for the two discriminators deliberately rejected.
func identityContentField(field string, siblings map[string]interface{}) string {
	if !strings.EqualFold(field, "name") {
		return ""
	}
	// Key PRESENCE, not the value: an item whose url is blank is still a listing
	// item, and this predicate's job is to fail toward exclusion.
	if _, hasURL := siblings["url"]; hasURL {
		return IdentityNameWithURLSibling
	}
	return ""
}

// Headline-class fields: where the construction is least forgivable, because it
// is the first thing a reader meets and it is what the owner actually quoted.
// Matched on the LAST path segment, so items[2].headline counts.
//
// `name` joined on 2026-09-04 (bugs_open/420): a feature/differentiator card's
// `name` IS the heading the card renders, and all 37 of the live hits that fix
// exposed sit in exactly that shape (an object keyed name+description, slots
// `features` / `differentiators`).
//
// ⚠ THIS LIST IS THE SAME DUAL-PURPOSE SHAPE as the never-prose list above, and
// the fix is ORDER, not a second regex. Read-only, at :ScanContentDataForNegation,
// a false positive mislabels a report line. Mutating, at planNegationRepairs, a
// headline hit is ALWAYS repaired and never forgiven by the page budget — so a
// false positive here FORCES a rewrite that would otherwise have stood. The
// identity exemption runs BEFORE that branch, so the forcing arm can no longer
// reach an identity field; a test pins that ordering, because it is the whole of
// the protection and nothing about the two regexes shows it.
var headlineFieldRe = regexp.MustCompile(`(?i)^(headline|sub_?headline|title|sub_?title|heading|sub_?heading|eyebrow|kicker|tagline|strapline|hero_text|lead|name)$`)

// IsHeadlineField reports whether a content path names a headline-class field.
func IsHeadlineField(path string) bool {
	seg := path
	if i := strings.LastIndexByte(seg, '.'); i >= 0 {
		seg = seg[i+1:]
	}
	seg = regexp.MustCompile(`\[\d+\]$`).ReplaceAllString(seg, "")
	return headlineFieldRe.MatchString(seg)
}

// ScanContentDataForNegation is the report form used by the renderer's
// annotation: every hit in the content map, attributed to its field, with no
// judgement and no repair. Returns nil when clean, so a caller can attach the key
// only when there is something to say.
func ScanContentDataForNegation(content map[string]interface{}) []map[string]interface{} {
	var findings []map[string]interface{}
	for _, f := range WalkContentStrings(content) {
		for _, h := range ScanDefineByNegation(f.Text) {
			finding := map[string]interface{}{
				"field":    f.Path,
				"shape":    h.Shape,
				"matched":  h.Matched,
				"sentence": TruncateString(h.Sentence, 240),
				"headline": IsHeadlineField(f.Path),
			}
			// Only when set, so a finding on ordinary copy stays byte-identical
			// to what this function returned before bugs_open/420.
			if f.Identity != "" {
				finding["identity"] = f.Identity
			}
			findings = append(findings, finding)
		}
	}
	return findings
}
