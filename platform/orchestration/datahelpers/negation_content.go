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
}

// Field names that never hold prose. A false NEGATIVE here costs a missed tell;
// a false POSITIVE sends a URL to a model and asks it to rewrite the sentence,
// so the list is deliberately generous.
var nonProseFieldRe = regexp.MustCompile(`(?i)(^|_)(url|urls|href|src|id|ids|uuid|slug|key|class|classes|icon|image|img|colour|color|hex|type|kind|mode|status|position|order|count|value|amount|price|rate|date|time|email|phone|target|rel|name)$`)

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
			if prosey(key, v) {
				*out = append(*out, NegationTextField{Path: path, Text: v,
					Set: func(s string) { m[key] = s }})
			}
		case []interface{}:
			walkContentSlice(v, path, key, out)
		case map[string]interface{}:
			walkContentMap(v, path, out)
		}
	}
}

func walkContentSlice(sl []interface{}, path, key string, out *[]NegationTextField) {
	for i := range sl {
		idx := i
		p := path + "[" + strconv.Itoa(idx) + "]"
		switch v := sl[idx].(type) {
		case string:
			if prosey(key, v) {
				*out = append(*out, NegationTextField{Path: p, Text: v,
					Set: func(s string) { sl[idx] = s }})
			}
		case map[string]interface{}:
			walkContentMap(v, p, out)
		case []interface{}:
			walkContentSlice(v, p, key, out)
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

func prosey(field, value string) bool {
	if nonProseFieldRe.MatchString(field) {
		return false
	}
	t := strings.TrimSpace(value)
	if len(t) < 12 || !strings.ContainsAny(t, " \t\n") {
		return false
	}
	return !nonProseValueRe.MatchString(t)
}

// Headline-class fields: where the construction is least forgivable, because it
// is the first thing a reader meets and it is what the owner actually quoted.
// Matched on the LAST path segment, so items[2].headline counts.
var headlineFieldRe = regexp.MustCompile(`(?i)^(headline|sub_?headline|title|sub_?title|heading|sub_?heading|eyebrow|kicker|tagline|strapline|hero_text|lead)$`)

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
			findings = append(findings, map[string]interface{}{
				"field":    f.Path,
				"shape":    h.Shape,
				"matched":  h.Matched,
				"sentence": TruncateString(h.Sentence, 240),
				"headline": IsHeadlineField(f.Path),
			})
		}
	}
	return findings
}
