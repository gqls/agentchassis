# HANDOFF 2026-08-25 — continue here (`bugs_open/206`)

**Supersedes `HANDOFF_2026-08-24_continue_here.md`** (keep it; accurate for its own day, except
its closure query — see item 1 below, which replaces it).
Read `bugs_open/206` bottom-up first — its last two sections are today, in order.

## State in one paragraph

**The consolidation is COMPLETE.** Both producers of a `needs_page:<name>` item now call
`builderForPageType`; `WriteBuildItemsAction`'s inline copy of the routing maps is deleted; and
`section-index` was added to the shared map in that same commit, which is precisely the condition
council round 4 imposed on 08-24. Commit `efec862f4`, council corr
`b92e624d-15c7-4ef7-a2e5-4a7f41187b38` — **APPROVED at round 1, 13 reviewers, 4 abstained, no
vetoes, every objection `low`.** The 08-24 half remains live on `v1.0.1334`. **The bug stays
OPEN**: today's code is committed but **NOT yet rolled**, and — unchanged from yesterday — nothing
has been proven at the artefact.

## What is DONE (do not redo)

- **The swap** (`efec862f4`): inline maps deleted; `section-index` → `directory-build-handler`;
  the `capability_gap` row's `handler_agent` → EMPTY at this door, matching the sibling door's
  round-2 ruling (`[MEASURED 2026-08-25]` that arm had minted **0 rows ever**, so nothing existing
  changed).
- **Five tests** on a door that had **zero** direct coverage before. Every assertion
  mutation-proven both ways. ⚠ The first version **passed under mutation** for two independent
  reasons and was rewritten — read the file header before editing it, the fixture's shape is
  load-bearing.
- **Two false comments corrected** in `reconcile_site_plan_action.go`, both committed by this lane
  on 08-24 inside an APPROVED round: the "different key namespace" claim (both doors file
  `needs_page:<name>`), and the "open set = 5 statuses, same as the dedup index" claim (6 and 7,
  differing by `unresolved`).
- **The third-copy question ANSWERED** (three seats asked; see item 3).
- **Docs**: `bugs_open/206` (two new sections), RUNBOOK §7/§7a/§8, NOTES, README, `WRONG_CALLS`
  (three entries), `LANDMINES` (one new entry + one correction), **BLD-027** entry and index row
  de-staled.

## What is LEFT — in priority order

### 1. Prove the fix at the artefact. STILL the free one — and USE THE CORRECTED QUERY.

Unchanged in substance: **the proof arrives free on the next greenfield build of any site carrying
an `entity-directory`, `section-index` or `entity-page` page.** No site needs touching, nothing
needs clearing. `[MEASURED 2026-08-25]` it has not happened yet — reconcile rows created after the
08-24 15:39 roll: **0**; sites reconciled after 08-24 12:00: **0**; newest reconcile anywhere is
`agritec.uk` 08-24 11:26, *before* the roll.

⚠ **The 08-24 handoff's closure query CAN BE PASSED BY A HAND REPAIR — do not use it as written.**
`handler_agent` is mutable and re-pointing a parked row is our own documented escape hatch.
`[MEASURED 2026-08-25, live UNION archive, all history]` **all three rows in existence that match
its PASS predicate are hand re-routes** (vetcomparison, created 07-17, updated 08-08 and 08-24).
Run without a domain filter it would have declared the fix proven by rows the replaced hardcode
minted.

**Use RUNBOOK §7 — and it takes TWO gates, not one.**

> **CORRECTED 2026-08-25 (later the same day), by an adversarial review of this file.** This
> paragraph originally said the gate is `spec ? 'page_type'` alone, "which an `UPDATE` of
> `handler_agent` cannot forge". True and **insufficient**: the stamp dates the **row**, not the
> **handler value**, so a stamped mint that the fixed binary *mis*-routed and someone then repaired
> reads `PASS`; and this lane's own promote-in-place recipe forges it outright, because reconcile's
> `capability_gap` spec **already carries `page_type`**. Full working in `bugs_open/206`, the
> "adversarial review" section, item 1.

