# HANDOFF — 2026-08-11 — both revalidators live on v1.0.1284; the cap scare is over and corrected; ONE decision open

**Read this file only.** It supersedes `HANDOFF_2026-08-10_continue_here.md` for state (that file
now opens with two correction banners; its §1–§4 reference material still holds). Working record:
`NOTES_deployed_asset_path.md` (newest at the bottom). Owner's log: `README_where_we_are.md`.

**Nothing is half-applied. Nothing is broken. One thing is deliberately not done** — see §0.

---

## 0. THE ONE OPEN ITEM — council round 4, deliberately NOT resubmitted

Round 4 died on infrastructure, not on content (see §2). The council is available again, so it
**can** be resubmitted unchanged, and the submission file is already correct and committed:

```sh
RESUBMIT_CORR=b67eb26a-14ef-45d7-b755-3e489fd57ef0 \
  ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/bugfix_168_deployed_asset_path/SUBMISSION_2026-08-09_claims_unverified_revalidator.json
```

**Why it has not been fired: cost, one day after the account hit its cap.** [MEASURED] one council
round is **~1.6M input tokens** (15 seats × 106k–118k), and `bugs_open/244` establishes the council
is **87.8% of all fleet input spend**. Rounds 1–3 were each right about something real, so the
review has value — but firing ~1.6M tokens the day after the budget blew is an owner's call, not a
thread's. **Ask before running it.**

Rounds 1–3 read-out with every objection and its answer:
`OBJECTIONS_2026-08-09_claims_unverified_council.md`. Round 4's content answers editquality's
round-3 objection by filing the `parkedReviewItem.CreatedAt` wiring as its own edit.

