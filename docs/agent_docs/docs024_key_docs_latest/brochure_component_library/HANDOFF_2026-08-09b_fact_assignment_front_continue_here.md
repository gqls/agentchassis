# HANDOFF 2026-08-09b — fact-assignment front (bug 151 / RFC_016): cold-start for a fresh chat

**Supersedes `HANDOFF_2026-08-09_fact_assignment_front_continue_here.md`.**
Written ~15:30 BST. Step 1 of that file (bug 215) is **done**; two of its
premises had gone stale and are corrected below. Everything here was checked
today unless marked otherwise.

**This is ONE OF TWO fronts in this directory. Do not confuse them.**
- **This file = the fact-assignment front** (bug 151 candidate 1, RFC_016,
  planner/writer prompts, seeds 327/328/329/330/333).
- **`HANDOFF_2026-08-09_sweep_front_continue_here.md` = the fundamentallyai
  sweep front** — a different live thread, same site, same directory.
  **Read it before touching the site.** ⚠ **Their council round
  `9da24d85` has come back `complete_revise` (checked 15:2x today) — their
  handoff's "THE ONE THING OWED" is now actionable and they may not know.**

Site id, needed everywhere: **`199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`**.

## 1. Verified live state (re-checked 2026-08-09 afternoon)

| thing | state |
|---|---|
| Chassis | **v1.0.1274**, both replicas — unchanged since this morning, so **my 215 fix is NOT in the running image yet** |
| Migration 327 | applied — `site_plan_sections.assigned_fact_ids` present |
| Seeds 329 / 333 | live in `build-site-planner` |
| Slice B (328/330) | **still held**, both `_HOLD`-named at HEAD |
| Current plan | `8ee5807b`, **71 sections, 17 carrying assignments** |
| Site pages | **25 active, all 25 `deployed`** |
| HEAD build | `go build ./...` clean |
| HEAD tests | ⚠ **GREEN — the previous handoff's "HEAD is RED" has EXPIRED.** `TestValidDocSubjectTypes_LockstepWithMigrationCheck` now passes in-tree AND at clean HEAD (`git archive`); another lane landed the migration work. **Do not inherit "HEAD is red" from any older doc** — it is exactly the claim that masks a failure you did cause |

## 2. What changed today: bug 215 crash mode is FIXED

Commit **`14b1cff28`**, **`Council-Submitted: 8ab18991-ee83-4048-8965-4f7990baa188`**
(round was EXECUTING_STEP at `review_guidelines` when this was written — **read
the verdict**; on REVISE resubmit with `RESUBMIT_CORR=8ab18991-…`).

`dedupePlanPageRows` collapses pages that canonicalise to the same name, between
canonicalisation and the transaction. Richer wins, tie keeps first, blank-only
metadata backfill; every merge logged naming both raw spellings. Seven tests,
all three guards mutation-proven. **Inert until the chassis rolls.**

Three findings that are now in `bugs_open/215` and worth knowing here:

1. **Two crash doors** — `site_plan_sections` has UNIQUE
   `(plan_id, page_name, ordering)` as well. Same fix closes both. No unique
   index on `url` (checked, negative).
2. **Three collision families**, and the filed one is the least likely: prefix
   (`tool-`/`guide-`/`game-`), **homepage** (role `index`, or slug `home`/
   `index` under content/landing/empty), section-index (`guides` /
   `guides-index`). The homepage one is the most ordinary emission slip.
3. ⚠ **215's own "how to verify a fix" step cannot come out otherwise.** It
   asked for a `SQLSTATE 23505` census over `orchestration_states.error`.
   **Failed rows are pruned at ~24h** (measured: `FAILED` spans 08-08→08-09
   only, 0 older, against 4,935 rows whose oldest is 07-13). It reads 0 before
   and after, and the 08-08 incident is already invisible to it. Replaced by a
   `duplicate_pages_merged` counter (zero consumers, measured) + log lines. Now
   a LANDMINE on `orchestration_states.error`.

**The quiet mode is NOT fixed and stays open in 215**: a plan row and a live
page holding two identities for one page (the three phantom 404s the sweep front
archived on 08-08). It belongs in the reconciler, not in `WriteSitePlanAction`.
20 phantom candidates fleet-wide today.

