# 295 — the owned-page guard in `save_page_sections` kills the work item and files NO review row, while the estate's two other guards on the same predicate both file one

**Filed 2026-08-17** by the `vigilant_designer_offer_analysis` lane, found by reading where the
offer analyser's first drained population actually ended up (NOTES 2026-08-16/17). **OPEN, unowned.**

**One line:** three code paths guard an `owned` page with the same predicate; two of them record
the refusal as an `owned_page_review` item that a human can act on, and the third — the only one
on `page-build-handler`'s route — returns a hard error, so the work item dies `failed` and the
refusal leaves no trace once the orchestration ages out at ~24h.

---

## STATE 2026-08-17 (same day) — **FIXED IN CODE, INERT UNTIL THE NEXT CHASSIS ROLL.** Council submitted, verdict not yet read

Commit `2a5798c4b`. The guard now calls `emitOwnedPageReviewItem(..., "save_page_sections", ...)`
before returning the same error, which is fix candidate 1 below, unchanged in scope.

**The item still fails, deliberately.** The save genuinely did not happen; recording otherwise
would trade a silent failure for a false success. What changes is that the refusal now leaves the
deduped, human-routable row its two siblings leave.

**Test is mutation-proven in BOTH directions, run before submitting** —
`TestSavePageSections_OwnedRefusalEmitsReviewItem`:
- delete the emit call → `ExpectationsWereMet` fails (the row is genuinely asserted, not assumed);
- downgrade the refusal to a silent skip → the must-still-refuse assertion fails.
Both mutations were applied, observed failing, and reverted; full package suite green after.

**Council:** `Council-Submitted: d4f49ea5-fa7d-4996-b04c-8d99d89728f4` (submitted before commit,
per the pre-verdict trailer rule — 098 credits it automatically once approved).
**Still owed: read the verdict and act on a REVISE/REJECTED**, because the code is already on the
shared branch.

⚠ **Do NOT close this file on the commit.** The defect stays reproducible until the fix is in a
running chassis — `make build-*` builds from committed HEAD, but the roll is whole-fleet and the
owner's. Verify at the artefact, not the tag, with the positive AND negative control in
"How to verify a fix" below; a new `owned_page_review` row carrying
`refused_by='save_page_sections'` is the proof, and there were **zero** of those when this was
filed.

**Not addressed here, still open:** fix candidate 3 (routing content findings on owned pages to
`section_edit`, which demonstrably works on them — 18 completes). This fix makes the refusal
visible; it does not make the page get fixed.

## STATE 2026-08-17 (16:15 UTC) — **A FRESH CHASSIS ROLLED AND DOES NOT CARRY THE FIX.** Still OPEN. The defect reproduced on demand and is now quoted a THIRD time

**The roll happened: pods restarted 14:42–14:43 UTC on a new ReplicaSet (`5bd56bdd9b`). The fix is
NOT in the binary.** Proven, not inferred — the startup `build provenance` line had already
scrolled out of range on both pods (which per LANDMINES means "out of range", never "unstamped"),
so this is the binary probe **with both controls in the same exec**:

| probe | reads | meaning |
|---|---|---|
| `(tool/widget-owned); a generic section save` — **semicolon**, introduced ONLY by this fix | **0** | the fix is absent |
| `(tool/widget-owned): a generic section save` — **colon**, the pre-existing error text | **1** | positive control: the grep works and this is the right binary |
| `a generic recomposition` — the sibling guard | **1** | second positive control |
| a fabricated string | **0** | negative control: a 0 here is a real absence |

**Why it did not ship — and this is the actionable part.** The fix has been at HEAD since
**12:12:01 UTC** (`2a5798c4b`, verified `git merge-base --is-ancestor` and by reading the file at
HEAD). But **the deployed tag is `v1.0.1305` and the makefile's `IMAGE_TAG` is ALSO `v1.0.1305`** —
the whole backend fleet is on that one tag, and the pods restarted onto it. CLAUDE.md's build
section states the consequence directly: *"Bump `IMAGE_TAG` (makefile ~line 16) for every build — a
same-tag rebuild ships the node's stale cached binary."*
⚠ **I cannot tell from here WHICH of the two happened** — the image was rebuilt and pushed at the
same tag and the node served its cached layer, or it was never rebuilt. Both are consistent with
everything I can see (pod digest `sha256:f90a7e88…`; I have no earlier digest to compare). **The
mechanism is undetermined; the outcome is not.** Either way the remedy is the same and it is the
owner's: **bump `IMAGE_TAG`, then rebuild and roll.** Re-rolling at `v1.0.1305` cannot help.

