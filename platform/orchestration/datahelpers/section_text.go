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

// SectionIdentityKey is the full definition of "these two rows are the same
// section rendered twice", shared by the detector and the repair for the same
// anti-drift reason as NormaliseSectionText above.
//
// SLOT EQUALITY IS NECESSARY, AND THAT IS NOT THE SLOT-KEYED RULE WE REJECTED
// ---------------------------------------------------------------------------
// bugs_open/156 rejected slot name as a SUFFICIENT condition: 10 of the fleet's
// 11 duplicate (page_id, slot_name) groups are legitimate repeated slots holding
// different content, so deleting on slot alone destroys real sections. Requiring
// slot equality as a NECESSARY condition is the opposite operation — it strictly
// SHRINKS the in-remit set, so it cannot introduce a deletion that content
// identity alone would not already have made.
//
// It is required because content_data does not always hold section content. On
// vonc.com/index.html two DIFFERENT components (`provocation-card` and
// `lobby-grid`, different component_id) both carry the byte-identical site-wide
// context blob — year, email, domain, nav_items, tone, _built_at — and nothing
// else. Their normalised text is therefore equal, and the pre-2026-07-31 rule
// (page + text, measured with the shipped function over live data) flagged them
// as one duplicate group and would have DELETED the lobby-grid row: a live
// section on the home page, removed because two unrelated components share
// boilerplate. The asset-key filter above does not save this — it strips
// url/id/class but not email/year/domain or nav labels, so on a boilerplate-only
// blob the boilerplate IS the comparison.
//
// Two different components in two different slots are not the same section
// rendered twice, whatever their content_data happens to contain.
//
// IDENTITY IS THE RAW BLOB, NOT THE NORMALISED PROSE — AND THAT IS DELIBERATE
// ---------------------------------------------------------------------------
// NormaliseSectionText is the right ruler for the similarity SCREEN and for
// sizing the residue. It is the wrong ruler for a DELETE, because everything it
// throws away is a difference that makes two rows not-interchangeable: every
// non-string value (numbers, booleans — a differing count, price, limit or flag),
// and every string under an asset-like key (a differing image, href or embedded
// id). Two rows whose prose matches but whose payload differs are two different
// sections, and deleting one loses the difference. That is the shape
// `bug_historian` objected to in council round 1
// (da3f2d9b-ae6f-492d-ad3b-748323b66367, HIGH, gating) and it is answered here
// rather than merely tested for.
//
// So: the deterministic half acts only where there is literally nothing to
// decide — same slot, byte-identical blob. Anything the normaliser calls similar
// but the bytes call different is residue, which needs judgement and gets a
// capability_gap and no handler.
//
// The caller must pass content_data::text READ FROM THE JSONB COLUMN. Postgres
// renders jsonb with sorted keys and normalised whitespace, so that form is
// canonical and a byte comparison is a true equality test on the document. A
// blob that has been through a Go map and re-marshalled is NOT canonical (Go
// randomises map order) — do not build this key from one.
//
// What this narrowing gives up, stated plainly: a genuine duplicate whose two
// rows differ in one stale asset URL is no longer in-remit. That is the intended
// trade — if the images differ, the rows are not interchangeable, and a human
// should decide which survives.
//
// SCOPE OF THIS CONTRACT — page_components ONLY (council round 2, architecture
// seat). The canonical-form guarantee above is a property of
// `page_components.content_data::text`. A third caller must NOT reuse this key
// against `site_plan_sections` rows, or any other table, without first
// establishing that ITS blob is canonical in the same sense; nothing in the
// signature enforces that boundary. If you need plan-side identity, derive it
// there and give it its own function rather than widening this one — the two
// tables are not interchangeable (site_plan_sections is the system of record,
// page_components is downstream of it).
func SectionIdentityKey(slotName, canonicalBlob string) string {
	// NUL cannot appear in either part, so the join is unambiguous — "a"+"bc"
	// and "ab"+"c" must not collide into one identity.
	return slotName + "\x00" + canonicalBlob
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
