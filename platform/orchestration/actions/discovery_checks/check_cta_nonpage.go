// FILE: platform/orchestration/actions/discovery_checks/check_cta_nonpage.go
//
// Detects the CTA shape check_misdirected_cta is STRUCTURALLY BLIND to
// (bugs_open/299, slug home_page_cta_names_the_brief_starter_tool_and_dials_
// the_phone_instead): an anchor whose visible copy names a real page while its
// href is a NON-PAGE destination — tel: or mailto: in round 1. The sibling
// check skips every such anchor before classification (ClassifyLinkScope files
// tel:/mailto: under LinkScopeMailto and its loop `continue`s on anything that
// is not page/empty), which is how a home-page button reading "…with the
// Website Brief Starter…" whose href DIALLED THE PHONE survived two discovery
// passes (2026-08-14, 2026-08-17) on the very site it was live on.
//
// EXTERNAL hrefs are a STATED residue, not an oversight: the 2026-08-18 fleet
// calibration measured ~211 of 226 all-scope findings as false positives, and
// almost all were external — news-listing headlines, regulator links,
// documentation links — whose prose text token-matches a page on one
// incidental word. "Copy names our page, href leaves the site" therefore goes
// UNDETECTED until the matcher has a better discriminator than one-token
// overlap (identity-overlap thresholds are the candidate; see the 299 lane's
// calibration report). The classifier's own comment carries the numbers.
//
// Two findings, BOTH review-only (needs_human_review, no handler) in round 1:
//
//   - cta_names_nonpage_destination: copy token-matches a real page
//     (ctaClassifyAnchor — REUSED from the sibling, never forked) and the href
//     is tel:/mailto:/external. Whether the copy or the destination is the
//     mistake is a content judgement, so no auto-repair. Deliberately NOT a
//     page_rerender/cta_links_stale item: on any binary that predates the
//     non-page keep (applyCTARecompute KEEP #3), that item is exactly the
//     LANDMINES.md clobber trigger — the repair would replace the phone number
//     rather than fix the copy. Promotion to auto-repair is a recorded
//     follow-up once the keep is proven live.
//
//   - cta_tel_malformed: a tel: href datahelpers.NormalizeTelHref cannot
//     accept (or would rewrite): spaces/parens RFC 3966 forbids, or the
//     collapsed-trunk "+440…" that is undialable and cannot be auto-repaired
//     without guessing digits. The unambiguous forms self-heal on the next
//     build/rerender via the keep branches; this finding exists for the
//     refused residue and for tel: anchors outside the CTA field pairs.
//
// False-positive posture: generic phone/email copy ("Call us on +44 …",
// "Get in Touch") reduces to no page-naming tokens and is never a misdirect —
// the same stopword bar the sibling applies. The matcher is NOT tightened
// here: a two-token minimum would miss the motivating case itself ("See how
// it works" reduces to the single distinctive token "works").
//
// Scan surface is ctaComponentScanQuery — the SAME constant the sibling
// scans with, so the two CTA checks cannot disagree on which components (or
// which page lifecycles) they examine.
//
// Registration: automatic via init(). Enable by adding "cta_nonpage_destination"
// to a discovery agent's checks array AFTER the image carrying this file is
// live — an unregistered name fails the whole step.

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func init() { Register(&CTANonPageDestinationCheck{}) }

type CTANonPageDestinationCheck struct{}

func (c *CTANonPageDestinationCheck) Name() string { return "cta_nonpage_destination" }

// nonPageCTAFinding is one flagged anchor.
type nonPageCTAFinding struct {
	Kind                 string `json:"kind"` // cta_names_nonpage_destination | cta_tel_malformed
	SlotName             string `json:"slot_name"`
	Text                 string `json:"text"`
	Href                 string `json:"href"`
	SuggestedTarget      string `json:"suggested_target,omitempty"`       // misdirect only
	SuggestedTargetTitle string `json:"suggested_target_title,omitempty"` // misdirect only
	NormalizedHref       string `json:"normalized_href,omitempty"`        // tel_malformed, when unambiguous
	Why                  string `json:"why"`
}

