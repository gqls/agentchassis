# HANDOFF 2026-09-03b — bugs_open/437, writer prompt nested item shapes

**COLD-START: read this file, then `bugs_open/437` (§CANDIDATE 1 and §THE 52 BLOCKED KEYS),
then `NOTES_writer_prompt_nested_shapes.md` from the bottom up.**
Supersedes `HANDOFF_2026-09-03_continue_here.md`, which is kept for its account of the
wrong-pod wrong turn.

## The one-line state

**✅ CANDIDATE 1 IS FIXED, LIVE, COUNCIL-APPROVED AND PROVEN AT THE SERVED ARTEFACT.** Nothing
is owed on it. **The bug stays OPEN**, and the reason is now a number rather than a caveat:
**52 of the 73 affected keys (71%) can never recover on their own.**

**There is no open technical question in this lane. There is one open DECISION, and it is
the owner's** — see §The one thing to decide.

## What is settled, with its evidence — do not re-derive any of this

| claim | evidence `[MEASURED 2026-09-03 14:00Z]` |
|---|---|
| the writer is shown the nested shape | **6 of 6** post-roll mechanism-flow writer calls carry `"branches": [{`; **0** carry the old flat exemplar |
| it produces the right shape | **4** pages built **and deployed** with `branches` as arrays, on **3** sites; **0** steps stored a string |
| the census could have failed | negative control in the same query — lendzy `/cant-pay.html` (09-02) still reads 3 steps / 3 **strings** / 0 arrays |
| it works at the served bytes | `advertise.co.uk/uk-advertising-regulation-map.html` → **200, 85,053 B, 7 `branch-label` + 7 `branch-body`**; invented-URL control on the same domain → **404** |
| stuck pages of the `failed` shape self-heal | that page re-minted and built **unattended**, as §Unsticking predicted |
| failures have stopped | last real failure **12:23:58Z**, none since (counted on `orchestration_states`, not on work items — see the trap below) |
| over-production is not happening | **7 of 21** steps carry a populated `branches` array (33%); empty arrays are the omission advice obeyed |
| council | **APPROVED** round 2, corr `6de0f6f2-4f37-492a-9cbd-1ae886311a9b`, 4 advisory objections, none high-severity — **all four answered by census in `bugs_open/437`** |
| the trailer | commit `a0044e73b` carries `Council-Submitted:`; `098` credits it automatically now the verdict is approved. **Nothing owed, do not amend** (forward-only) |

> ⛔ **CORRECTED 2026-09-03 17:10Z, same session. THE SECTION BELOW IS WRONG AND THE ACTION
> IT ASKS FOR MUST NOT BE TAKEN.** The rows block nothing. **20 of the 52 keys closed
> themselves by 17:08Z** (`resolution_path='auto:revalidated'`, 16:08Z, no human action):
> cv1.co.uk and remortgagecalculator.uk are fully drained. `loadOpenPageItems` — which the
> "blocks re-minting" claim cites — governs only `needs_page`, `owned_page_review`,
> `page_build_failed`; **every row counted is `unbuilt_internal_link`**, governed instead by
> `idx_swi_dedup`, which lists `'unresolved'` among the statuses that **free** the slot.
> **Do not clear any rows.** loanzy (14 keys) and farmerinsurance (18) have not drained
> because their target pages have not been rebuilt — the remedy is a **build**, which lets
> the existing drain (`revalidate_unbuilt_link.go`) close the items, as it already did for
> the other two sites. Correction in `bugs_open/437`; full account in `WRONG_CALLS.md`.
> Kept below unedited because the reasoning is the transferable part.

## ~~THE ONE THING TO DECIDE~~ ⛔ WITHDRAWN — and it is not a technical question

`[MEASURED 2026-09-03 14:00Z]` **73 keys** ever carried this error. **52 (71%) are blocked
for ever** by **251 rows** in status `unresolved`, branded `[unresolved after 2 attempts]` —
a state deliberately kept in `loadOpenPageItems`' open set
(`reconcile_site_plan_action.go:751-756`) so it **blocks re-minting** rather than being
re-minted past.