### The defect reproduced ON DEMAND, exactly as predicted, on a binary now PROVEN unfixed

Recorded as a falsifiable prediction *before* the item dispatched (NOTES 2026-08-17, prediction 5),
on a named subject, and collected after:

- **Subject:** gamesdesign.co.uk `content_rewrite` on **`tool-ttk-calculator`**
  (`rebuild_policy='owned'`), filed by the sweep-driven B4 run at 12:24.
- **Outcome:** `failed` at **13:02:18**, orchestration `763b227b-4585-4dfb-b3bc-d744214ad7c1`
  (`page-build-handler`, `complete_error`).
- **Cause, quoted:** `step save_sections failed: failed to execute action save_page_sections: page
  tool-ttk-calculator is rebuild_policy=owned (tool/widget-owned): a generic section save would
  clobber it. … Refusing to overwrite.`
- **`owned_page_review` rows with `refused_by='save_page_sections'`: 0.**

**This is the third quoted instance and the first one predicted in advance.** It matters more than
the two from 08-15 for two reasons: the item was filed by the sweep and died without a session
touching it, and the "unfixed binary" premise is now **measured with controls** rather than assumed
from a commit date. **It is also the clean pre-fix baseline** — when the fix does roll, the same
site, same page and same item type is the natural re-test, and the bar is a row appearing where
there are now provably zero.

## COUNCIL VERDICT 2026-08-17 — **APPROVED at round 1**, all 13 reviewers, 4 abstained (corr `d4f49ea5`)

`decided_by: "all reviewers approve"`. The commit's `Council-Submitted:` trailer resolves to this
automatically at `098` report time. **Three objections were worth acting on; all three are answered
below rather than banked.**

**1. `debug_historian`, MEDIUM — the one that mattered.** *The error text and the review item's
`spec.fix` route a human to `apply_section_edit`, and a LANDMINE says BOTH sanctioned actions
refuse when ADDING a section to an owned page.* **Checked, and the objection is half right —
which is the useful half.** LANDMINES `apply_section_edit / ApplySectionEditInputSpec`:
`apply_section_edit` only edits an EXISTING `page_components` row (both edit types target a
`page_component_id`); **there is no add**. So:
- For the two casualties here — `content_rewrite` on `about`, `tone_shift` on `learn-index`, both
  **rewrites of existing components** — the guidance is **correct** and the route works.