// classifyNonPageAnchor is the whole per-anchor decision, extracted pure so it
// is testable without a database. Returns (nil, false) for anchors that are
// not this check's remit: page/empty/asset/anchor scopes (the sibling's and
// the phantom check's), javascript: and no-op hrefs (check_dead_controls') —
// and, deliberately, EXTERNAL hrefs.
//
// ROUND-1 SCOPE IS tel:/mailto: ONLY, set by calibration, not by taste. The
// full fleet run (2026-08-18, 698 anchors on the candidate surface) with
// external in scope produced 226 findings of which ~211 were false positives
// in two classes: anchor text that IS its own mailto address (tokens of
// "agents@…" accidentally match page titles), and legitimate external
// content links — news-listing headlines, regulator references (ico.org.uk),
// tool documentation links — whose prose text token-matches some page. A
// review queue that opens 200 wrong items protects nothing (the 248 lane
// measured the same failure on the excluded-area arm: 103 filed, 18/18
// sampled correct, all ignored). Narrowed to tel:/mailto:, the same run
// yields the true-positive classes only. External returns when it has a
// discriminator better than one-token overlap; that residue is stated in the
// header, not hidden.
func classifyNonPageAnchor(a datahelpers.Anchor, slotName string, pages []datahelpers.LabelMatchCandidate) (*nonPageCTAFinding, bool) {
	if datahelpers.ClassifyLinkScope(a.Href) != datahelpers.LinkScopeMailto {
		return nil, false
	}
	lower := strings.ToLower(strings.TrimSpace(a.Href))
	if strings.HasPrefix(lower, "javascript:") || datahelpers.IsNoopHref(a.Href) {
		return nil, false
	}

	// Self-agreement: copy that NAMES its own destination is correct however
	// its tokens score against page titles. Two forms, both live in the
	// calibration corpus:
	//   - mailto whose text contains the address ("Email us at x@y" → mailto:x@y);
	//   - tel whose text contains the number ("Call us on +44 (0) 7934 524 911"
	//     → tel:+447934524911) — compared on a trailing digit run, because
	//     display form and URI form legitimately differ (trunk "(0)",
	//     separators), which is also why the comparison is not string equality.
	// Self-agreement suppresses the MISDIRECT only — a phone button whose
	// copy states its own malformed number is still malformed, and the
	// malformed branch below must see it.
	selfAgrees := false
	if strings.HasPrefix(lower, "mailto:") {
		addr := lower[len("mailto:"):]
		if i := strings.IndexByte(addr, '?'); i >= 0 {
			addr = addr[:i]
		}
		selfAgrees = addr != "" && strings.Contains(strings.ToLower(a.Text), addr)
	}
	if strings.HasPrefix(lower, "tel:") && telTextNamesNumber(a.Text, a.Href) {
		selfAgrees = true
	}

	// Copy naming a real page while the href dials or mails: the anchor can
	// never agree with the page it names — a non-page href normalises to no
	// page URL — so named==true is always a mismatch here.
	if !selfAgrees {
		if misdirect, named := ctaClassifyAnchor(a, slotName, pages); named && misdirect != nil {
			return &nonPageCTAFinding{
				Kind:                 "cta_names_nonpage_destination",
				SlotName:             slotName,
				Text:                 a.Text,
				Href:                 a.Href,
				SuggestedTarget:      misdirect.SuggestedTarget,
				SuggestedTargetTitle: misdirect.SuggestedTargetTitle,
				Why:                  "copy names a real page; href is a non-page destination",
			}, true
		}
	}

	if strings.HasPrefix(lower, "tel:") {
		norm, okNorm := datahelpers.NormalizeTelHref(a.Href)
		switch {
		case !okNorm:
			return &nonPageCTAFinding{
				Kind: "cta_tel_malformed", SlotName: slotName, Text: a.Text, Href: a.Href,
				Why: "tel: href cannot be normalised without guessing (collapsed trunk prefix, junk characters, or out-of-range length) — a human must state the intended number",
			}, true
		case norm != a.Href:
			return &nonPageCTAFinding{
				Kind: "cta_tel_malformed", SlotName: slotName, Text: a.Text, Href: a.Href,
				NormalizedHref: norm,
				Why:            "tel: href carries separators RFC 3966 forbids; the keep branches normalise it on the next build/rerender of its CTA field — outside those fields this record is the only signal",
			}, true
		}
	}
	return nil, false
}

