# HANDOFF 2026-08-09c — fact-assignment front (bug 151 / RFC_016): cold-start for a fresh chat

**Supersedes `HANDOFF_2026-08-09b_…`** (which superseded the morning's). Written
~17:00 BST after the v1.0.1276 roll. Bug 215 is **done and live**; the front's
next move is **candidate 1b**, and §3 below is the groundwork for it — read that
before opening `v3_site_actions.go`, it is the expensive part and it is already
paid for.

**This is ONE OF TWO fronts in this directory. Do not confuse them.**
- **This file = the fact-assignment front** (bug 151 candidate 1, RFC_016,
  planner/writer prompts, seeds 327/328/329/330/333).
- **`HANDOFF_2026-08-09_sweep_front_continue_here.md` = the fundamentallyai
  sweep front** — different live thread, same site, same directory.
  ⚠ **Their council round `9da24d85` came back `complete_revise`** (checked
  2026-08-09 ~15:2x). Their handoff's "THE ONE THING OWED" is actionable and
  they may not have seen it. **Read their file before touching the site.**

Site id, needed everywhere: **`199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`**.

## 1. Verified live state (re-checked 2026-08-09 evening)

| thing | state |
|---|---|
| Chassis | **v1.0.1276**, both replicas (`agent-chassis-767d7f5674-5sxdc`, `-sfct5`) |
| **215 dedup fix** | **LIVE + ARTEFACT-VERIFIED.** 4 positives → 1 on both pods; negative control `"collapsed after canonicalization"` (US spelling) → 0. **Never yet exercised in the wild** — no plan write since the roll, so `duplicate_pages_merged` has never been non-zero |
| Migration 327 | applied — `site_plan_sections.assigned_fact_ids` present |
| Seeds 329 / 333 | live in `build-site-planner` |
| Slice B (328/330) | **still held**, both `_HOLD`-named at HEAD |
| Current plan | `8ee5807b`, 71 sections, 17 carrying assignments |
| Site pages | 25 active, all 25 `deployed` |
| HEAD build/tests | `go build ./...` clean; `platform/orchestration/actions` **green** at pinned sha via `git archive`. ⚠ The older handoffs' "HEAD is RED" is **EXPIRED** — do not inherit it |

## 2. Bug 215 — CLOSED IN SUBSTANCE, stays OPEN for the quiet mode

Commit `14b1cff28`, council **APPROVED** round 1 (`8ab18991`, 3 advisory
objections, none high-severity). Full disposition, verdict JSON pinned against
`expires_at`: **`REVIEW_2026-08-09_council_verdict_215_dedup.md`**.

**One item is an OWNER CALL and is not mine to settle** (guardian seat, medium):
when two *composed* pages collide, the fix keeps the richer and logs the other at
Warn — silent partial data loss. The alternative is failing the write, i.e.
today's whole-replan loss. My position is that proceeding is right and the branch
is rare-squared (needs a collision **and** both entries composed; the observed
shape is composed-plus-stub, which loses nothing), but "how much silent loss is
acceptable" belongs to the owning pipeline. **Ask; do not quietly re-decide it.**

Three smaller follow-ups are listed at the end of the REVIEW file: a durable
merge record on the `bugs_open/156` model, `output_contract` parity for
`duplicate_pages_merged`, and a candidate identity-resolution RFC.

**The quiet mode is NOT fixed** — a plan row and a live page still hold two
identities for one page, so a replan still generates phantom 404s and still costs
the sweep front cleanup. It belongs in the reconciler. 20 phantom candidates
fleet-wide.

## 3. Candidate 1b — the groundwork, and why it is smaller than it looked

RFC_016 §3b. (i) prompt: for deployed pages the planner is shown the realised
section list and assigns facts to those names. (ii) Go: Pass B2 carries facts
onto restored sections by component-name match, logging misses durably. (i)
makes (ii)'s match near-total; (ii) alone carries whatever coincides.

**⚠ Line numbers drift daily. Cite by symbol and re-grep.** As of this evening:
`reconcilePlanWithRealised` declared **`:5278`**, called **`:3101`**; Pass B2
**`:5418-5447`**; Pass B **`:5402-5416`**; header **`:5240-5277`**; object-form
normalisation **`:3277-3317`**. File is 5,875 lines. (The 08-09b handoff said
Pass B2 was at `:3031` — already wrong by the afternoon.)

**The loss mechanism, measured not inferred.** Realised sections are a JSON array
of **plain strings** (`jsonb_typeof(sections->0) = string` on every composed live
page: `hero-about`, `hero-services`, `hero`, …). The LLM's sections under seed
333 are **objects carrying `facts`**. Pass B2's restore is `lm["sections"] = rs`
(`:5435`) — the richer shape overwritten wholesale by strings.

**THE FINDING THAT CHANGES THE DESIGN.** I began measuring the blast radius of
emitting object-form sections and found 15+ non-test readers of `["sections"]`,
which looks like a serious widening. **It is not, and the reason is our own
code.** `ValidateSitePlanAction` **already** normalises object-form sections at
`:3277-3317` — objects → `sections` (strings) + page-level `section_facts`
aligned by index, `sawObject`-gated so a page with no object entries is left
byte-identical. That is this front's Slice A work, placed deliberately: *"the
split happens HERE, after the last transformation that can remove an entry."*

**Order verified, not assumed: reconcile runs at `:3101`, normalisation at
`:3277` — 176 lines later, same function.** Therefore:

> **Pass B2 can carry facts onto restored sections in OBJECT form and the
> existing normalisation will split them exactly as it does the LLM's own
> emission. Nothing downstream ever sees object form. The 15+ `["sections"]`
> consumers are NOT in the blast radius.**

**Read these two before editing — both are unestablished and both can bite:**
1. **Pass B** (`:5402-5416`) does `snapped["sections"] = ls`, carrying the LLM's
   sections onto a renamed realised identity. It may already preserve facts, or
   may lose them inside `normaliseRealisedToPlanPage`. **Read that function**;
   decide whether B needs the same treatment as B2.
2. **`sameSectionList`** (`:5667`) compares with `fmt.Sprintf("%v", …)`. Once
   entries are objects it will report "changed" where it previously said "same",
   which only affects logging and the `snappedSections` counter — but **a counter
   that silently changes meaning is how a later measurement goes wrong.** Either
   keep it name-only or rename the counter.

**Build order:** (ii) is independently useful and testable; (i) is a live-config
seed and needs the byte-exact dump procedure (`scratchpad/gen_seed_333.py` is the
worked example — regenerate from a fresh dump; its anchors are asserted 1/1/1 and
will refuse a drifted row). **1b goes IN the Slice B council round (RFC_016 §3b),
never as a bug patch.**

**Mutation-test the carry.** A name-match that silently matches nothing looks
identical to one that works — assert a MISS is logged and counted, not just a hit.

## 4. Do these, in this order

1. **Build candidate 1b (ii)** — Pass B2 fact carry, per §3. Unit-test against
   both shapes; mutation-test the miss path.
2. **Build 1b (i)** — planner prompt shown the realised section list for deployed
   pages. Live row is truth: dump before editing.
3. **Revise + submit the Slice B round.** Draft is
   `COUNCIL_DRAFT_slice_b_2026-08-08.json` with a HOLD note inside — **do not
   submit as-is**; add 1b's edits and a fresh compliance observation (facts must
   SURVIVE onto restored sections). `COUNCIL_SUBMISSION_215_dedup_2026-08-09.json`
   is a worked example of the schema and of a submission that measured its own
   blast radius.
