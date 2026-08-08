# HANDOFF — 117 chrome staleness reference — cold start for a fresh chat

**Written** 2026-08-08. **State: RESEARCH COMPLETE, NO CODE CHANGED, NO PLAN YET.**
Nothing under `platform/` or `internal/` has been touched by this lane. There is
nothing live to verify and nothing to roll back.

## Read these first, in this order

1. `bugs_open/117_HANDOFF_2026-07-27_site_chrome_is_a_stored_artefact_no_page_rerender_regenerates.md`
2. `NOTES_chrome_staleness_reference.md` (this dir) — the measurements + 3 missteps
3. `RUNBOOK_chrome_staleness_reference.md` (this dir) — every query, with its gotcha
4. `docs/agent_docs/docs024_key_docs_latest/016b_debugging_guide_8_consolidated.md`
   line ~598, §9 "Light site renders dark chrome" — prior art on the same mechanism

## The one-paragraph version

Chrome (head/header/footer) is pre-rendered into `site_components.rendered_html`
and served verbatim by `assemblePage`, so no page re-render regenerates it. That
is 117 as filed and it is still true. **What is new:** a drift detector for this
already exists, is live, and compares the wrong two timestamps — it measures
stored chrome against the site's most recent *page content* render, which is not
what chrome is built from. Measured over all 53 chrome rows with provenance:
**36 of 39 firings are unrelated to chrome drift, and the one genuinely stale row
(oufe.com/footer) does not fire.** Wrong in both directions, and its output is
believed and drains real rebuild capacity.

## Verified facts, with their evidence

| claim | evidence | marker |
|---|---|---|
| chrome is served from a frozen string | `getSiteComponents` `rerender_single_page_action.go:662`; `assemblePage` :352 | [MEASURED] re-confirmed after the 2026-08-08 build |
| the detector's reference is `MAX(page_components.updated_at)` | `check_integrity.go:320-331` | [MEASURED] |
| 36 false positives / 1 false negative / 3 coincidental | RUNBOOK R1 cross-tab, 53 rows | [MEASURED] 2026-08-07 |
| the check is live and draining, not dormant | `site_work_items` `item_key LIKE 'stale\_sc\_%'`: 7 complete per slot, latest 2026-08-06 | [MEASURED] |
| 17 of 19 `head` rows point at an `is_active=false` component | RUNBOOK R2 | [MEASURED] — **already covered** by the sibling `deactivated_site_components` check; do not duplicate |
| 3 rows have `component_id IS NULL` (loanandmortgagecalculator.co.uk) | RUNBOOK R2 | [MEASURED] — no provenance at all; decide what the fix does with these |
| `fix_harcoded_colours_action.go:180` writes `html_template` with **no** `updated_at`, and targets chrome | that line + its selection query ~:145-160 | [MEASURED] — the **only** such writer of ~9 |
| a re-render would actually change those 4 footers' output | — | **[NOT ESTABLISHED]** — no class literal in `footer-theme-chrome` splits stored footers by the 08-02 change; see NOTES misstep 1 |
| the 090 diagnosis confirmed any of this | — | **[UNVERIFIED]** — run completed (5 bundles), verdict not retrievable; see below |

## The 090 diagnosis — ran, completed, verdict NOT retrieved

- intake `9366c2c5-412e-498c-9431-c45a37dd8411`
- **run `0001d9ee-c0ad-4ef2-9304-57e1b4757ec8`** ← the key artifacts are under
- item `complete` 2026-08-07 08:54:52Z; 5 `kind='bundle'` rows in
  `diagnosis_artifacts`; final metadata `truncated:false, symbol_count:12,
  symbols_unreadable:1`

No `doc_notes` row joins to either correlation and no bundle carries a `VERDICT`
line — the known "verdict computed then thrown away" defect (commit `0252b3cae`).
**Do not write that the loop confirmed this.** The case rests on first-hand
measurement, which is the owner ruling's named escape hatch, declared here.
If you want a verdict, re-run 090 rather than inferring one.

## Where the work stopped

A `Plan` agent on **fable** was briefed with everything above and cut off by a
session limit partway through reading `render_site_components_action.go`. Its
brief is worth reconstructing — the six questions it was asked are the plan's
table of contents:

