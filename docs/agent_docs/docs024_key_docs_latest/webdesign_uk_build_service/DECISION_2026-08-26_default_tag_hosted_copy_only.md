# DECISION 2026-08-26 — the Google tag defaults ON, but only on the copy WE host

**Owner, 2026-08-26:** make the Google tag the default unless the customer specifies none
or a different one; asked whether to use his existing tag, or else drop the service and
rewrite the copy. **Ruled (option question, same day): his tag, HOSTED COPY ONLY.**

## The shape

- The 30-day site at our address (`<slug>.ugg2.com`) carries the OWNER'S GTM container by
  default. Purpose: we can see whether a delivered site is being visited at all (feeds the
  pre-delivery review and the discretionary-refund judgement, DECISION_2026-08-25).
- **The ZIP ships CLEAN.** No owner tag ever leaves in the customer's files: after
  handover their visitors' data is not ours, and the attested audience (experienced web
  designers) reads the files.
- The customer may supply THEIR OWN tag id at intake — it then goes into BOTH the hosted
  copy and the ZIP — or say "none", which removes the tag from the hosted copy too.

## Facts measured before the ruling (2026-08-26)

- The copy promises NO analytics service today: one outward-pointing FAQ mention
  (Fathom/Plausible) across all five pages. **Nothing needs rewriting in any branch.**
- ~~No per-site tag field exists (`sites.settings` carries none; zero Go references)~~
  > **CORRECTED 2026-08-26, same day, by the analytics_gtm lane: a FALSE ABSENCE.** The
  > field EXISTS: `site_specs` aspect `site_config`, `data.analytics.gtm_container_id`
  > (register STY-050, live since 07-31), read via `{{if .gtm_container_id}}` in every
  > head/header template — no literal in chrome at all. I measured `sites.settings` and
  > Go references; the seam is config-table state, and `site_config` was even in the
  > aspect list I had already read. A customer's own id and "none" both WORK TODAY
  > (write the key / leave it absent). What does NOT exist is the DEFAULT-seeder
  > (bugs_open/397 §6.2, analytics lane's build). WRONG_CALLS entry logged.

## What the build needs (a follow-up work package, not started)

1. ~~A per-site field~~ **EXISTS (correction above).** What remains: the DEFAULT-seeder
   plus an explicit `analytics.mode: default|custom|none` beside the id (adopted by the
   analytics lane 2026-08-26: once a seeder exists, "none" must be a stored fact the
   seeder honours, never an absence it would re-fill). The default value lives in ONE
   place both the seeder and the export path read. Structural build = analytics lane.
2. The ZIP path: PREFERRED design (analytics lane's suggestion, 2026-08-26) is
   re-RENDERING for export with mode=none so the template gate simply omits the block —
   no wrapper-marking, no regex over HTML. A marked wrapper is second-best if the
   exporter must work from stored artefacts. Route with site_delivery_and_editor.
3. Intake: an optional "your Google Tag id / no tag" question on the brief.
4. ONE attested copy line — **only when the mechanism ships** (reference only what
   exists): the hosted copy carries our tag unless you give us yours or say no; the ZIP
   is clean.
5. ~~Constraint: the default container must stay cookie-light~~ **RULED 2026-08-26
   (night): a SECOND, cookie-light GTM container for customer sites** — the owner's main
   container (GTM-PQ3WCTBD) is free to take GA4 for the estate; customer hosted copies
   default to the new container, which stays zero-cookie by construction. Creation =
   analytics_gtm lane (told same night); its id becomes the single-place fleet default
   the seeder and the export path read. If that container ever grows a cookie-setting
   tag, this default needs re-ruling.

   > **THE CONTAINER EXISTS — id recorded 2026-09-02, handed over by the analytics_gtm
   > lane (cross-session message):** **`GTM-TH5XGNQ4`** (second container under the
   > leopardessconsulting GTM account; owner created it 2026-09-02). Verified live by
   > that lane at 20:01Z: gtm.js serves 200, version 1, **0 tags, no G- id** — which for
   > THIS container is the CORRECT steady state, cookie-light by construction. ⚠ The
   > monitoring verdict reads INVERTED here: a tag or G- id ever APPEARING in it is this
   > section's re-ruling trigger, not progress. Check:
   > `CONTAINER=GTM-TH5XGNQ4 docs024_key_docs_latest/analytics_gtm/scripts/check_gtm_state.sh`.
   > Canonical written record until the seeder/export mechanism ships: `bugs_open/397`
   > §6.2, the analytics_gtm handoff, and this section. The seeder + `analytics.mode`
   > field build stays with the analytics_gtm lane per the ownership ruling; this lane's
   > intake question (item 3) and the ZIP path (item 2) now have their concrete default
   > value. Estate container GTM-PQ3WCTBD still had 0 tags at the same read — the
   > owner's GA4 Publish click remains the one outstanding Google action.

## Routing

Chrome mechanics: analytics_gtm lane (told, same day — their durable fix should leave
room for the per-site field rather than a hardcoded tag). ZIP: site_delivery_and_editor.
Intake + copy + register: this lane. Not a launch gate for webdesign.uk itself.
