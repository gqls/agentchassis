// FILE: platform/orchestration/datahelpers/links_tel.go
//
// The shared vocabulary for AUTHORED NON-PAGE CTA destinations — tel:, mailto:,
// external http(s), and named in-page fragments — plus the one tel: normaliser.
//
// Why this exists (bugs_open/299, slug home_page_cta_names_the_brief_starter_
// tool_and_dials_the_phone_instead): every CTA keep/repair predicate in the
// platform was keyed on validPages.Contains, which is FALSE for any non-page
// href. So a genuine "Call us on …" button could never take a keep branch and
// fell to the positional pick — the LANDMINES.md "cta_links_stale repair CANNOT
// tell a genuine 'Get in Touch' from …" trap in a second form (the page-scheme
// half is bugs_open/248's storedCTADestinationIsAuthored; the two predicates
// are deliberately DISJOINT — 248's requires validPages membership, this one
// requires a non-page scheme, so no url can satisfy both).
//
// One definition, three consumers: setCTAField (build), applyCTARecompute
// (repair), check_cta_nonpage (detection). The drift class this package keeps
// closing — two spellings of one judgement — is why these live here and not
// inline at any call site.

package datahelpers

import "strings"

// IsAuthoredNonPageCTADestination reports whether an href is a deliberate
// destination that is not an internal page: tel:, mailto:, an external
// http(s)/protocol-relative URL, or a NAMED in-page fragment.
//
// Deliberately EXCLUDED, with reasons:
//   - javascript: — ClassifyLinkScope lumps it into LinkScopeMailto, but a
//     javascript: href is a control, not a destination; treating it as
//     authored would preserve dead controls for ever.
//   - no-op hrefs (IsNoopHref: "#", "#!", javascript:void…) — those are
//     check_dead_controls' remit. One owner per class.
//   - empty, page paths, asset paths — the existing page machinery's remit.
func IsAuthoredNonPageCTADestination(href string) bool {
	h := strings.TrimSpace(href)
	if h == "" || IsNoopHref(h) {
		return false
	}
	lower := strings.ToLower(h)
	switch {
	case strings.HasPrefix(lower, "javascript:"):
		return false
	case strings.HasPrefix(lower, "tel:"),
		strings.HasPrefix(lower, "mailto:"):
		return true
	case strings.HasPrefix(lower, "http://"),
		strings.HasPrefix(lower, "https://"),
		strings.HasPrefix(lower, "//"):
		return true
	case strings.HasPrefix(h, "#"):
		return true // a NAMED fragment — bare "#"/"#!" already excluded above
	}
	return false
}

// NormalizeTelHref canonicalises a tel: href to RFC 3966 shape: "tel:" plus an
// optional leading "+" and 4–15 digits, nothing else. ok=false means the input
// is not a tel: href, or cannot be normalised WITHOUT GUESSING — the caller
// must keep the original and route it to a human, never invent digits.
//
// The two traps this encodes, both measured live on webdesign.uk (2026-08-18):
//
//  1. "tel:+44 (0) 7934 524 911" — spaces are not legal in RFC 3966 and the
//     "(0)" is a UK trunk prefix that must be DROPPED in international form.
//     The "(0)" group is removed FIRST, while the "+" prefix is still visible
//     and before separators are stripped, because…
//  2. "tel:+4407934524911" — …stripping the parens first COLLAPSES the trunk
//     zero into the number, which is undialable. That exact value is live on
//     webdesign.uk/contact. A "+440…" result is therefore refused rather than
//     repaired: the intent (drop the zero? a typo for something else?) is a
//     human's call. The refusal is specifically "+440" — other countries
//     (e.g. Italy, +39 0…) legitimately keep a leading zero, so a general
//     "+<cc>0" rule would be wrong. The table grows by evidence, not
//     speculation.
//
// Visual separators the fleet actually writes (space, "-", ".", "(", ")") are
// stripped; anything else — letters, ";ext=", "," — returns ok=false.
func NormalizeTelHref(href string) (normalized string, ok bool) {
	h := strings.TrimSpace(href)
	if len(h) < 4 || !strings.EqualFold(h[:4], "tel:") {
		return "", false
	}
	num := strings.TrimSpace(h[4:])

	// Drop an explicit "(0)" trunk-prefix group — only in international form,
	// and BEFORE separator stripping (see trap 2 above).
	if strings.HasPrefix(num, "+") {
		num = strings.Replace(num, "(0)", "", 1)
	}

	var b strings.Builder
	for i, r := range num {
		switch {
		case r >= '0' && r <= '9':
			b.WriteByte(byte(r))
		case r == '+' && i == 0:
			b.WriteByte('+')
		case r == ' ' || r == '-' || r == '.' || r == '(' || r == ')':
			// visual separator — drop
		default:
			return "", false // not ours to guess
		}
	}
	cleaned := b.String()

	digits := strings.TrimPrefix(cleaned, "+")
	if len(digits) < 4 || len(digits) > 15 { // 15 = E.164 maximum
		return "", false
	}
	if strings.HasPrefix(cleaned, "+440") {
		return "", false // collapsed UK trunk prefix — undialable, human's call
	}
	return "tel:" + cleaned, true
}

// DescribeCTADestination renders a non-page destination as the short human
// phrase the *_target_title convention carries for pages, so a content writer
// told "the destination is already fixed: write copy FOR it" has something to
// write for. Uses the AUTHORED display form (a visitor reads "+44 (0) 7934…",
// not E.164). Returns "" for anything that is not an authored non-page
// destination — callers gate on IsAuthoredNonPageCTADestination first.
func DescribeCTADestination(href string) string {
	h := strings.TrimSpace(href)
	lower := strings.ToLower(h)
	switch {
	case strings.HasPrefix(lower, "tel:"):
		return "a phone call to " + strings.TrimSpace(h[4:])
	case strings.HasPrefix(lower, "mailto:"):
		addr := h[7:]
		if i := strings.IndexByte(addr, '?'); i >= 0 {
			addr = addr[:i] // ?subject=… is plumbing, not identity
		}
		return "an email to " + strings.TrimSpace(addr)
	case strings.HasPrefix(lower, "http://"), strings.HasPrefix(lower, "https://"), strings.HasPrefix(lower, "//"):
		return "an external site: " + hostOfURL(h)
	case strings.HasPrefix(h, "#") && !IsNoopHref(h):
		return "a section on this page (" + h + ")"
	}
	return ""
}

// hostOfURL extracts the host from an absolute or protocol-relative URL
// without net/url's error surface — for display only, never for routing.
func hostOfURL(u string) string {
	rest := u
	if i := strings.Index(rest, "//"); i >= 0 {
		rest = rest[i+2:]
	}
	for j, r := range rest {
		if r == '/' || r == '?' || r == '#' {
			return rest[:j]
		}
	}
	return rest
}