Gate 1: `swi.spec ? 'page_type'` — the fixed emit stamps it; population currently **empty** (508
reconcile-minted rows, none stamped), so a first hit is necessarily the fix.
Gate 2: `swi.updated_at < swi.created_at + interval '1 second'` — nothing has written to the row
since the mint. Sound because `trg_site_work_items_updated_at` is `BEFORE UPDATE … FOR EACH ROW`.
**It expires**: a legitimate claim bumps `updated_at` too, so read the mint while the row is still
`triaged`. If it has moved, that row is not proof — find a fresher one, don't relax the gate.

Keep the *un-gated* form for detecting FAIL, where the stamp is absent by construction. Two
questions, two instruments.

⚠ **And check `created_by` on whatever the build actually mints before reading an empty result as
FAIL.** §7 filters `created_by='reconcile_site_plan'`. Yesterday's evidence says reconcile is the
greenfield door (garden-tools.uk's 13 items were minted by it at plan time), but
`WriteBuildItemsAction` is wired live too — and if a build mints through *that* door, §7 returns
empty on a **successful** build.

Also note: **today's code is committed, not rolled.** For the mint to show the `section-index`
route, the chassis must carry `efec862f4`. Check with the build stamp, per service:
`git merge-base --is-ancestor efec862f4 <the stamp>`. (`debug_historian` made this point in the
verdict: verify rollout at the running pod, not at green tests.)

### 2. The parked rows — ⛔ **NOTHING ON `garden-tools.uk` IS TO BE CLEARED (OWNER RULING, 2026-08-25)**

> **CORRECTED 2026-08-25, within hours of writing this file.** The paragraph below originally
> listed the three `garden-tools.uk` rows first, as "an operator action", and pointed at the
> re-triage recipe. **That action is forbidden.** The owner **retracted** the parked-row
> authorisation on 2026-08-25 — see
> `CONTRIB_2026-08-25_from_loanzy_lane_the_owner_retracted_the_parked_row_authorisation.md` in this
> directory, filed by the `loanzy_uk_example_site` lane at 11:05, **five minutes before this
> handoff was committed, and I had not read it.** Their words: *"the third option is the live one:
> we record the gap honestly. Nothing is to be cleared on `garden-tools.uk`."*
>
> The CONTRIB was filed precisely because the **08-24** handoff had this same defect — a §1 that
> says "do NOT clear a parked row" and a §2 eleven lines later that spends the authorisation
> anyway. **I reproduced it in the document that supersedes it.** That is the failure this estate
> keeps paying for: a corrected plan surviving, uncorrected, in a second place in the same file.
> Caught by an adversarial review of this lane's own output, not by me.

**`garden-tools.uk` — brand-directory-index / brand-profile / buying-guides-index: DO NOT TOUCH.**
Its whole value is that it is an unassisted greenfield build, and it is the baseline four lanes are
measuring against. Clearing a row there is also **inert** on its own (nothing schedules
`reconcile_site_plan`) and destructive with a re-plan. The gap is recorded honestly instead; that
IS the resolution, not a deferral of one.

**`dartsonline.com` brand-detail and `loanzy.uk` guides-index are NOT covered by that ruling** —
the CONTRIB says so explicitly ("this note says nothing about them"). If clearing those is right on
your own reading, RUNBOOK §4 has the recipe and the `priority ASC` trap. Read §7a first: an
`unresolved` row cannot be freed this way at all.

**What changed today: `section-index` is no longer expected to stay parked** — the hold-out ended,
so `guides-index` now routes like the rest. `entity-page` still files a deferred `capability_gap`
by design.

⚠ **And a re-triage still proves NOTHING about this fix** — setting `handler_agent` by hand only
re-demonstrates that `directory-build-handler` works, known since 08-08. Worse than useless for the
closure test: per item 1, a hand-set handler is now a known false-PASS path.

⚠ **But check §7a first.** A row at `unresolved` is treated as OPEN by `loadOpenPageItems` (so
reconcile skips the page and the new routing never reaches it), is NOT covered by `idx_swi_dedup`,
and is undispatchable at both claim gates. Nothing frees it. `[MEASURED 2026-08-25]` one live
instance — `adversecreditmortgage.co.uk` `blog-index` — which this routing fix **cannot** reach,
and not for any reason to do with routing.

### 3. Follow-ups, with the evidence that sizes them

- **The `unresolved` status-set divergence** (`loadOpenPageItems` vs `idx_swi_dedup`, differing by
  exactly one status, in the damaging direction). `guidelines` asked in the verdict that this be
  filed so it does not rot. Evidence and the one live casualty are in `bugs_open/206` §5(b). **Not
  this lane's to sneak in** — it changes a dedup contract.
- **`needs_directory` is a write-only item_type.** `[MEASURED 2026-08-25]` 0 rows ever minted, 0 Go
  readers outside `builder_routing.go`, 0 live agent configs naming it. Retiring it touches
  `create_tool_cross_link_items.go:263`'s gate. Cosmetic-looking, real blast radius.
- **Residual (b), the larger class** — unchanged from 08-24 and still the better fix: a type mapped
  to a handler that *cannot fill a missing layout*, i.e. everything on bare `page-build-handler`,
  because `ensure_page_section_layout` exists only in `directory-build-handler`'s workflow.
  `blog-post`/`blog-index` casualties measured on four sites. The right shape is making the
  layout-ensuring step reachable from the generic path, **not** routing more types to
  `directory-build-handler`. Its own submission.

- **~~Is there a THIRD routing copy?~~ ANSWERED 2026-08-25 — do not re-run this.** `[MEASURED]`
  six Go sites mint `needs_page:<name>`, not two; three hardcode `page-build-handler`
  (`rerender_page_sections_action.go:1424`, `apply_adoption_plan_action.go:731`,
  `discovery_checks/check_incomplete_page_group.go:202`), one of which (`page-rerender`) is the
  fleet's most active producer. **None is a latent 206**: 26 typed-page rows across those
  producers, **0** carrying `no sections ready to build`; their failures are the owned-page guard
  working as designed. The mechanism: 206 is the *layout-less* case, and rerender/adoption targets
  already have a layout — the two doors that consult the map are the two that mint at *plan* time.
  Commands + caveat in RUNBOOK §8.

## The single most useful thing to know before touching anything

**A column that more than one actor writes is not evidence about any one of them.**
`handler_agent` is written by the producer *and* by every hand repair, retry endpoint and admin
resolve — so it can detect this bug's FAILURE (nobody repairs a row they never noticed) and cannot
confirm its FIX. Find something the repair path cannot forge, and check that thing's population is
empty before you trust your first hit. That is the whole of RUNBOOK §7, and it is now a landmine.

Still true from 08-24, and still worth its line: a census of this class must join
`pages.page_type`; **never** filter `swi.spec->>'page_type'` on pre-fix rows. And **grep by the
SYMPTOM, not the bug number** — this population is named under four numbers (206, 220, 328, closed
187) and `who-owns.py` cannot find it.

## Closure test — when may this move to `bugs_closed/`?

All three, at the artefact, not at a status:

1. A `reconcile_site_plan`-minted row for a typed page carrying `directory-build-handler`,
   **carrying `spec.page_type`** (minted by the fixed code) **and with `updated_at` still at
   `created_at`** (nothing has touched it since). Both gates — the first alone is forgeable two
   ways, see item 1. (RUNBOOK §7.)
2. That page built and serving, verified by `curl`, not by `build_status`.
3. The parked `entity-directory` and `entity-page` rows resolved to their designed outcomes
   (built / deferred gap respectively).

`section-index` staying parked is **no longer** part of the deliberate narrowing — that ended
today. An `unresolved` row still will not move, for the separate reason in §7a.
