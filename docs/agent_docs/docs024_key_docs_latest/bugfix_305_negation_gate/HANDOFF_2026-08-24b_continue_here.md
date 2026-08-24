# HANDOFF 2026-08-24b — continue here (`bugfix_305_negation_gate`)

**Supersedes** `docs/agent_docs/docs024_key_docs_latest/bugfix_305_negation_gate/HANDOFF_2026-08-24_continue_here.md`
(read that one for how D2/D3 were reasoned; this one for state).

> ## ▶ ONE-LINE STATE
> **The lane is FINISHED. `bugs_open/305` is CLOSED, both owner decisions are RULED, BUILT, LIVE and
> DEMAND-PROVEN on chassis `v1.0.1335`, and all eight council submissions are APPROVED.**
> Nothing is inert. Nothing is blocked on this lane. **There is no outstanding task here** — the
> remaining items belong to other lanes or are watch-only, and §5 says which.

---

## 1. Where the bug file is

**`bugs_closed/305_HANDOFF_2026-08-18_v2_voice_does_not_suppress_define_by_negation.md`**
(closed 2026-08-24; §26–§29 are this week's work). ⚠ It is in **`bugs_closed/`** — several docs still
cite `bugs_open/305`, which is a stale path, not a live bug.

## 2. Verified state `[MEASURED 2026-08-24 18:52Z]`

Running chassis commit `48f55f21834ac3e2d95aa43716f6e63e40ac12ee` (`v1.0.1335`, pods 18:32Z).

| thing | state | evidence |
|---|---|---|
| the gate (detect → select → rewrite → change pages) | **LIVE, proven at the artefact** | 2026-08-22 |
| §26 repair-log accounting | **LIVE + demand-proven** | `no_answer_for_target` fired **43×** |
| §27 output ceiling (`569`, 2000→16000) | **LIVE + demand-proven** | 124 calls at 16000, `cut = 0` |
| §29 invariant correction + test | **LIVE** | council `022169af` |
| `</th` sentence boundary | **LIVE** | council `bccf772a` |
| **D2** tagline correction (`597`) | **LIVE** | council `941ca857` |
| **D3** mild-only budget | **LIVE + demand-proven** | see §3 |

**How to re-prove any of it** — ancestry, not a grep, and it has no shelf life:
```sql
SELECT git_commit FROM service_binary_capabilities
 WHERE service='agent-chassis' AND kind='build' ORDER BY last_seen_at DESC LIMIT 1;
```
then `git merge-base --is-ancestor <your-commit> <that>`, **always with a control** (a commit made
after the build must read NOT live). ⚠ **Do NOT `grep -a <sha> /proc/1/exe`** — `buildinfo.GitCommit`
is ONE string, not an ancestry, so it returns ABSENT for a binary that certainly contains your commit;
two lanes have been burned. ⚠ A **40-zeros control comes back PRESENT** (Go's internal digit table),
so that control can never fail — worse than no control. ⚠ Pod `.status.startTime` dates the **roll**,
not the image. ⚠ The `build provenance` **startup line scrolls** and was already out of `--tail=6000`
twelve minutes after a roll.

## 3. D3 is demand-proven, and here is the falsifiable check

**OWNER RULING 2026-08-24: "`rather than` is a little bit of a tic."** Implemented as *who gets
forgiven*: `page_budget` is a **TOLERANCE, not a repair cap** — it lets a page KEEP N constructions and
repairs the rest. Before the ruling, the survivors were whichever the scanner walked past **first**
(document order, nothing to do with severity), so a page could keep both its `x_not_y` constructions
and spend the gate's effort rewriting two `rather than`s. Now only a **mild** shape may spend it.

`[MEASURED 2026-08-24 18:52Z, 38 markers since the 18:32Z roll]`

- **`within_budget > mild_hits`: 0 violations.** This is the check that could have come out otherwise —
  if a sharp shape ever bought forgiveness, this is non-zero.
- All 38 markers carry the new **`mild_hits`** field, which did not exist before the roll.
- 41 hits, **17 mild / 24 sharp**; 12 forgiven; 27 targets; **0 non-reconciling**.
- Shapes repaired: `rather_than` 12, `x_not_y` 11, `not_x_but_y` 1 — sharp constructions are being
  repaired rather than tolerated.

```sql
-- the D3 invariant, re-runnable
WITH m AS (SELECT (e.val->>'mild_hits')::int mild, (e.val->>'within_budget')::int forgiven
  FROM orchestration_states os, LATERAL jsonb_each(os.collected_data) AS e(key,val)
  WHERE e.key LIKE 'copy\_gate%' AND e.val->>'mild_hits' IS NOT NULL)
SELECT count(*) FILTER (WHERE forgiven > mild) AS violations FROM m;   -- must be 0
```

⚠ **`staccato` is classed sharp by OMISSION** and its frequency is unmeasured. If it turns out common
it needs its own ruling, not a quiet edit — `rather_than` alone is **71%** of rewrites and **43%** of
sections, so moving any shape in or out of the mild set is a large change dressed as one line.

## 4. D2 — done, but the visible effect is deferred

The mandated tagline dropped its negation (`in days, not months` → `in days`), migration `597`.

- ⚠ It was in **five keys across three aspects**, not one. Correcting fewer achieves **nothing
  observable**, because the exemption is computed over the **flattened** brief corpus.
- ⚠ It lives in **`identity.core_value_proposition`** — NOT `content_direction`, which is what every
  earlier doc in this lane said. ⚠ And **not** in the fields named `tagline`; those hold a different
  sentence ("Production-Grade Multi-Agent Systems. Built Right.").
- **The elegant half:** correcting the brief **un-exempts the stored copy**. `adoption-tracker`'s hero
  moved from `exempt:brief_supplied_sentence` to **`REPAIRABLE`**, so the gate repairs it on the next
  ordinary render. **Nothing is hand-edited.**
- ⚠ **The pages will not change until they re-render**, and that site's re-render is blocked on 30+
  unrelated `needs_human_review` items. **Measure the SPEC, not the page, to check `597` landed.**

## 5. What is left — nothing for this lane

1. **`protocol-tracker`** (2 repairable hits): one rerender, **already queued** as a `needs_page` item,
   blocked behind that site's own `claims_unverified` item. **Another lane's.** Holding anything here
   does not advance it.
2. **Watch-only, no action:** `no_answer_for_target` should keep appearing. A permanent absence would
   be worth investigating (omissions ran at 15.3% under the same ceiling before the fix).
3. **Not ours, open and unowned:** the accounting-loop **sibling audit** a council seat asked for —
   `evidence_citations.go`, `revalidate_unverified_claims.go`, and other plan-vs-answer loops that may
   share the "iterate the answer, not the request" shape.
4. **D4** (`negation_density` threshold >12) and **D5** (`brief_supplies_negation` routing) — untouched,
   both blocked on `bugs_open/033` giving that queue a working surface.

## 6. Standing cautions

- **The reconciliation invariant is NOT `targets == rewritten + rejected`.** It is
  `targets == rewritten + rejected − count(reason='no_such_sentence')`, **and only for
  `status='repaired'`**. Five early returns precede the accounting call, so a `repair_unavailable`
  marker accounts for none of its targets **by design**. ⚠ **Expect a small non-zero `over_counted`
  and do NOT chase it**, and **never** close the gap by loosening `matchTarget` — that would splice
  rewrites into copy the model was not describing. `TestReconciliationExcludesHallucinatedReplacements`
  fails if anyone tries.
- **Split any marker census at the change points.** A window spanning a roll is a MIXED batch and reads
  as one rate; splitting it is what revealed that the ceiling accounted for ~60% of what §26 blamed on
  the loop.
- **`\y`, never `\b`, in Postgres.** Cost this lane two wrong figures; a peer lane hit the same thing
  the same week and got a clean zero across every row.
- **`llm_call_log.step_name` is the LOOP-EXPANDED name** (`process_sections_loop_iter_N_rewrite_negations`).
  A bare-name filter returns zero and reads as "never truncated". In `LANDMINES.md`.
- **The step lives in a SUB-WORKFLOW.** Top-level `workflow.steps` queries return 0 rows and read as
  "no agent dispatches this" — use `jsonb_path_query(default_config,'$.**.steps')`.
- **`/tmp` is a 16 GB tmpfs (RAM).** A full one presents as
  `link: mapping output file failed: no space left on device`, which reads like a compiler fault.
  Point `TMPDIR` at disk. **Do not hand-roll `git archive HEAD | tar`** — use
  `scripts/verify-head-builds.sh [--with <file>]`; the canary's targeted extract is `RUNBOOK` §7
  (1.7 MB, not 459 MB, and it `trap`s its own cleanup).
- **A fleet roll KILLS an in-flight council run**, and the corpse is indistinguishable from latency:
  the row sits at `EXECUTING_STEP` for ever. **The tell is `updated_at` vs pod `.status.startTime`** —
  if the pods are younger than the last step transition, it died in the roll. Resubmit.
- **A brief-supplied phrase is exempt BY DESIGN.** Check a marker's `exempt` count before concluding
  the gate missed something.

## 7. Assets

**Migrations owned:** `509`, `517`, `548`, `569`, `597` — all applied and ledger-recorded.
**Councils:** `c48b7612`, `a696e2a3`, `f3046f0c`, `4829bd48`, `022169af`, `bccf772a`, `941ca857`,
`c72ef85c` — **eight submitted, eight APPROVED**.
**Register:** CQ-026 (the family, now carrying the D3 policy and the corrected D2 landmine), CQ-027,
MDL-043. **016b §9:** the loop-over-the-answer + O(N)-ceiling pattern.
**`WRONG_CALLS.md`:** the single-cause attribution entry (~60% of a symptom was a different mechanism).
**Docs:** `PLAN`, `RUNBOOK` (§7 canary, §8 persistence, §9 reconciliation census, §10 ceiling census,
§11 reading this step's config), `NOTES`, `README_where_we_are`, four `SUMMARY_*` files.