| domain | blocked keys | unresolved rows |
|---|---|---|
| remortgagecalculator.uk | 19 | 64 |
| farmerinsurance.uk | 18 | 62 |
| loanzy.uk | 14 | 113 |
| cv1.co.uk | 1 | 12 |

**Clearing those rows would let 52 pages retry against a writer that now succeeds.** The
precondition §Unsticking set — "a verified build on one page first" — is satisfied four times
over, so this is **no longer blocked on evidence**. It is blocked on a judgement: bulk state
change across four sites belonging to other lanes.

**It has been put to the owner and NOT done.** Do not do it unilaterally. If he says go:
close the branding rows only, never `UPDATE` the terminal `failed` rows (§Unsticking: a fresh
row is the supported path, and there is no re-arm route from `failed` anywhere in the code).
Do one site first and watch a page build before doing the rest.

⚠ **Do not reach for the re-mint-window table as the measure of this.** It counts a rolling
7 days, decays on its own, and cannot tell a success from a failure — remortgagecalculator
went 6 → 2 in three hours while farmerinsurance sat at 21, and advertise's *one-away* count
rose 2 → 4 purely from its four **successful** builds landing as terminal siblings. The
`unresolved` rows are the number that matters and they do not decay.

## The trap this lane hit today — it will catch the next person verifying any fix

A failure census over `site_work_items.error` bucketed by `updated_at` reported **3 failures
after the fix**. There were **0**. The `error` column outlives the failure, and
`trg_site_work_items_updated_at` bumps `updated_at` on every write, so stale text is re-dated
into your post-fix window. **Two of the three rows were `complete`.**

Count the failure's own event instead — `orchestration_states.updated_at` — and project
`status` beside `error` whenever you must query work items. Written up as its own LANDMINE
entry (footprint `site_work_items`; the pre-existing sibling entry carries half of it but is
footprinted on `RegisterActionInputSpec`, so nobody verifying a fix would find it).
RUNBOOK §6 has all four verification queries, including the `jsonb_array_length` trap that
kills the artefact census outright.

## Still open, and unchanged in substance

- **Candidate 2** — no repair path for a type-mismatch refusal: it fails, retries
  identically, goes terminal. Nothing regenerates the one bad field with the error in hand.
  The 52 blocked keys are what this gap costs.
- **Candidate 3** — nothing escalates an active, linked, never-built page. **This is the more
  valuable of the two** and is why these sat for weeks: the failure was loud in the queue and
  invisible everywhere a person looks. Worked example, still stuck and verified at 14:00Z:
  loanzy `/your-rights.html` (`needs_rebuild`), `/guides/index.html`,
  `/guides/tool-loans-consolidation-guide.html` (`planned`), all `deployed_at NULL`.
- **Residual on the migration** (accepted, from `debug_historian`): 724 edited
  `agent_definitions.default_config` with no `snapshot_agent()` backup, and its fail-loud
  guard means it is not safely re-runnable after success. Take the snapshot if you next edit
  that row.

## Other lanes

- **`portfolio_positioning`** — told they can proceed; advertise.co.uk has since built and
  served on its own, so the live test they were holding for has effectively happened.
- **`components` / `bugsweep4`** — the legacy JSON-Schema dialect question is answered (it is
  `bugs_open/240`, a different defect in the same place); their independent census of 5
  JSON-Schema components cross-checks our blast radius at 1. Nothing owed.
- **`bugs_open/453`** — carries a CONTRIB from this lane (`<no value>` in 65% of writer
  prompts, plus a correction to their fix candidate 1). Nothing owed.

## Commits

Previous session: `a0044e73b` fix · `f88789e37` gofmt + register sha · `53b2f46af`
omission-spelling test · `01e98a6d0` NOTES/RUNBOOK council round 1 · `b8d8862c0` 453 CONTRIB ·
`f9550f8ef` re-mint hazard · `58b166955` dialect settlement · `1929b0610` proven at the
artefact.