- For any finding needing a **NEW** section on an owned page, the guidance is a **dead end**, and
  the landmine names the real route: a direct INSERT plus an **assemble-only** deploy
  (`owned_page_guard.go:29-36`: re-assembly of existing components "is deliberately NOT gated — it
  is how owned pages deploy"; worked precedent `267_tool_guide_intro_recovery_waterfall.sql`).
**Deliberately NOT fixed in this commit.** That text lives in `emitOwnedPageReviewItem`'s shared
`spec.fix` and is read by all four producers; rewriting it inside a point fix is exactly the scope
creep four seats praised this change for avoiding. **Follow-up, unowned:** widen the shared
`spec.fix` to name the add-a-section route. Until then a reviewer acting on one of these rows for
an ADD case will be sent somewhere that refuses.

**2. `bug_historian`, LOW — "no census was run to confirm exactly three call sites".** Fair, and
now run:
```
grep -rn "pageIsOwnedForGuard" --include=*.go platform/ internal/ pkg/ | grep -v _test.go
```
**Exactly TWO call sites invoke the predicate** — `save_page_sections_action.go:186` and
`multipage_actions.go:43` — plus its definition. **Both now emit.** No fourth, uncovered caller
exists, so the fix is complete rather than partial.
> **CORRECTION 2026-08-17 to this file's own wording, and to the council submission.** I wrote
> "three call sites, one predicate". That is loose: there are **three GUARDS on the same ownership
> policy**, but only **two callers of `pageIsOwnedForGuard`**. The third —
> `censusExcludedOwnedPages` — reaches the same policy through the inverse SQL
> (`ownedPageExclusionSQL`), not through the function. The conclusion is unchanged and the fix is
> unchanged; the count was wrong and a reviewer was right to ask for the census rather than accept
> the number. Logged in `WRONG_CALLS.md`.

**3. `guidelines`, flagged for a human — `ON CONFLICT DO NOTHING` vs the `idx_swi_dedup`
DELETE+INSERT contract.** Pre-existing in the shared helper, called identically by the two live
siblings; this change neither creates nor worsens it. Note for whoever routes it: the emit uses
`ON CONFLICT DO NOTHING` with **no conflict target**, so it does not do index inference and cannot
raise the `42P10` that the dedup-index/Go-list drift produces. Worth a separate look at the
helper, not a blocker here — which is what the seat itself concluded.

Two "missing" notes were already satisfied and are recorded so nobody re-checks: the `guardian`
asked whether the helper's signature matches its live definition (it does — the package builds and
the full suite is green), and `editquality` asked what grounds the `item_key` literal
(`owned_page_guard.go:211`, `"owned_page_review:"+pageName`; the test asserts it and passes).

## The evidence

**Two live work items died this way** (offer-analysis, webdesign.co.uk, 2026-08-15). Quoted from
`orchestration_states.collected_data->>'__step_error'` on the orchestrations the items name in
`result.completed_by_orchestration_id`:

- `bbee148d-03e5-46a1-8358-e395e7af39cd` (item `1279d84e`, `content_rewrite`, page `about`):
  `step save_sections failed: failed to execute action save_page_sections: page about is
  rebuild_policy=owned (tool/widget-owned): a generic section save would clobber it. Use
  apply_section_edit for targeted edits or the tool pipeline for rebuilds. Refusing to overwrite.`
- `1bdbbedc-bf29-490f-8b81-b41d21c61aec` (item `9a29ec67`, `tone_shift`, page `learn-index`):
  same message, page `learn-index`.

**The guard is right to refuse.** This is the TL-001 clobber guard (`bugs_open/208`), unified
through `pageIsOwnedForGuard`. Nothing here argues it should save the page. The defect is that it
refuses *without leaving a record*, and that its two siblings do not.

### Three call sites, one predicate, two behaviours

| call site | on `owned` | files `owned_page_review`? |
|---|---|---|
| `owned_page_guard.go:294` (selection census, `get_pages_to_build`) | excludes the page | **YES** |
| `multipage_actions.go:62-84` (`assemble_page`) | returns `skipped:true` + `skip_reason`, **no error** | **YES** (`:75`) |
| `save_page_sections_action.go:186-196` | `return nil, fmt.Errorf(...)` — **hard error** | **NO** |

`emitOwnedPageReviewItem` (`owned_page_guard.go:167-200`) is reviewed, deduped machinery: one
deterministic `item_key` per page, `ON CONFLICT DO NOTHING`, errors swallowed so reporting can
never break the guard. **Its `spec.fix` already names the correct route in prose** — *"edit via
`apply_section_edit`/section-editor, or change `pages.rebuild_policy`…"*. The save path is simply
not wired to it.

### Measured in the live DB, 2026-08-17

- **`owned_page_review` rows by `refused_by`:** 34 `(none)` at `needs_human_review` (reconcile's
  own path, 07-17→08-15) · 1 `get_pages_to_build` (cancelled, the 208 canary) · **0 from
  `assemble_page`, 0 from `save_page_sections`.** The 34 are the **positive control**: the item
  type reaches its terminal human-review state and stays there, so a zero from this path is a
  real absence, not a dead mechanism.