4. **Then**: un-`_HOLD` 328/330 → apply 328 then 330 → rebuild flagged pages →
   census: overlap pairs fall on fundamentallyai, the five fact-blind sites must
   **not** move (the disconfirming half).

Owner-owed, carried forward: the §2 lossy-merge policy call.
Standing/unowned: `bugs_open/214` (imagery `scope_ref` keys off the **raw** LLM
name while pages use the canonical one — two name-spaces on one path).

## 5. Traps (this front's, still live)

- **`orchestration_states` failures are pruned at ~24h; completed rows too.**
  Pin raw evidence the day you cite it, and **never use an error census over that
  table as proof a fix worked** — it reads 0 before and after. Now a LANDMINE.
  `diagnosis_artifacts` also carries `expires_at`: pin council verdicts into the
  repo, do not link them.
- **Line numbers in this directory's docs go stale within hours.** Cite by symbol.
- **A roll is not evidence.** Grep a string your change ADDED **and** one that
  should be absent, same exec, every replica. This change removed no strings, so
  the negative control was a plausible-but-absent spelling.
- **Cite the line where the failing code READS the data** (`WriteSitePlanAction`
  reads `page_plan`/`site_plan` via `extractPagesFromPlan`,
  `site_db_actions.go:749-782` — NOT `validate_plan`).
- **A replan on this site IS a build dispatch** and, until the quiet mode is
  fixed, **still a phantom-page generator**. Co-ordinate with the sweep front.
- **Never `run-migrations.sh --apply`** — the pending list carries other lanes'
  files plus the two `_HOLD` slices.
- **The 097 trigger validates `operation`** against `modify|add|remove|config_change`;
  "create" and compounds like "move + extend" are refused client-side.
- **Don't leave a half-finished edit to `v3_site_actions.go` in the tree.** It is
  5,875 lines, bug-laden, and shared — uncommitted WIP here is what gets swept
  into another session's commit.
- Rollback for the prompt seeds: `agent_definitions_bak_329` / `_bak_333`.

## 6. Commit trail (this front)

`d6e9dcf06` decisions + seed 333 · `47620cb53` bug 215 filed · `c589779a3` Pass
B2 correction · `9b61d04b1` 08-08 handoff · `f58357515` 08-09 re-check ·
**`14b1cff28` the 215 dedup fix + submission** · `90414d055` bug-file + docs ·
`fa483dcdc` landmine footprint fix · `1c854175b` council verdict pinned ·
this file's commit (live verification + 1b groundwork + this handoff).