## 3. Where this front actually stands — unchanged, and this is still the job

Owner decided all three RFC_016 questions on 2026-08-08: rule **ratified**,
sliced order **approved**, §3a = **option (a)**, which is built, live (seed 333)
and proven at the emission. The owner **approved the v4 writer prompt text**
on 2026-08-09 (`sql/page_content_writer_prompt_v4_2026-08-06.txt`) — the last
human gate on seed 330. **Cleared; do not re-ask.** If that text changes after
this date the approval does not travel with it.

**The blocker is structural and is the whole job now.**
`reconcilePlanWithRealised` Pass B2 (`v3_site_actions.go:3031`, header `:5118`)
restores realised sections over the LLM's for every **deployed** page — correct
and deliberate (bugs 001/037/050/051 lineage), and fact assignments ride *inside*
the LLM's section entries, so they are discarded with them. **All 25 active pages
are deployed**, so candidate 1 reaches none of them today.

**Candidate 1b** (designed, NOT built — RFC_016 §3b): (i) prompt — for deployed
pages the planner is shown the realised section list and assigns facts to those
names; (ii) Go — Pass B2 carries facts onto restored sections by component-name
match, logging misses durably. (i) makes (ii)'s match near-total.

## 4. Do these, in this order

1. **Read the 215 verdict** (`8ab18991-…`). Act on REVISE/REJECTED — the code is
   already on the shared branch, so a verdict is not optional.
2. **Build candidate 1b.** Small, but the Go half touches a bug-laden merge —
   read Pass B/B2's header before editing, and mutation-test the name-match
   carry (a name-match that silently matches nothing looks identical to one that
   works). Goes **in** the Slice B council round (RFC_016 §3b), never as a bug
   patch.
3. **Revise + submit the Slice B round.** Draft is
   `COUNCIL_DRAFT_slice_b_2026-08-08.json` with a HOLD note inside — **do not
   submit as-is**; add 1b's edits and a fresh compliance observation
   (post-215 replan: facts must SURVIVE onto restored sections).
   `COUNCIL_SUBMISSION_215_dedup_2026-08-09.json` is a worked example of the
   schema, incl. the `operation` enum trap (§5).
4. **Then**: un-`_HOLD` 328/330 → apply 328 then 330 → rebuild flagged pages →
   census: overlap pairs fall on fundamentallyai, the five fact-blind sites must
   not move (the disconfirming half).

Standing/unowned: `bugs_open/214` (imagery scope_refs — note page-scope imagery
`scope_ref` keys off the **raw** LLM name while pages use the canonical one; the
two name-spaces coexist on this path and 215's fix did not change that).

## 5. Traps (this front's, still live)

- **Failures in `orchestration_states` are pruned at ~24h; COMPLETED rows too
  (~24h).** Pin raw evidence into a doc the day you cite it, and **never use an
  error census over that table as proof a fix worked** (§2.3).
- **Cite the line where the failing code READS the data**, not the key you had
  open (`WriteSitePlanAction` reads `page_plan`/`site_plan` via
  `extractPagesFromPlan`, `site_db_actions.go:749-782` — NOT `validate_plan`).
- **A replan on this site IS a build dispatch** and, until the 215 fix rolls,
  **still a phantom-page generator**. Co-ordinate with the sweep front first.
- **Never `run-migrations.sh --apply`** — the pending list carries other lanes'
  files plus the two `_HOLD` slices.
- **Live row is truth for prompts.** Dump before editing.
- **The 097 trigger validates `operation` against `modify|add|remove|config_change`.**
  "create" and compound values like "move + extend" are refused client-side —
  cheap to hit, cheap to fix, but it costs a round-trip.
- Rollback for the prompt seeds: `agent_definitions_bak_329` / `_bak_333`.

## 6. Commit trail (this front)

`d6e9dcf06` decisions + seed 333 · `47620cb53` bug 215 filed · `c589779a3` Pass
B2 correction · `9b61d04b1` 08-08 handoff + held draft · `f58357515` 08-09
re-check + handoff · **`14b1cff28` the 215 dedup fix + submission** · this
file's commit (bug-file update, LANDMINE, NOTES, README, this handoff).
