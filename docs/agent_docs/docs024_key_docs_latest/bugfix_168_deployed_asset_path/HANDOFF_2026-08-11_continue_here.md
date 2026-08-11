# HANDOFF — 2026-08-11 — both revalidators live on v1.0.1284; the cap scare is over and corrected; ONE decision open

**Read this file only.** It supersedes `HANDOFF_2026-08-10_continue_here.md` for state (that file
now opens with two correction banners; its §1–§4 reference material still holds). Working record:
`NOTES_deployed_asset_path.md` (newest at the bottom). Owner's log: `README_where_we_are.md`.

**Nothing is half-applied. Nothing is broken. One thing is deliberately not done** — see §0.

---

> ## ✅ UPDATE, later on 2026-08-11 — `bugs_open/244` IS ALREADY FIXED AND LIVE. DO NOT BUILD IT.
>
> Another session shipped both halves on 08-10 evening, ~2 hours after I filed it: `3d6851d9b`
> (opt-in `cache_control` breakpoint on the shared client + the `llm_call_log` counters, migration
> 376) and `071adc44c` (shared prefix hoisted in all 17 council seats). **I nearly rebuilt it** —
> my grep was a day old on a tree that moves fast, and the owner stopped me.
>
> **Measured live:** full-price input per council round **806,024 → 127,783**, with 973,554 cache
> reads and 93,333 writes ⇒ **~58% cheaper per round, ~69% cheaper per token**; hit rate **157/170
> = 92.4%** on read-eligible seats.
>
> **Two of my recommendations were wrong and are corrected in `244`:** the "≈76%" was optimistic
> (real ~58%), and `ttl: "1h"` was unnecessary — the data refutes my TTL concern, because reads
> keep the entry alive and seats past 5 minutes hit *more* often, not less.
>
> **Still open in `244`: adoption.** Only `council-gate` carries the marker (17 steps, no other
> agent type). That changes §0 below — a resubmitted round now costs ~58% less than the ~1.6M
> figure quoted there.