⚠ **Query the verdict by YOUR correlation, never `doc_notes ... ORDER BY created_at DESC LIMIT 1`**
— that returns whichever lane finished last and reads entirely plausibly until it starts
discussing someone else's bug.
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='b67eb26a-14ef-45d7-b755-3e489fd57ef0' AND kind='council_report' ORDER BY created_at;
-- objections are in `body`, NOT `metadata`
```

## 1. State — verified 2026-08-11 09:45Z, chassis **v1.0.1284**, both replicas

| thing | state |
|---|---|
| `bugs_closed/168` (asset path) | CLOSED, live since v1.0.1229 |
| **`voice_tells` revalidator** | **LIVE + PROVEN** |
| **`claims_unverified` revalidator** | **LIVE + PROVEN** — 8 items closed, all with genuinely-edited copy |
| **Owner's copy-changed gate** | **LIVE, HELD on 8/8, FIRED 0 times.** Still `[UNEXERCISED]` |
| Pod-grep on v1.0.1284 | `ownergate=1 claims=1 voice=1 CONTROL_pos=2 CONTROL_absent=0`, both replicas |
| Latest sweep | **2026-08-11 08:44:19Z**, ran normally |
| Covered types | 6 |
| Council | rounds 1–3 REVISE (all answered); **round 4 dead, awaiting a decision to resubmit** |
| **`bugs_open/244`** | **OPEN** — council prompts 98.6% identical, uncached, uncacheable as ordered |
| API usage cap | **LIFTED** 2026-08-10 18:12Z (was ~3h20m, not the 3 weeks I first claimed) |

### The measurement this lane owes on every visit — re-run it

```sql
SELECT count(*) FILTER (WHERE result #>> '{revalidation,reason}' LIKE '%register moved, not the page%') AS refused_by_gate,
       count(*) FILTER (WHERE result #>> '{revalidation,verdict}'='resolved')    AS resolved,
       count(*) FILTER (WHERE result #>> '{revalidation,verdict}'='still_holds') AS still_holds,
       count(*) FILTER (WHERE result #>> '{revalidation,verdict}'='unknown')     AS unknown
FROM site_work_items WHERE item_type='claims_unverified' AND result ? 'revalidation';

-- the load-bearing invariant: EVERY closure must show the copy moved. Zero `f` rows, always.
SELECT (result #>> '{revalidation,evidence,newest_component_update}')::timestamptz
         > (result #>> '{revalidation,evidence,item_filed_at}')::timestamptz AS copy_actually_changed,
       count(*) FROM site_work_items
WHERE item_type='claims_unverified' AND resolution_path='auto:revalidated' GROUP BY 1;
```
**2026-08-11: `0 | 8 | 19 | 3`, invariant `t` for 8 of 8.** ⚠ **Do not describe the gate as having
prevented anything until `refused_by_gate` is non-zero.** Its four failure modes are pinned by
unit tests, not by observation.

## 2. What happened on 08-10, and the correction that matters more than the incident

Round 4 reached `COMPLETED @ complete_invalid` with `plan_valid: true` — **the submission was fine
and had been accepted**; a seat's LLM call was refused because the Anthropic account had hit its
usage limit. The 400 stated a reset of `2026-09-01`.

> **I then made the mistake worth reading this section for: I reported the stated reset as a
> forecast** — "a 21-day fleet-wide LLM outage" — into five files, and escalated to the owner as
> "plan around three weeks". **It was ~3h20m** (last failure 17:02:12Z, first success 18:12:11Z):
> the owner raised the cap. **The stated reset is the worst case if nobody intervenes.**
> **Verify a lift on the SUCCESS side of `llm_call_log`** — the failures stop appearing either way:
> ```sql
> SELECT date_trunc('hour',created_at), count(*) FILTER (WHERE success) AS ok
> FROM llm_call_log WHERE created_at > now() - interval '24 hours' GROUP BY 1 ORDER BY 1;
> ```

Logged in `WRONG_CALLS.md` together with two siblings from the same session (a pod-log grep whose
window was ~2 minutes wide; a near-duplicate filing another lane had already made). **The
generalised check, asked before writing the claim, not after: "if I were wrong, what would this
measurement look like?" If the answer is "the same", it is not evidence.**

## 3. `bugs_open/244` — the real finding, and it is unfinished work

The budget was exhausted on the **10th of the month**, so raising the cap **moved the wall rather
than removing it**. Measured Aug 1–10:

- fleet **188.1M** input tokens; **`council-gate` 165.2M = 87.8%**, over **209 rounds**
- **790,551 input tokens per round**; 11–15 seats at 106k–118k each
- three seat prompts from one round, byte-wise: **common prefix 20 chars; shared block 268,980
  chars = 98.6%**; seat-specific head only 1,387–5,159 chars
- `grep -rn "cache_control" --include=*.go platform/ pkg/ internal/` → **nothing**
- `platform/aiservice/anthropic.go:103-116` sends **one `user` message, no `system` field**

**Two defects; fixing either alone buys nothing:** caching is off, *and* the shared block sits
**after** the seat header so a prefix cache could never hit. Fix = move the shared block into
`system`, seat instruction last, one `cache_control` breakpoint, **`ttl: "1h"`** (rounds run
**459s mean / 1022s max**, so the 5-minute default expires mid-round). ≈**76% off the council**.

⚠ **`llm_call_log` has no cache columns** — a caching fix is unverifiable until they are added;
that is part of the work. Watch for `cache_read_input_tokens` staying 0, which is what a surviving
silent invalidator looks like. **This change is platform-scope and should go through the council
gate** (which is also the thing it makes cheaper).

## 4. Traps specific to this lane

- ⚠ **A roll is not evidence.** Pod-grep every replica; needles + gotchas are now in the RUNBOOK.
  **Honest limit: `CONTROL_absent` is a fabricated string**, so the check proves grep returns 0 —
  it does **not** prove the binary is newer than v1.0.1279. `pc.locked_at IS NULL` is unusable as a
  negative control (**17 hits** in other `platform/` files).
- ⚠ **The revalidation stamp key is `at`, NOT `checked_at`.** A wrong key returns 0 rows and reads
  exactly like "nothing was scanned".
- ⚠ **`uncovered_backlog` CANNOT confirm an adoption** — flat at 625 before and after a working
  roll. Confirm at the per-type map and at `scanned` decomposed by type.
- ⚠ **A dispatch of this sweep CANNOT be scoped** — both filters read from step config; the live
  `sweep` step has no `input_mapping`, so filters in `input_data` are inert and the run goes
  fleet-wide while looking scoped.
- ⚠ **The council refuses a plan with >8 edits** and takes only `modify|add|remove|config_change`.
- ⚠ **`landmines-sync.py --apply` before `landmines-verify-dispatch.sh` consumes the "new entry"
  status** — CLAUDE.md's own ordering. I hit this on 08-11; the entry's verification never fired.
  Deliberately not re-triggered by hand: the sibling landmine records that `code_symbols` is 100%
  Go while **81% of footprints** (incl. this entry's DB tables and step names) can never resolve,
  so the verdict would be noise. Use `landmines-verify-dispatch.sh` **instead of** `--apply`.
- **Registering a revalidator is the whole wiring** — `coveredItemTypes()` derives from
  `reviewRevalidators`; `TestRevalidatorCoverageIsDeliberate` pins the set deliberately.
- **`platform/orchestration/actions` has FAILING TESTS THAT ARE NOT THIS LANE'S** —
  `TestEveryCheckProducedItemTypeIsClassified` fails at clean HEAD from `e1628f7df`. **Reproduce
  against `git archive HEAD` before attributing any failure to your change.**
- **`/tmp` is a near-full 16G tmpfs** — use `TMPDIR=/home/ant/.cache/buildtmp go build ./...`.
- **Use `git commit -F <file>`**; backticks in `-m` execute.

## 5. Still open from the 08-10 handoff (unchanged, all still valid)

§2.1 tighten the gate from component-granular to **claim-granular** (two seats named it
independently; compare `spec.findings[].matched` against current copy instead of timestamps) ·
§2.2 resolve the **two-standards asymmetry** with `voice_tells` before a seventh type ·
§2.3 pin `ScanDeployedClaims` to its intended callers · §2.4 the **invisible backlog** (467 rows
across six unselected statuses; **do not file a third diagnosis run** — `code_symbols` indexes no
package-level vars, so membership is unreadable by the loop and is first-hand verified at
`work_items_common.go:140-143`) · §2.5 Decision 2's dedup half (**47 pairs / 168 rows**, owner
judgement) · §2.6 more sweep coverage (**use the status filter or the census lies**) · §2.7 the
armed-but-inert cap at `check_image_source_unsatisfiable.go:167`.

## 6. Commits and correlations

`ef80216be` voice_tells · `4030cadb9` claims_unverified + CQ-021 · `6ab7ff594` producer-count
correction · `9a9fef332` the owner's gate · `c70c8e1de` round 4 · **`2d979ddf0` file 244 +
escalation** · **`5c3322aa8` the three-weeks correction**.

| what | id |
|---|---|
| claims_unverified council (r1–r3 REVISE, r4 dead) | `b67eb26a-14ef-45d7-b755-3e489fd57ef0` |
| round 4's orchestration (died at `complete_invalid`) | `2f1b43f6-d92b-49eb-843b-204d0da235fa` |
| voice_tells council (APPROVED r1) | `4d430ca8-7e34-479a-95f3-71fdc12fdef6` |
| selection filter council (APPROVED r1) | `f64da546-e1a4-42e4-98c0-d94cf42af71c` |
