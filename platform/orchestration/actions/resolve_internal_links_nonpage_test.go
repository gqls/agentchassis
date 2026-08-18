// FILE: platform/orchestration/actions/resolve_internal_links_nonpage_test.go
//
// Pins the NON-PAGE keep branches (bugs_open/299, slug home_page_cta_names_
// the_brief_starter_tool_and_dials_the_phone_instead) in BOTH CTA writers —
// setCTAField (build) and applyCTARecompute (repair) — plus the destination
// stamp helper. Sibling of resolve_internal_links_authored_destination_test.go
// (bugs_open/248's page-scheme half): the two keeps are DISJOINT (248's
// requires validPages membership, this one a non-page scheme), and the
// fixtures here mirror the live webdesign.uk shape the branch was built for —
// genuine "Call us on …" tel: buttons on faq/how-it-works that the positional
// pick would have replaced with a tool link.
//
// The malformed-tel cases double as the URI-repair pin: the write IS the
// repair (spaces/parens normalise on the next build/rerender), and the
// collapsed-trunk "+440…" is kept RAW — inventing digits is a human's call.

package actions

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func webdesignLikeCandidates(t *testing.T) []datahelpers.LabelMatchCandidate {
	t.Helper()
	briefStarter, ok := datahelpers.NewLabelMatchCandidate(
		"1", "tool-website-brief-starter", "Website Brief Starter | Tools",
		"/tools/website-brief-starter/index.html", true, "")
	if !ok {
		t.Fatal("brief-starter fixture produced no tokens")
	}
	return []datahelpers.LabelMatchCandidate{briefStarter}
}

func webdesignLikePages() datahelpers.PageURLSet {
	return datahelpers.NewPageURLSet([]string{
		"/index.html", "/how-it-works.html", "/tools/website-brief-starter/index.html",
	})
}

// --- setCTAField (build path) ---

// TestSetCTAFieldKeepsAuthoredPhoneButton: the core case. A stored tel: whose
// label names no page must be KEPT — normalised — not replaced by the
// positional pick. Deleting the keep branch fails this with the tool URL.
func TestSetCTAFieldKeepsAuthoredPhoneButton(t *testing.T) {
	positional := contentHub{Name: "tool-website-brief-starter",
		Title: "Website Brief Starter | Tools", URL: "/tools/website-brief-starter/index.html"}
	stored := map[string]interface{}{"secondary_cta_url": "tel:+44 (0) 7934 524 911"}

	resolved := map[string]interface{}{}
	var unresolved []map[string]interface{}
	setCTAField(resolved, stored, "secondary_cta_url", positional, webdesignLikePages(),
		"call-to-action", "call-to-action", "secondary", &unresolved,
		"Call us on +44 (0) 7934 524 911", webdesignLikeCandidates(t))

	if got := resolved["secondary_cta_url"]; got != "tel:+447934524911" {
		t.Errorf("secondary_cta_url = %v, want the NORMALISED kept tel:, not the positional pick", got)
	}
	if got := resolved["secondary_cta_target_title"]; got != "a phone call to +44 (0) 7934 524 911" {
		t.Errorf("secondary_cta_target_title = %v, want the computed destination phrase", got)
	}
	if len(unresolved) != 0 {
		t.Errorf("a kept destination must not be reported unresolved, got %v", unresolved)
	}
}

// TestSetCTAFieldLabelMatchStillBeatsStoredTel pins the branch ORDER — the
// same bar as bugs_open/248's verification #2: a stored wrong destination
// whose label names a real page is still repaired. This is exactly
// bugs_open/299 as filed (copy naming the Brief Starter, href dialling).
func TestSetCTAFieldLabelMatchStillBeatsStoredTel(t *testing.T) {
	positional := contentHub{} // irrelevant: label match decides first
	stored := map[string]interface{}{"secondary_cta_url": "tel:+44 (0) 7934 524 911"}

	resolved := map[string]interface{}{}
	var unresolved []map[string]interface{}
	setCTAField(resolved, stored, "secondary_cta_url", positional, webdesignLikePages(),
		"call-to-action", "call-to-action", "secondary", &unresolved,
		"Or answer a few short questions first with the Website Brief Starter", webdesignLikeCandidates(t))

	if got := resolved["secondary_cta_url"]; got != "/tools/website-brief-starter/index.html" {
		t.Errorf("secondary_cta_url = %v, want the label-matched page — label match stays AHEAD of the keep", got)
	}
}

// TestSetCTAFieldKeepsUndialableTelRaw: the collapsed-trunk form the
// normaliser refuses is kept AS AUTHORED — never "repaired" by guessing, and
// never surrendered to the positional pick. check_cta_nonpage files it.
func TestSetCTAFieldKeepsUndialableTelRaw(t *testing.T) {
	positional := contentHub{Name: "tool-website-brief-starter",
		Title: "Website Brief Starter | Tools", URL: "/tools/website-brief-starter/index.html"}
	stored := map[string]interface{}{"secondary_cta_url": "tel:+4407934524911"}

	resolved := map[string]interface{}{}
	var unresolved []map[string]interface{}
	setCTAField(resolved, stored, "secondary_cta_url", positional, webdesignLikePages(),
		"hero", "hero", "secondary", &unresolved, "Call us", webdesignLikeCandidates(t))

	if got := resolved["secondary_cta_url"]; got != "tel:+4407934524911" {
		t.Errorf("secondary_cta_url = %v, want the RAW undialable tel kept for a human", got)
	}
}

