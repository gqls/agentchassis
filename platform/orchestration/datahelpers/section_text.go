// FILE: platform/orchestration/datahelpers/section_text.go
//
// One definition of "what text is this section actually saying", shared by the
// detector and the repair.
//
// WHY IT LIVES HERE AND NOT IN EITHER CALLER
// ------------------------------------------
// `discovery_checks.ContentDuplicationCheck` decides which sections are
// content-identical; `actions.RemoveDuplicatePageSectionsAction` deletes the
// duplicates it re-derives. Those two MUST agree on what "identical" means — if
// they drift, the detector flags one population and the repair deletes a
// different one, which is a content-loss bug of the class this codebase has been
// bitten by repeatedly (bugs_open/012, /021).
//
// `actions` cannot import `discovery_checks` without a cycle, so the first draft
// kept a copy in each package with a comment promising a drift test. A test that
// compares two implementations is a worse answer than not having two: both
// packages already import datahelpers and datahelpers imports neither, so the
// duplication was avoidable. Reuse before build — one function, no drift surface.
//
// WHAT IT DELIBERATELY IGNORES
// ---------------------------
// Keys naming assets or identifiers (url/href/src/image/icon/slug/id/class/
// target/colour/color) and values that look like paths, URLs or fragments. Two
// sections pointing at the same image are not saying the same thing. This filter
// is load-bearing rather than cosmetic: on vonc.com two unrelated sections
// matched at 1.00 similarity purely on captured footer/nav text before it
// existed.
//
// Map keys are walked in SORTED order. Go randomises map iteration, so without
// the sort the same content_data would normalise to different strings on
// different runs and "identical" would be a coin toss.
package datahelpers

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
)

var (
	sectionTagStripper  = regexp.MustCompile(`<[^>]+>`)
	sectionWSCollapser  = regexp.MustCompile(`\s+`)
	sectionAssetKeyLike = regexp.MustCompile(`(?i)(url|href|src|image|icon|slug|id|class|target|colour|color)$`)
	sectionTokenSplit   = regexp.MustCompile(`[^a-z0-9]+`)
)

// NormaliseSectionText reduces a page_components.content_data blob to the
// lower-cased, whitespace-collapsed prose it actually renders, so two sections
// can be compared for saying the same thing.
//
// Returns "" for a blob that is not valid JSON — callers treat that as "no
// comparable text" rather than as an empty match, because every unparseable blob
// would otherwise be identical to every other one.
func NormaliseSectionText(rawJSON string) string {
	var doc interface{}
	if err := json.Unmarshal([]byte(rawJSON), &doc); err != nil {
		return ""
	}
	var parts []string
	var walk func(v interface{}, key string)
	walk = func(v interface{}, key string) {
		switch t := v.(type) {
		case string:
			if sectionAssetKeyLike.MatchString(key) || len(t) < 3 {
				return
			}
			if strings.HasPrefix(t, "/") || strings.HasPrefix(t, "http") || strings.HasPrefix(t, "#") {
				return
			}
			parts = append(parts, t)
		case map[string]interface{}:
			keys := make([]string, 0, len(t))
			for k := range t {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				walk(t[k], k)
			}
		case []interface{}:
			for _, item := range t {
				walk(item, key)
			}
		}
	}
	walk(doc, "")
	joined := sectionTagStripper.ReplaceAllString(strings.Join(parts, " "), " ")
	return strings.ToLower(strings.TrimSpace(sectionWSCollapser.ReplaceAllString(joined, " ")))
}

// SectionTokenSet is the token set used for the similarity SCREEN in
// check_content_duplication. Tokens shorter than 4 characters are dropped:
// articles and glue words inflate every pair equally and only raise the floor.
//
// This is a screen for sizing a population, never a basis for an edit. See
// check_content_duplication.go for why — two sections can assert the identical
// facts while being 18% textually similar (measured, bugs_open/151).
func SectionTokenSet(text string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, t := range sectionTokenSplit.Split(text, -1) {
		if len(t) >= 4 {
			out[t] = struct{}{}
		}
	}
	return out
}