1. **Name the real render inputs, per slot, with file:line.** Read
   `render_site_components_action.go` and say what `head`, `header` and `footer`
   are each actually rendered from. They may differ (head pulls brand assets via
   `derive_brand_head_assets_action.go`). **The design depends on this and it is
   the one thing not yet done.**
2. **Pick the core mechanism and justify it.** Candidates:
   - (a) fix the staleness *reference* — compare against what chrome is built from;
   - (b) **stamp a render-input fingerprint** on the `site_components` row at
     render time and detect drift by recomputing it — immune to writers that
     forget `updated_at`, to timestamp bumps that change nothing, and to
     unrelated churn. **The evidence points here** (see D4 in PLAN);
   - (c) invalidate-on-write at every chrome-template writer;
   - (d) render chrome at assemble time (117's candidate 3, worst blast radius).
   Say what each rejected option fails to make *unrepresentable*. If (b): where
   does the fingerprint live — new column, or a key in the existing
   `site_components.content_data` jsonb (no migration)? What exactly is hashed?
3. **Decide the fate of the existing `stale_site_components` check.** It is live
   and firing on noise. Narrowing a detector can make it inert — a known trap
   here. State the **induction test** that proves it can still fire, because a
   detector's 0 findings has two causes with opposite fixes.
4. **Blast radius by measurement, not argument** — give queries that could come
   out either way.
5. **Verification with a positive AND a negative control** (a string the change
   REMOVES, expect 0). The natural check — change component, re-render, curl —
   returns the old output and reads as "my edit was wrong". A roll is not
   evidence a fix shipped.
6. **Scope call:** council gate or RFC? Correcting a predicate changes which rows
   a check emits, not the seam's contract → council. But a new provenance
   *column* on a shared table may cross into architecture scope. PLAN D5 marks
   this `[ASSUMED]`.

## Constraints the design must honour (do not rediscover these)

- Work-item dedup is `UNIQUE (site_id, item_key)` over non-terminal statuses. Two
  producers sharing a key silently swallow one another — see the
  `deactivated_pin_` comment in `check_integrity.go`. Do not create the shape
  that `bugs_open/213` is open about.
- Route only at a handler that can satisfy the item, or it ages to `unresolved`.
- **Reuse first:** `RequestNavRebuild` (`nav_rebuild_request.go`) is the existing
  council-APPROVED way to ask for a chrome rebuild (`nav_drift` → `nav-updater`,
  whose workflow is `populate_nav_tables → render_site_components →
  create_rerender_items → get_pages_for_rerender`). Only 2 callers today. Read
  its header comments in full — it documents the delivery risk and the dedup.
  `sha256hex` already exists in `code_symbols_actions.go`. A shared
  `set_updated_at()` trigger already exists in the DB (check `pg_proc`/`pg_trigger`
  before adding one — 016b records that reuse-gate).
- Chrome **locks** must be respected: `site_component_lock_guard.go`,
  `chromeFixLockSkip` in `fix_component_template_action.go`. A rebuild must not
  stomp a locked slot.
- A platform seam must be registered in
  `docs/agent_docs/docs026_concept_register/register/` **in the same commit that
  ships it** (owner ruling 2026-07-28, condition 2 — still binding).
- Owner ruling 2026-08-02 (RFC_010 #2): new authority on a shared seam ships as
  an **opt-in field with the unsafe default OFF**, not a documented contract.
- Go changes are inert until an image is rebuilt and rolled; DB config is live
  immediately.

## Ownership at time of writing

`scripts/who-owns.py 117` → no owning workstream. Symbol-level transcript grep
(`check_integrity`, `StaleSiteComponents`, `render_site_components_action`)
across all sessions active in the previous 5 hours → **only this session**.
**Re-run both before you start** (RUNBOOK R8) — both checks lag, and this is a
shared tree.

## First three moves for the next session

1. Re-run RUNBOOK R8 (ownership) and R1 (the cross-tab). If R1's four cells have
   moved materially, say so in NOTES before designing against the old numbers.
2. Read `render_site_components_action.go` and answer question 1 above — the
   per-slot render inputs. Write the answer into PLAN as D6.
3. Then design the fingerprint and put it through the council gate. Commit with
   `Council-Submitted: <corr>` rather than holding the code for a verdict.