// TestSetCTAFieldJavascriptHrefIsNotAuthored: javascript:/no-op hrefs are dead
// controls (check_dead_controls' remit), not authored destinations — they fall
// through to the positional pick exactly as before this branch existed.
func TestSetCTAFieldJavascriptHrefIsNotAuthored(t *testing.T) {
	positional := contentHub{Name: "tool-website-brief-starter",
		Title: "Website Brief Starter | Tools", URL: "/tools/website-brief-starter/index.html"}
	stored := map[string]interface{}{"cta_url": "javascript:void(0)"}

	resolved := map[string]interface{}{}
	var unresolved []map[string]interface{}
	setCTAField(resolved, stored, "cta_url", positional, webdesignLikePages(),
		"hero", "hero", "primary", &unresolved, "", webdesignLikeCandidates(t))

	if got := resolved["cta_url"]; got != positional.URL {
		t.Errorf("cta_url = %v, want the positional pick — a dead control is not an authored destination", got)
	}
}

// --- applyCTARecompute (repair path, KEEP #3) ---

// TestApplyCTARecomputeKeepsAndNormalisesPhoneButton: the LANDMINES clobber,
// closed. A cta_links_stale rerender of a page holding a genuine malformed
// phone button must repair the URI, not the destination.
func TestApplyCTARecomputeKeepsAndNormalisesPhoneButton(t *testing.T) {
	target := contentHub{Name: "tool-website-brief-starter",
		Title: "Website Brief Starter | Tools", URL: "/tools/website-brief-starter/index.html"}
	stored := map[string]interface{}{"secondary_cta_url": "tel:+44 (0) 7934 524 911"}

	resolved := map[string]interface{}{}
	applyCTARecompute(resolved, stored, "secondary_cta_url", target, webdesignLikePages(),
		"/faq.html", "Or call us on +44 (0) 7934 524 911, if you'd rather talk it through.",
		webdesignLikeCandidates(t))

	if got := resolved["secondary_cta_url"]; got != "tel:+447934524911" {
		t.Errorf("secondary_cta_url = %v, want the normalised kept tel:, not the positional target", got)
	}
	if got := resolved["secondary_cta_target_title"]; got != "a phone call to +44 (0) 7934 524 911" {
		t.Errorf("secondary_cta_target_title = %v, want the computed destination phrase", got)
	}
}

// TestApplyCTARecomputeKeepsMailto: same keep, email form, no normalisation.
func TestApplyCTARecomputeKeepsMailto(t *testing.T) {
	target := contentHub{Name: "tool-website-brief-starter",
		Title: "Website Brief Starter | Tools", URL: "/tools/website-brief-starter/index.html"}
	stored := map[string]interface{}{"primary_cta_url": "mailto:hello@webdesign.uk"}

	resolved := map[string]interface{}{}
	applyCTARecompute(resolved, stored, "primary_cta_url", target, webdesignLikePages(),
		"/faq.html", "Send us an email", webdesignLikeCandidates(t))

	if got := resolved["primary_cta_url"]; got != "mailto:hello@webdesign.uk" {
		t.Errorf("primary_cta_url = %v, want the kept mailto:", got)
	}
}

// TestApplyCTARecomputeLabelMatchStillBeatsStoredTel: branch order on the
// repair path too — the misdirect the detector files (copy names a real page,
// href dials) is exactly what a rerender must still fix.
func TestApplyCTARecomputeLabelMatchStillBeatsStoredTel(t *testing.T) {
	stored := map[string]interface{}{"secondary_cta_url": "tel:+44 (0) 7934 524 911"}

	resolved := map[string]interface{}{}
	applyCTARecompute(resolved, stored, "secondary_cta_url", contentHub{}, webdesignLikePages(),
		"/index.html", "Answer a few questions with the Website Brief Starter",
		webdesignLikeCandidates(t))

	if got := resolved["secondary_cta_url"]; got != "/tools/website-brief-starter/index.html" {
		t.Errorf("secondary_cta_url = %v, want the label-matched page — the misdirect repair still wins", got)
	}
}

// --- the destination stamp ---

// TestStampCTADestinationGuidance: the datum lands on the paired LABEL field's
// spec description (the pipe the writer prompt renders), leaves sibling specs
// untouched, and no-ops on an unknown label field or absent specs.
func TestStampCTADestinationGuidance(t *testing.T) {
	section := map[string]interface{}{
		"llm_field_specs": []interface{}{
			map[string]interface{}{"name": "headline", "description": "The heading."},
			map[string]interface{}{"name": "secondary_cta", "description": "Short button text."},
		},
	}
	resolved := map[string]interface{}{
		"secondary_cta_target_title": "a phone call to +44 (0) 7934 524 911",
	}

	stampCTADestinationGuidance(section, "secondary_cta", resolved, "secondary_cta_url")

	specs := section["llm_field_specs"].([]interface{})
	cta := specs[1].(map[string]interface{})
	want := "Short button text. Destination (fixed): a phone call to +44 (0) 7934 524 911. " +
		"Write this CTA's text to name or clearly promise this destination; never promise a different one."
	if got := cta["description"]; got != want {
		t.Errorf("stamped description = %q, want %q", got, want)
	}
	if got := specs[0].(map[string]interface{})["description"]; got != "The heading." {
		t.Errorf("sibling spec mutated: %q", got)
	}

	// no title resolved → untouched
	before := cta["description"]
	stampCTADestinationGuidance(section, "secondary_cta", map[string]interface{}{}, "secondary_cta_url")
	if cta["description"] != before {
		t.Error("stamp ran without a resolved title")
	}
	// unknown label field / absent specs → no panic, no write
	stampCTADestinationGuidance(section, "", resolved, "secondary_cta_url")
	stampCTADestinationGuidance(map[string]interface{}{}, "secondary_cta", resolved, "secondary_cta_url")
}