> ## 🔴 ROUND 4 VERDICT: **REVISE** (2026-08-11 12:51:34Z) — 16 reviewers, 1 abstained, 0 unreadable, NOT truncation-gated
>
> `decided_by`: **gating objection from `prior_art_librarian`**. Fourth revise, and — again — right.
> **Both HIGH objections are answerable with a QUERY, not an argument.** The seat did not claim the
> facts are false; it said it *cannot verify them from the submission*, and it named the exact
> traces it would accept. Both exist. The submission simply never handed it the queries.
>
> ### The two HIGH objections, and the evidence that answers them [MEASURED 2026-08-11]
>
> 1. **"OWNER RULING 2026-08-09 … I have no check tier for markdown files; the only DB-visible
>    trace would be a `diagnosis_artifacts` council_report or a `doc_notes` row … If the ruling is
>    fabricated framing, the gate's legitimacy claim collapses."**
>    → **9 `doc_notes` rows carry it.** Put this in `grounded_in` verbatim:
>    ```sql
>    SELECT count(*) FROM doc_notes WHERE body ILIKE '%register moved, not the page%'
>       OR body ILIKE '%copy-changed gate%'
>       OR (body ILIKE '%owner ruling%' AND body ILIKE '%claims_unverified%');  -- 9
>    ```
> 2. **"Extensive self-cited ROUND 1/2/3/4 history … no visibility into a prior round for THIS
>    submission unless it is in `diagnosis_artifacts` kind='council_report'."**
>    → **It is. 4 rows, this exact correlation** (09-09 ×3 revise, 08-11 revise):
>    ```sql
>    SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
>    WHERE correlation_id='b67eb26a-14ef-45d7-b755-3e489fd57ef0' AND kind='council_report' ORDER BY created_at;
>    ```
>
> **This is the lane's recurring fault for the fourth time: describing the work less carefully than
> it was done.** The fix is not to argue — it is to cite.
>
> ### The other objections worth actioning (none gating, all real)
>
> - **`editquality` MEDIUM — a genuinely missing edit.** `grounded_in` cites
>   `TestUnverifiedClaimsNeverResolvesWhenOnlyTheRegisterMoved` as pinning the gate (and as having
>   caught the zero-date defect), but **no edit adds it** — edit 4 is the only one touching
>   `revalidate_unverified_claims_test.go` and predates the gate. Also LOW: the risks section says
>   *"edit 9"* while the array holds **8**. Same class as round 3's finding.
> - **`guardian` MEDIUM — blast radius belongs in risks, not a code comment.**
>   `loadParkedReviewItems`/`parkedReviewItem` is the shared loader for **all 6** covered item
>   types; a SELECT/Scan mismatch breaks every revalidator, not just this one.
> - **`debug_historian` MEDIUM — and it caught a real defect in THIS session's own verification.**
>   See the pod-population correction below.
> - **`tooling_provenance` MEDIUM** — the producer-count question was answered with ad-hoc grep
>   where `cmd/bundle`/contextkit is the designated tool; that is the same method that produced the
>   false two-producer claim in round 1.
> - `compliance` (approve, 2 non-blocking): the gate is component-granular, not claim-granular —
>   already the §2.1 next job. `architecture`, `constitution`, `mission`, `reuse_agent`,
>   `bug_historian`, `guidelines`, and four guardians: **approve**.
>
> ### ⚠ CORRECTION TO THIS SESSION'S OWN POD-GREPS — `debug_historian` was right
>
> I reported *"both replicas"* on v1.0.1284 and v1.0.1286. **`-l app=agent-chassis` returns 2 pods;
> 26 pods run that image**, across 20+ deployments (`agent-build-dispatch-loop`,
> `agent-color-variable-fixer`, `agent-diagnose-agent`, …). That was a false completeness claim.
>
> **The correct proof is the image DIGEST, not a pod count** — cheaper and stronger than grepping 26:
> ```sql/sh
> kubectl -n ai-persona-system get pods -o jsonpath='{range .items[*]}{.status.containerStatuses[0].imageID}{"\n"}{end}' \
>   | grep agent-chassis | sort | uniq -c     # 21 pods, ONE digest ⇒ provably one binary fleet-wide
> ```
> Result: **one digest** (`sha256:dcd256f9…`), so the grep on any pod is evidence about all of them.
>
> ⚠ **And a NEW trap the runbook now carries: `n=${n:-0}` MASKS A FAILED EXEC AS A ZERO.** A pod
> returned `ownergate=0 cachemarker=0` and read exactly like a stale binary; it was a **completed
> job pod** (`phase Succeeded`) that cannot be exec'd at all. The `${n:-0}` idiom is required for
> `grep -c`'s exit-1-on-zero **and** it silently converts "I could not look" into "it is not there".
> **Always pair it with a per-pod positive control**, and filter to
> `--field-selector=status.phase=Running`.
>
> ### What round 5 needs (small, precise, no code change)
>
> 1. Add the two verification queries above to `grounded_in`.
> 2. File the missing test edit; fix the `edit 9` / 8-edit mismatch.
> 3. Move the shared-loader blast radius into `risks`.
> 4. Replace the "both replicas" deploy claim with the **digest-uniformity** proof.
> 5. Re-run the producer-count check through `cmd/bundle`/contextkit rather than grep.
>
> Then resubmit with `RESUBMIT_CORR=b67eb26a-…`. **Do not change the code** — every seat with
> standing on the design approved it; the objections are all about the plan's evidence, not its
> behaviour.

> ## ⏳ ROUND 4 RESUBMITTED 2026-08-11 12:42:22Z — §0 below is now HISTORY, read this instead
>
> Fired at the owner's instruction, **unchanged**, under `RESUBMIT_CORR=b67eb26a-…` on chassis
> **v1.0.1286**. Orchestration **`ae0915c2-e77a-4d02-94ce-32ced673317a`** — began executing seats
> within seconds (no 29-minute queue).
>
> **Pre-flight done, and worth repeating next time:** pod-grepped both replicas of v1.0.1286
> (`ownergate=1 claims=1 voice=1 cachemarker=1 CONTROL_pos=2 CONTROL_absent=0`) and checked the
> **300s post-restart dispatch window** (2,330s elapsed) — CLAUDE.md's silently-dropped-spawn trap.
> **`cachemarker` is a new standing needle**: it puts the caching fix in the *running binary*
> rather than only in `git log`.
>
> **Get the verdict by correlation, never by `doc_notes ... LIMIT 1`:**
> ```sql
> SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
> WHERE correlation_id='b67eb26a-14ef-45d7-b755-3e489fd57ef0' AND kind='council_report' ORDER BY created_at;
> -- 3 rows = rounds 1-3 (all revise). A 4th row is this round. Objections are in `body`.
> ```
> **If it REVISES again, read the objection before assuming it is procedural** — this council has
> been right three times running, twice about substance. If APPROVED, the trailer is
> `Council-Reviewed: b67eb26a-14ef-45d7-b755-3e489fd57ef0`; **never write that trailer on a verdict
> you have not read.** The docs commit for today already carries `Council-Submitted:` with the same
> correlation, which `098` credits automatically once approved — so no amend is needed.

## 0. THE ONE OPEN ITEM (HISTORICAL — superseded by the block above) — council round 4

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
