# PLAN — bugfix 203 follow-on: phantom-CTA cleanup + class closure

**Started 2026-08-06** by the session that claimed the remaining work in `bugs_open/203`
(claim commit `dfdcdecd2`). The SOURCE fix (`880a405a6`) is done, council-APPROVED
(corr `42eda9a5`, r1, 3 advisory objections none high) and **LIVE** — see Decisions D1/D2.

## What this lane owes

1. ~~Read the council verdict for `42eda9a5`~~ **DONE** — APPROVED r1. Full report read
   from `diagnosis_artifacts` (kind=`council_report`). The advisory objections are
   **inputs to this plan**, listed below.
2. ~~Prove the source fix live~~ **DONE** — see D2.
3. Re-run the 13-instance census (08-05 figures are stale; the census reads STORED
   `page_components.rendered_html`, which no rerender has touched — LANDMINE: the
   phantom check reads stored html, not served).
4. Clean up the shipped instances **through the framework**: rerender on the fixed
   binary regenerates from `content_data`, so the fabricated href cannot re-ship
   (the phantom lived in render-time context, never in `content_data`). Where a real
   CTA target exists, prefer re-running `resolve_internal_links` first so the button
   comes back correct rather than absent. NOT hand-set URLs (owner ruling 2026-08-04:
   the framework writes the content, not us).
5. The **class audit** the council asked for (bug_historian, medium): every fabricated
   default in `component_library.go` paired with a presence guard — found live already:
   `contextToMap`'s defaults map fabricates `primary_cta_url: /contact.html` and
   `secondary_cta_url: /about.html` (lines ~1136–1147 at `880a405a6`), and the alias
   block above it copies `cta_url → primary_cta_url`. Same class, still at HEAD.
6. Detector coverage: `check_misdirected_cta` caught 2 of 13 — measure its cadence and
   scope before proposing anything.

## Council objections carried as constraints (corr 42eda9a5, read 2026-08-06)

- **bug_historian M1**: audit the whole fabricate-a-fallback class in this file, not
  per-field patches. (= item 5 above.)
- **bug_historian M2 / guardian L**: `contextToMap` is the regex-fallback path; its
  defaults exist "to prevent raw {{.field}}" (the file's own comment). An absent key on
  that path may ship LITERAL template syntax, not a hidden button. Any fix to the
  `primary_*` defaults must first establish what the regex renderer does with an empty
  or absent key. Do not assume the Go-template guard semantics.
- **guardian M**: confirm every `content_components` template consuming `cta_url` (and
  `primary_cta_url`/`secondary_cta_url`) gates on absence — not just the six components
  named in the original round.
- **prior_art_librarian M×2**: cite-and-verify discipline — the `bugs_open/109` citation
  and the caller enumeration were unevidenced. When this lane submits its own round,
  attach the greps/queries, don't assert.
- **architecture (note)**: check whether 109's round recommended consolidating
  contextToMap/contextToInterfaceMap into one shared path; if so the duplicated-rule
  shape is the real defect and a point fix perpetuates it.
- **debug_historian (note, binding on item 4)**: when the cleanup touches
  `page_components` jsonb/text across 7 sites it must follow full needle-gate
  discipline — backup, counted needles, guarded idempotent update, separate
  verify/rollback.

## Decisions

- **D1 (2026-08-06): take the remaining 203 work, not a new bug.** 206/207/205/181/084/
  189/155 all claimed by other live sessions (transcript-verified, not just who-owns).
- **D2 (2026-08-06): the source fix is LIVE — by ancestry, twice over.** `880a405a6` is
  an ancestor of `1e349d046` (197's fix), which was pod-proven on v1.0.1259 on real
  traffic; deploys since (v1.0.1261, confirmed on both pods) build from later HEAD.
  Builds are `git archive` from committed HEAD, forward-only tree, so ancestry ⇒
  inclusion. No independent pod-grep needed for the same binary fact.
- **D3 (2026-08-06): cleanup order is resolver-first, rerender-second.** A bare rerender
  makes every phantom button vanish (correct-or-absent). But 13 CTAs with real intent
  ("Run MatchMatrix" → the tool page) deserve resolution, not deletion. So: re-run
  `resolve_internal_links` per affected page where a plausible target exists, THEN
  rerender. Buttons the resolver still can't place go absent + `unresolved_cta` work
  item — which is the designed behaviour.

## Phasing

- **P0** standing docs + claim (done), verdict + liveness (done).
- **P1** measurements: census re-run; regex-path semantics; template-consumer survey;
  detector cadence. All read-only.
- **P2** class fix for `primary_cta_url`/`secondary_cta_url` defaults (shape depends on
  P1's regex-path answer) + council round. Include the class audit result in the
  submission's grounded_in.
- **P3** shipped-instance cleanup via framework (D3), needle-gate discipline, verify at
  the served page per row.
- **P4** detector follow-up: fix or file separately, by what P1 measures.
