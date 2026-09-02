// FILE: platform/orchestration/datahelpers/cta_label_agreement.go
//
// JudgeCTALabel is THE question "does this button's copy name the page its
// destination actually is?", asked in one place so that the half which DETECTS
// a wrong button and the half which WATCHES one being written cannot drift
// apart. That is RFC_047 §9's ruling in code: "writers refuse while the
// detector guesses" was considered and rejected as "a deliberate re-drift of
// the two halves — the exact thing bugs_open/203's extraction and
// bugs_open/308 both exist to stop".
//
// WHY THIS IS NOT A COMPARISON AGAINST *_target_title, which is what
// bugs_open/399 proposed and what the adjacent keys invite. The resolver writes
// <field>_target_title from the same `pages` row LoadCTALabelUniverse reads, so
// that title's tokens are ALREADY inside the destination candidate's
// identityTokens. A label-vs-title token test is therefore not a cheaper
// version of this question — it is a THIRD, weaker definition of "misdirected"
// beside the detector's and the writers', which is the drift class above. It is
// also the shape measured brittle here: bugfix_203's CALIBRATION_2026-08-11
// records all 9 already-correct CTA labels on gaswholesalers.com flipping to
// the wrong tool over a stray hyphen in "Break-Even".
// The title's job is in the RECORD (it names, in the operator's language, what
// the destination was believed to be) and in SCOPING (its presence marks a
// destination the resolver wrote). Not in the judgement.
//
// ⚠ SCOPE IS THE CALLER'S, DELIBERATELY — do not add a ClassifyLinkScope guard
// here. The two existing consumers filter hrefs differently and both are right:
// check_misdirected_cta admits LinkScopePage and LinkScopeEmpty, while
// check_cta_nonpage admits LinkScopeMailto ONLY and then asks this same
// question about tel:/mailto: buttons (check_cta_nonpage.go:141). A scope guard
// in here would read as tidying and would silently switch off misdirect
// detection on every phone and email button on the estate — a live check
// broken without being edited.
//
// WHAT IT CANNOT SEE, stated because the gap is structural and a later reader
// will otherwise assume coverage:
//   - A CONFIDENT FALSE MATCH (RFC_047 §8.3). dartsonline's "See how each brand
//     differs, spec by spec" resolves to /about.html on an identity-token win
//     ("spec"), not a tie. No tie rule sees that and no stopword list can.
//   - THE LABEL-LOCKED DEFECT (bugs_open/391). When the framework chose the
//     destination AND then told the writer to name it — which is exactly what
//     stampCTADestinationGuidance does — copy and destination AGREE and the
//     button is still pointed at the wrong page. Measured 2026-08-25: 16 of 17
//     resolver-minted fields on the password-entropy family were in that state,
//     including all three buttons the owner reported. Agreement between two
//     framework-written strings is evidence of CONSISTENCY, never of
//     CORRECTNESS. Only a ranking fix (bugs_open/391) or a copy pass
//     (cta_target_content_pass) reaches that class.

package datahelpers

// CTALabelVerdict is the answer to "does this copy name the page this
// destination is?". Three values, not two: the matcher declining to answer is
// a first-class outcome, never folded into "agrees" (which would silently
// bless every generic button) nor into "contradicts" (which would convict
// every one).
type CTALabelVerdict int

const (
	// CTALabelNoOpinion — the copy names no page the matcher will stand behind:
	// it is generic ("Get Started" reduces to zero distinctive tokens), it
	// matches nothing, it names the page it already sits on, or it names two
	// pages equally well. The last of these is RFC_047's ruled refusal and is
	// reported separately on the Ambiguous field, because "this button is
	// undecidable" is a real signal that today reaches nobody.
	CTALabelNoOpinion CTALabelVerdict = iota

	// CTALabelAgrees — the copy names a page and that page IS the destination.
	CTALabelAgrees

	// CTALabelContradicts — the copy names one page and the destination is a
	// different one. This is bugs_open/399's defect and the detector's
	// misdirected_cta finding: the same fact, seen before the bytes exist
	// rather than five lossy stages later.
	CTALabelContradicts
)

// String renders the verdict for records and logs. Kept stable: these strings
// land in agent_error_log context and any later query keys on them.
func (v CTALabelVerdict) String() string {
	switch v {
	case CTALabelAgrees:
		return "agrees"
	case CTALabelContradicts:
		return "contradicts"
	default:
		return "no_opinion"
	}
}

// CTALabelJudgement carries the verdict and, when there is one, the page the
// copy named. Named is what makes a Contradicts actionable by a human: "the
// words say X, the link goes to Y" is only useful with X in hand.
type CTALabelJudgement struct {
	Verdict CTALabelVerdict
	// Named is the page the copy names. Set iff Verdict is Agrees or
	// Contradicts; the zero value otherwise.
	Named LabelMatchCandidate
	// Silence says WHY there is no opinion, and it exists because "no opinion"
	// was hiding a population too large to leave as residue.
	//
	// The 391 lane made the case on 2026-08-26 with its own measurement: of 186
	// live mismatches, the copy names NO page at all in 95 — and a large part of
	// that bucket is copy expressing a destination KIND rather than a page
	// identity ("Book a discovery call", "Write to <address>"). Those are not
	// harmless: re-resolution sends them to whichever tool ranks first, so a
	// button reading "Write to …" opens an ROI estimator. They measured 23 such
	// contact-intent labels among 41 fields on one site, and one live on
	// leopardess/careers.html.
	//
	// ⚠ THIS IS A REASON CODE, NOT A SECOND QUESTION. JudgeCTALabel stays a
	// PAGE-IDENTITY test; it does not and must not classify intent. What the
	// reason buys is a SEAM: a caller that wants a kind-check can hang one off
	// SilenceNamesNothing without this function acquiring a second definition —
	// which is the drift RFC_047 §9 exists to stop, and the 391 lane explicitly
	// asked NOT to widen the judge for exactly that reason.
	Silence CTALabelSilence
}

