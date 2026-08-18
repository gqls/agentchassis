# PLAN 2026-08-18 — `bugs_open/299` (slug `home_page_cta_names_the_brief_starter_tool_and_dials_the_phone_instead`)

> **Refer to this bug BY SLUG.** 299 is an ambiguous number — `bugs_closed/299` is the
> unrelated skipped-render-audit case.

## What we are fixing

A CTA on preview.webdesign.uk/index reads "See how it works" and its href is
`tel:+44 (0) 7934 524 911`. The framework question (the part that survives the site rewrite):
why does a chassis carrying the `bugs_closed/268` fix still produce a label/destination pair
that disagree, and why does nothing detect it?

## The mechanism, measured (2026-08-18)

Four independent defects and one wiring defect compose:

1. **Producer wiring (the proximate cause — `bugs_open/312`, filed by this lane).**
   `page-content-writer`'s `select_sections` extracts
   `resolved_links.response.link_resolution.sections_ready` — a path with a `link_resolution`
   level the resolver's response does not have (resolution object sits directly under
   `response`). 0 of 150 retained runs match; the silent fallback feeds the render the
   pre-resolver plan, whose resolved_data is the PBP-039 carry of the stored row. On the
   traced index build (orch `05e3839d`, child `a907e946`, 08-18 10:27) the resolver computed
   `/tools/website-brief-starter/index.html` for BOTH CTA fields, with target titles, and the
   render consumed `/contact.html` + the tel: instead. **The resolver's label-match fixes
   (203-follow-on, 253) have been inert on every fresh build in the retained window.**
2. **Detector blindness.** `check_misdirected_cta.go` skips every anchor whose
   `ClassifyLinkScope` is not page/empty; `tel:`/`mailto:`/`javascript:` classify as
   `LinkScopeMailto`. The check ran on this site 08-14 and 08-17 with the broken anchor live.
3. **Repair clobber (non-page half of the `LANDMINES.md` bug-203 trap).**
   `applyCTARecompute`'s keep branch requires `validPages.Contains(current)` — false for any
   non-page href — so a genuine phone button falls to the positional pick. The page-scheme
   half is `bugs_open/248` (owned elsewhere, active, complementary). Their notes now hold a
   CONFIRMED production instance (finetuning.uk/services, 08-17 19:11).
4. **`tel:` shape unvalidated.** 5 tel: CTA urls fleet-wide (all webdesign.uk): 4 with
   spaces/parens (RFC 3966 forbids spaces), 1 undialable (`tel:+4407934524911` — the `(0)`
   collapsed in). Both the raw and the naively-cleaned forms are broken.
5. **Archived pages leak into the scan.** No `p.status` filter on the check's page query;
   `index-rejected-v1-20260806` minted two `cta_links_stale` items, both failed.

## The fix (approved by the owner 2026-08-18; full detail in the session plan)

1. **`datahelpers/links_tel.go`** (new): `IsAuthoredNonPageCTADestination` (tel/mailto/
   external/named-fragment; explicitly NOT `javascript:`/noop — those are
   `check_dead_controls`' remit), `NormalizeTelHref` (drop `(0)` while `+` present and BEFORE
   stripping separators; refuse `+440…`), `DescribeCTADestination` (human phrase for the
   `*_target_title` convention, authored display form).
2. **Keep branches in both writers** — `setCTAField` (gains an `existingURL` param, read from
   the already-loaded `loadExistingSectionContentData` map) and `applyCTARecompute`: an
   authored non-page destination is KEPT (a keep WRITES — 248's finding), tel normalised
   where `ok`, companion `*_target_title` written. Branch order matches 248's: label-match
   ahead of keep. **Default ON** (owner decision 08-18: unconditional, calibrated first —
   the branch removes authority; divergence from the 2026-08-02 ruling stated to council
   with the PBP-039 precedent).
3. **Guidance stamp** (`stamp_cta_destination_guidance`, opt-in default OFF, single live
   consumer `internal-link-resolver`): append "Destination (fixed): <title>…" to the paired
   label field's `llm_field_specs[].description` — the pipe the writer prompt already
   renders. Measured: the `*_target_title` VALUE reaches 0 prompts today; only the guidance
   sentence naming it does.
4. **New check `cta_nonpage_destination`** (new name = the opt-in; `misdirected_cta` stays
   byte-identical): reuses `ctaClassifyAnchor`; findings `cta_names_nonpage_destination` and
   `cta_tel_malformed`, both review-only round 1 (auto-repair only after §2 proven live).
5. **Archived filter** on the misdirected scan query (unconditional; spelled as
   `loadCTAMatchIndex` spells it — the shared constant is import-cycle-unreachable).
6. **Wiring fix for 312**: `select_sections` path 1 → `resolved_links.response.sections_ready`.
   **`_HOLD` migration — see the interlock below.**

## ⚠ THE INTERLOCK (do not reorder)

The dead wiring is currently an accidental safety device: the traced run proves fixing it
against today's binary repoints the authored "Get in touch" → `/contact.html` at the
brief-starter tool (setCTAField has no keep branch at all). **Order: Go commit → build →
roll → pod-verify → keep halves proven (coordinate 248's) → only then unhold the wiring
migration → then arm the check + stamp.**

## Owner decisions (asked 2026-08-18)

1. Wiring fix gated on BOTH keep halves (recommended) or mine alone — awaiting.
2. Home secondary CTA: phone button with corrected copy (default) or Brief Starter link
   (webdesign lane's move) — awaiting.
3. Confirm intended number for contact/hero's undialable `tel:+4407934524911`
   (display text says +44 (0) 7934 524 911 ⇒ `tel:+447934524911`) — awaiting.
4. Enable the new check after calibration review — awaiting the calibration numbers.

## Not doing, and why

- Widening the repair's label-match candidate set — `cta_target_content_pass` lane's.
- Page-scheme keep (utility pages) — `bugs_open/248`'s, active; we contribute, not compete.
- Rewriting webdesign.uk copy — the webdesign lane's and the owner's.