// telTextNamesNumber reports whether the anchor's visible text carries the
// same number its tel: href dials. Compared on the trailing 7-digit run of
// the href's digits, because the display form and the URI form legitimately
// differ — "+44 (0) 7934 524 911" against tel:+447934524911 shares no full
// digit string (the trunk "(0)" breaks contiguity) but the same tail.
func telTextNamesNumber(text, href string) bool {
	hrefDigits := digitsOnly(href)
	if len(hrefDigits) < 7 {
		return false
	}
	return strings.Contains(digitsOnly(text), hrefDigits[len(hrefDigits)-7:])
}

func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteByte(byte(r))
		}
	}
	return b.String()
}

func (c *CTANonPageDestinationCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	pages, _, err := loadCTAMatchIndex(dctx)
	if err != nil {
		return nil, err
	}

	rows, err := dctx.DB.QueryContext(dctx.Ctx, ctaComponentScanQuery, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("cta_nonpage_destination page query failed: %w", err)
	}
	defer rows.Close()

	type flagged struct {
		pageName, pageID string
		f                nonPageCTAFinding
	}
	var all []flagged

	for rows.Next() {
		var pageName, pageURL, pageID, slotName, html string
		if err := rows.Scan(&pageName, &pageURL, &pageID, &slotName, &html); err != nil {
			dctx.Logger.Warn("cta_nonpage_destination: scan error", zap.Error(err))
			continue
		}
		for _, a := range datahelpers.ExtractAnchors(html) {
			if f, ok := classifyNonPageAnchor(a, slotName, pages); ok {
				all = append(all, flagged{pageName: pageName, pageID: pageID, f: *f})
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cta_nonpage_destination iteration failed: %w", err)
	}

	result := &CheckResult{}
	for _, fl := range all {
		result.Findings = append(result.Findings, map[string]interface{}{
			"check":     "cta_nonpage_destination",
			"kind":      fl.f.Kind,
			"page_name": fl.pageName,
			"slot_name": fl.f.SlotName,
			"text":      fl.f.Text,
			"href":      fl.f.Href,
			"why":       fl.f.Why,
		})

		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":     "cta_nonpage_destination",
			"page_name": fl.pageName,
			"page_id":   fl.pageID,
			"finding":   fl.f,
			"fix": "Copy and destination disagree, or the tel: URI is malformed. " +
				"Content decision: point the href at the page the copy names, or rewrite " +
				"the copy for the authored destination (the *_target_title convention " +
				"carries it once the keep branches have run). Never bulk-promote to " +
				"cta_links_stale — see LANDMINES.md on the recompute clobber.",
		})

		var pageIDPtr *uuid.UUID
		if parsed, perr := uuid.Parse(fl.pageID); perr == nil {
			pageIDPtr = &parsed
		}

		severity, priority := "medium", 40
		if fl.f.Kind == "cta_tel_malformed" {
			severity, priority = "low", 30
		}
		summary := fmt.Sprintf("CTA %q on %s (%s): %s", fl.f.Text, fl.pageName, fl.f.SlotName, fl.f.Why)
		if len(summary) > 250 {
			summary = summary[:247] + "..."
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			PageID:   pageIDPtr,
			Source:   "discovery",
			Pipeline: "content",
			ItemType: fl.f.Kind,
			Severity: severity,
			Summary:  summary,
			SpecJSON: string(specJSON),
			Priority: priority,
			// review-only: no HandlerAgent, deliberately — see the header
			Status:    "needs_human_review",
			CreatedBy: dctx.AgentType,
			// Keyed page:slot:href — TEXT is excluded so a copy rewrite over
			// the same broken href cannot mint a second item (the dead_url
			// item-key lesson), while a changed href is a new fact.
			ItemKey: fmt.Sprintf("cta_nonpage:%s:%s:%s", fl.pageName, fl.f.SlotName, fl.f.Href),
			BatchID: dctx.BatchID,
		})
	}

	if len(result.WorkItems) > 0 {
		dctx.Logger.Warn("cta_nonpage_destination: found non-page CTA defects",
			zap.Int("count", len(result.WorkItems)),
			zap.String("site_id", dctx.SiteID.String()))
	}
	return result, nil
}
