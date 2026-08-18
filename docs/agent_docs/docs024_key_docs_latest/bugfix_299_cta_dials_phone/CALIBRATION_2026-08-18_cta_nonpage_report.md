# CALIBRATION 2026-08-18 — cta_nonpage_destination, fleet run before shipping

Method: the REAL classifier (`classifyNonPageAnchor`, run in-package via a throwaway
env-guarded test, deleted after — never a SQL re-implementation, per the LANDMINES
NormalizePagePath lesson) over a dump of every `page_components.rendered_html` on
non-deleted/archived pages whose HTML carries a tel:/mailto:/http/`//` href
(94 components), with per-site candidates built exactly as `loadCTAMatchIndex` builds
them. Disconfirming shapes stated up front: a flood means the matcher is wrong; a run
flagging ONLY the motivating anchor is equally suspect (a detector tuned to its example).

## Round A — scope as originally chosen (tel:/mailto: + external): REJECTED BY MEASUREMENT

698 anchors examined, **226 flagged, ~211 false positives** in two classes:
- anchor text that IS its own mailto address ("agents@contactforsales.com" →
  mailto:same) — the address tokens accidentally match page titles;
- legitimate EXTERNAL content links: news-listing headlines (dartsonline, relojistas,
  robot-hands, gaswholesalers), regulator/reference links (ico.org.uk, fca.org.uk,
  "Test in Google"), whose prose text token-matches a page on one incidental word.

A review queue opening 200 wrong items protects nothing — the 248 lane measured the
same failure shape on the excluded-area arm (103 filed, 18/18 sampled correct, all
ignored). **External is therefore OUT of round 1, as a stated residue in the check
header**, until the matcher has a better discriminator than one-token overlap.
This narrows the owner's earlier scope choice (tel/mailto + external) on measurement.

## Round B — shipped scope (tel:/mailto:, with self-agreement suppression): PASS

Self-agreement rules added: a mailto whose text contains its own address, and a tel
whose text states the dialled number (trailing-7-digit comparison — display and URI
forms legitimately differ), can never be misdirects. Suppresses the misdirect ONLY —
a self-stating malformed tel is still malformed.

**698 anchors, 17 flagged, hand-reviewed: 17/17 defensible. 0 false positives.**

| kind | count | rows |
|---|---|---|
| cta_names_nonpage_destination | 2 | webdesign.uk faq/hero "See how it works" → tel:; faq/call-to-action "See how the process works" → tel: — the motivating class, exactly |
| cta_tel_malformed | 15 | 8× `tel:+44 (0) 7934 524 911` (spaces/parens, contact-info across 8 sites); 4× empty `tel:` (leopardess ×2, gamesdesign, robot-hands); 1× collapsed-trunk `tel:+4407934524911` (webdesign contact/hero — the undialable form, refused by design, a human names the number); 1× national-with-spaces `tel:07934 524 911` (dartsonline); plus webdesign index "Read the full terms" → tel: and how-it-works "See what's included" → tel: (copy naming nothing that exists — caught by the malformed arm since the misdirect arm rightly declines) |

Note the check sees MORE malformed tels (15) than the content_data census (5): contact-info
renders phones from SITE IDENTITY (RenderContext Email/Phone), not content_data — the
detector reads the served artefact, which is the correct surface.

The motivating anchor's current form (index, text "Read the full terms", href tel:) is
flagged; the original filing's form ("…Website Brief Starter…" → tel:) is pinned in
`check_cta_nonpage_test.go` and classifies as the misdirect.

## Keep-branch disposition half (the "zero page-scheme rows change" proof)

Not a run but a predicate argument + census, stated as such: the new keeps fire only when
`IsAuthoredNonPageCTADestination(stored)` — false for every page-scheme URL (boundary pin
`("/contact.html") == false` in links_tel_test.go), so 248's territory is untouched by
construction. Rows whose disposition changes = exactly the non-page stored CTA urls:
5 tel + 2 external + 1 mailto + 1 anchor fleet-wide (census 2026-08-18, RUNBOOK query).
Effect: kept (tel normalised where unambiguous) instead of positionally replaced or
left to the carry.