- **172 of 704 pages (24%) are `rebuild_policy='owned'`** — the exposed surface.
- Failed items sitting on owned pages come from several filers, not just this lane:
  `design-audit` (`needs_content_page` ×2, `content_rewrite` ×1), the checker layer with no
  `audit_source` (`literal_markdown` ×8 on 08-14, `placeholder_contact` ×2, `content_rewrite` ×3),
  offer-analysis (×2, quoted above).
  ⚠ **`[UNMEASURED]` for all but the two quoted:** their orchestrations are past the ~24h
  retention, so "failed item on an owned page" is not proof it was this guard. Do not repeat the
  aggregate as if it were.

### Why the normal flow does not catch it

`page-build-handler`'s live workflow has **no `assemble_page` step** (read from
`agent_definitions`, 2026-08-17):
`… plan_sections → check_has_ready_sections → spawn_content_writer → call_content_writer →
check_content_produced → validate_content → save_sections → update_status → …`

So the sibling that *does* file the review item is never on this route, and
`SavePageSectionsAction`'s upstream-skip honour path (keyed to `ownedPageSkipReasonPrefix` by
commit `6a9d85777`) can never fire either — there is no upstream skip to honour. The direct
ownership check was written as a backstop; on this route it is the **only** guard, and it is the
one that reports nothing.

**The correct destination already exists in the same workflow.** `validate_content`'s
`error_step` is `mark_needs_review`; `save_sections`'s `error_step` is `mark_item_failed`. A
content-validation failure gets a human; an owned-page refusal gets a silent `failed`.

## Fix candidates, ordered by what closes the door

1. **(Preferred, code, needs a roll) Call `emitOwnedPageReviewItem` in the save-path guard**,
   matching its two siblings, then keep returning the error. The item still fails — honest, the
   rewrite genuinely did not happen — but the refusal becomes a deduped, durable,
   human-routable row naming `apply_section_edit`. Closes it for **every** filer (design-audit,
   the checker layer, any future auditor), not just for offer-analysis, and needs no per-agent
   config. ~6 lines at the existing call-site pattern; `siteID` is already in scope for the
   error path via the same `CollectedData` lookup the sibling uses.
2. **(Config-only, live immediately, WRONG SCOPE) Point `save_sections.error_step` at
   `mark_needs_review`.** No roll, but it sends *every* save failure to human review, including
   transient ones — it cannot tell an ownership refusal from a database blip. Records the symptom
   at the cost of a mislabelled queue. Use only as a stopgap, and say so.
3. **(Upstream, larger) Route content findings on owned pages to `section_edit` at triage.**
   `section_edit` **completes** on owned pages — 18 of them — so the working route exists; nothing
   puts a content finding onto it. This is the real repair rather than the honest refusal, but it
   is a routing-policy change with its own design, and 1 should land first regardless.

## How to verify a fix

Positive and negative control in one run, because a zero here is otherwise ambiguous:
- Dispatch a `content_rewrite` at a **known owned** page → expect a NEW `owned_page_review` row
  with `refused_by='save_page_sections'` and the item terminal.
- Dispatch one at a **known generic** page → expect NO review row and a real save.
Running only the first cannot distinguish "the fix works" from "the emit fires unconditionally".

## Verification basis (owner ruling 2026-07-31)

**Not** put through the `090` diagnosis loop; the substitute, stated plainly: all three call
sites and both governing commits (`cb7b4d759`, `6a9d85777`) read first-hand at file:line; the
absent `assemble_page` step read from the live `agent_definitions` row, not from a repo seed; and
the data claim carries its own positive control (34 rows prove the item type reaches
`needs_human_review`, so 0 from this path is an absence rather than a dead mechanism). The two
failures are quoted verbatim from the orchestrations the items themselves name.

## Relates to

`bugs_open/208` (the guard this extends — its lane last touched it 08-08; the selection and
assemble paths it fixed are the two that DO file the review row) · `bugs_open/210` (a page-build
failure stamped deployed — same family: the build path's failure reporting) ·
`features_open/030` + BIZ-032 (the offer analyser, whose findings were the two casualties) ·
`bugs_open/115` (findings that terminate nowhere — this is the same harm reached by a different
route).