// CTALabelSilence says why a NoOpinion verdict has no opinion. Zero value is
// SilenceNone, which is what Agrees and Contradicts carry.
type CTALabelSilence int

const (
	SilenceNone CTALabelSilence = iota
	// SilenceNamesNothing — the copy reduced to no distinctive tokens, or matched
	// no candidate at all. The large bucket (95 of 186 live mismatches as of
	// 2026-08-26) and the one where a destination-KIND check would go.
	SilenceNamesNothing
	// SilenceAmbiguous — the copy names two pages equally well. RFC_047's ruled
	// refusal: recorded, never acted on, because the owner ruled on 2026-08-23
	// that the undecidable case belongs to an agent that knows the site's
	// premise, not to a token counter.
	SilenceAmbiguous
	// SilenceNamesItsOwnPage — the copy names the page it already sits on, which
	// names nothing (bugs_open/308). Distinct from NamesNothing because it is a
	// CONTENT defect with a known shape: the button should not be there.
	SilenceNamesItsOwnPage
	// SilenceNamesIneligiblePage — the copy names a page whose row opts out of
	// CTA targethood (pages.eligible_as_cta_target=false, bugs_open/436). The
	// matcher found it and refused to stand behind it, exactly as for a
	// self-link. Distinct from NamesNothing because it is actionable copy: the
	// button promises a destination the estate has ruled out, so the copy —
	// not the link — is what needs a human's eye. Folding it into
	// NamesNothing would silence the only signal that the lock-in loop's raw
	// material (copy naming an excluded page) still exists on a page.
	SilenceNamesIneligiblePage
)

func (s CTALabelSilence) String() string {
	switch s {
	case SilenceNamesNothing:
		return "names_nothing"
	case SilenceAmbiguous:
		return "ambiguous"
	case SilenceNamesItsOwnPage:
		return "names_its_own_page"
	case SilenceNamesIneligiblePage:
		return "names_ineligible_page"
	default:
		return ""
	}
}

// Ambiguous preserves the original predicate for existing readers, so adding the
// reason code changed no call site.
func (j CTALabelJudgement) Ambiguous() bool { return j.Silence == SilenceAmbiguous }

// JudgeCTALabel answers the shared question for one (copy, destination) pair.
//
// pageName and pageURL identify the page the button SITS ON, because a label
// naming its own page names nothing (bugs_open/308's 2026-08-23 hand audit:
// 12% of that widening's writes were self-links). Callers hold different ones —
// the build path knows the name, the repair path knows both — so pass "" for
// whichever you lack, exactly as BestLabelMatchForPage documents.
//
// destination is compared on the NORMALISED path, so /contact/ and
// /contact/index.html are one destination and not a disagreement.
func JudgeCTALabel(label, destination string, candidates []LabelMatchCandidate,
	pageName, pageURL string) CTALabelJudgement {
	best, ok, ambiguous := BestLabelMatchForPage(label, candidates, pageName, pageURL)
	if !ok {
		if ambiguous {
			return CTALabelJudgement{Verdict: CTALabelNoOpinion, Silence: SilenceAmbiguous}
		}
		// Separate "names its own page" and "names an ineligible page" from
		// "names nothing". BestLabelMatchForPage folds all three into !ok, so
		// ask the underlying matcher whether it found a page at all: if it
		// did, one of the two refusal rules is why we are here. Self-link is
		// checked first, matching the refusals' order in BestLabelMatchForPage
		// — a page that is both its own page and opted out reports the more
		// specific content defect.
		if selfBest, selfOK, _ := BestLabelMatch(label, candidates); selfOK {
			if (pageName != "" && selfBest.Name == pageName) ||
				(pageURL != "" && NormalizePagePath(selfBest.URL) == NormalizePagePath(pageURL)) {
				return CTALabelJudgement{Verdict: CTALabelNoOpinion, Silence: SilenceNamesItsOwnPage}
			}
			if selfBest.IneligibleAsCTATarget {
				return CTALabelJudgement{Verdict: CTALabelNoOpinion, Silence: SilenceNamesIneligiblePage}
			}
		}
		return CTALabelJudgement{Verdict: CTALabelNoOpinion, Silence: SilenceNamesNothing}
	}
	if NormalizePagePath(best.URL) == NormalizePagePath(destination) {
		return CTALabelJudgement{Verdict: CTALabelAgrees, Named: best}
	}
	return CTALabelJudgement{Verdict: CTALabelContradicts, Named: best}
}
