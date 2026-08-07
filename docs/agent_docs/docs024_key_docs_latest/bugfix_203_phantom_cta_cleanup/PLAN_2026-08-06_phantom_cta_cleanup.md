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

- **D4 (2026-08-06): P2 is NOT "delete lines 1138/1140" — that trade is a regression, and
  the council said so before I measured it.** `renderGoStyleSubstitutions` returns the
  literal `{{.field}}` for an absent key, so removing a URL default on the regex path ships
  template syntax inside an `href`. The platform's own recorded preference settles the
  direction — `sanitiseFormAction`'s doc comment: a repair that "makes the form look
  repaired while still losing the message … is worse than the visible breakage: the failure
  stops being detectable from outside." So the door-closing shape is **make the absence
  detectable, not decorative**: have the regex path substitute empty for an absent key (the
  existing `missingBareFields` already logs a blanked `href=` at ERROR as a dead control),
  and only then remove the two URL defaults. `RenderTemplateWithValidation` being dead code
  (F7) makes the stronger option — deleting the regex fallback and failing the render loudly
  — genuinely available, and it is the version that makes the bad state unrepresentable.
  **Both are guarantee changes on shared rendering plumbing: measure how often
  `executeGoTemplate` actually errors first, then submit.** Neither is urgent: F5 shows both
  members inert.
- **D5 (2026-08-06): candidate 3 is re-aimed, not executed.** The detector is not
  under-running (F11). Its output has no handler (F12) and its predicate can only flag an
  anchor when it can already name a better page (F13). Draining 123 items is `bugs_open/083`'s
  class and wants its own lane; the predicate gap wants its own bug. **Not annexed here** —
  this lane would do both badly.
- **D6 (2026-08-06): no cluster dispatch started this session, deliberately.** The
  resolver-then-rerender cleanup (D3) is the right shape and is now *evidenced* rather than
  assumed (F3: the original symptom page cleared itself by exactly that route). But it is a
  multi-step dispatch with two known traps (a hand-made `page_rerender` needs `page_id` in
  the spec AND the column; `save_page_sections` can refuse on the claims guard, so a green
  orchestration is not a written page). Starting it without budget to verify at the served
  artefact would leave the next session a half-finished dispatch and no way to tell whether
  it worked — worse than a clean handoff. Handed off with the worklist pinned.

- **D7 (2026-08-07): the cleanup needs an OWNER DECISION, because every available route is
  either wrong, outward-facing, or new machinery.** D3 ("resolve, then rerender") is
  **retracted** — F18 shows the operation it names does not exist: the resolver is a pure
  function whose only caller is `page-content-writer`, so resolution happens at
  content-writing time and there is no repair-time entry point. With F17 (the correct targets
  exist) and F19 (no misdirected-link repair path), the three real options are:
  1. **Bare `page-rerender`.** Cheapest and safest; makes the pages stop lying immediately.
     But it **deletes** three buttons whose correct destinations exist on the same site. Right
     for the four fabricated "Get Started" heroes, wrong for the three tool CTAs.
  2. **`page-content-writer` with `spec.mode=edit_live`** (`bugs_open/178`'s channel, which
     hands the writer the page's current prose to EDIT rather than regenerate). This is the
     framework's own path and needs no new machinery: the resolver already sits in that
     workflow, so the CTA would come back pointing at the real tool page. **But it is an LLM
     content operation on live customer pages** — outward-facing, costs credits, and can
     change published copy. Not mine to trigger unilaterally, and `edit_live` is itself an
     open bug's new channel whose maturity I have not verified.
  3. **Build the missing capability: relink-an-existing-page.** The robust framework answer,
     and the one that closes the class rather than these 8 rows — but it is new shared
     machinery (council round, concept-register entry), and it is NOT the clean composition I
     first assumed: `load_current_section_content` is bound to `spec.mode=edit_live` and
     yields a writer `section_plan`, not sections in the resolver's shape.
  **Recommendation: (1) for the four "Get Started" blog heroes — nothing there is worth
  preserving — and (2) for the three tool CTAs plus `finetuning.uk/about`, once the owner
  says yes to editing live copy.** Do not mix them in one dispatch.
- **D8 (2026-08-07): F12 is downgraded, against my own earlier note.** "Give the
  `cta_names_unknown_destination` queue a handler" is premature: F20 shows its output is
  dominated by correct contact CTAs flagged by the excluded-area arm, with `affected_url`
  empty. A handler applying `suggested_target` would re-break working buttons at scale.
  **Precision before handler.**

- **D9 (2026-08-07, after the canary): route 2 is PROVEN and REJECTED for this purpose — stop
  the per-page dispatches.** The canary (worklist row 1) completed cleanly and delivered two of
  three goals: the phantom is gone from the served page, and `edit_live` protected the prose
  (both other sections byte-identical). **But it deleted the button instead of re-aiming it**,
  which is what a bare rerender would have done for free. The cause is structural, not a bad
  prompt: `cta_url` is **not** in the hero's `llm_fields`, so no instruction to the writer can
  set it (F23) — and the resolver, which does own it, **assigned the wrong slot**, sending
  "Run the Risk Checker" to `/tools/password-entropy.html` while the *secondary* CTA got the
  risk checker (F22). We were saved from shipping that only because `resolved_data.cta_url` was
  never persisted (F25).
  **So rows 2–4 are NOT dispatched, and route 3 is re-aimed.** The class cannot be fixed
  page-by-page from the outside; it needs (a) the resolver's slot assignment corrected, and
  (b) the `resolved_data` → `content_data` gap understood — noting 129 of 1,247 components DO
  carry `cta_url`, so "it never persists" is false and must not be assumed.
  **This is what the canary was for.** Batching all four would have deleted four buttons and
  taught us none of it.

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
