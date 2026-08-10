# HANDOFF — 2026-08-10 — both revalidators LIVE and closing; council round 4 in flight; the owner's gate is proven present but has never fired

**Read this file only.** It supersedes `HANDOFF_2026-08-09_continue_here.md` for state; that file
keeps its reasoning and its corrections to the 08-08b handoff, all of which still hold. Working
record: `NOTES_deployed_asset_path.md` (newest at the bottom). Milestone read-out:
`SUMMARY_2026-08-10_deployed_asset_path.md`.

**Nothing is half-applied and nothing is blocked on a decision.** The owner has ruled; the code is
live and demonstrably working. One council round is running and one measurement is owed.

---

## 0. THE OPEN ITEMS

> ## ⛔ CORRECTED 2026-08-10 (later session) — READ THIS BEFORE §0a AND §0b
>
> **§0a is WRONG: round 4 is not in flight, it is DEAD.** Its orchestration
> (`2f1b43f6-d92b-49eb-843b-204d0da235fa`) reached `COMPLETED @ complete_invalid` at 14:51:49Z
> with `plan_valid: true` — the submission was accepted; a seat's LLM call was refused.
> `__step_error` carries the only tell: **the Anthropic account hit its usage limit, stated reset
> `2026-09-01 00:00 UTC`.** Fleet-wide: last successful LLM call **14:51:45Z**, nothing since.
> **DO NOT RESUBMIT** (`LANDMINES.md`, the usage-limit entry — resubmitting burns a round against
> a wall). **The council gate and the `090` diagnosis loop are both unavailable until the owner
> raises the limit**; that is escalated and is not a thread's decision.
>
> **§0b is still exactly right and was re-run: `refused_by_owner_gate = 0`** (8 resolved / 19
> still_holds / 3 unknown), invariant clean at **8 of 8**. The gate stays `[UNEXERCISED]`.
>
> **The sweep itself is unaffected** — no LLM step — so it keeps closing items through the outage.
>
> Chasing the cause produced **`bugs_open/244`**: the council gate is **87.8% of fleet input
> tokens**, sends **98.6%-byte-identical** prompts to 11–15 seats per round, uses no prompt
> caching, and orders the prompt so caching could not fire if enabled. ~76% reduction available.
> Working: `NOTES_deployed_asset_path.md` (bottom) and `README_where_we_are.md` (bottom).

### 0a. ⏳ Council round 4 — `b67eb26a-14ef-45d7-b755-3e489fd57ef0`, run `a7b35edc-4094-42b4-a2e7-a863de831e6b`

Rounds 1, 2 and 3 all came back **REVISE** and **each was right**. Full read-out with every answer
and its check: **`OBJECTIONS_2026-08-09_claims_unverified_council.md`**.

| round | gated by | what it caught | answered how |
|---|---|---|---|
| 1 | `editquality` | I claimed **two converging producers**; there is **one**, and I had invoked an owner ruling that only applies to two | corrected in all five places, visibly |
| 2 | `compliance` | the HITL policy question is the **owner's**, and treating it as a notification rather than a gate was wrong | escalated → **owner ruled**, gate built |
| 3 | `editquality` | the plan described the `CreatedAt` wiring only inside another edit's rationale | **filed as its own edit** (round 4) |

```sql
-- decisions, one row per round, oldest first
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='b67eb26a-14ef-45d7-b755-3e489fd57ef0' AND kind='council_report' ORDER BY created_at;
-- THE OBJECTIONS ARE IN `body`, NOT `metadata` (which holds only decision/reviewers/abstained)
SELECT body FROM diagnosis_artifacts
WHERE correlation_id='b67eb26a-14ef-45d7-b755-3e489fd57ef0' AND kind='council_report'
ORDER BY created_at DESC LIMIT 1;
```

> ⚠ **NOT `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC
> LIMIT 1`** — CLAUDE.md documents that command and on this tree it returns whichever lane finished
> last. It handed me `bugs_open/228`'s verdict, which read as entirely plausible until it started
> discussing contact forms. **Always query by your own correlation.**

**If round 4 comes back REVISE again, read the objection before assuming it is procedural.** This
council has been right three times running, twice about substance.

### 0b. 👀 THE ONE MEASUREMENT OWED — has the owner's gate ever actually fired?

**As of 2026-08-10 14:39Z: no. `refused_by_owner_gate = 0`.** The gate is proven present at the
binary and proven correct on all 8 closures, and it has **blocked nothing**. Re-run this after the
next scheduled sweep (~08:37Z) and after any day with new closures:

```sql
-- has the gate refused anything yet? (0 = still unexercised, and say so)
SELECT count(*) FILTER (WHERE result #>> '{revalidation,reason}' LIKE '%register moved, not the page%') AS refused_by_gate,
       count(*) FILTER (WHERE result #>> '{revalidation,verdict}'='resolved')    AS resolved,
       count(*) FILTER (WHERE result #>> '{revalidation,verdict}'='still_holds') AS still_holds,
       count(*) FILTER (WHERE result #>> '{revalidation,verdict}'='unknown')     AS unknown
FROM site_work_items WHERE item_type='claims_unverified' AND result ? 'revalidation';

-- and the load-bearing invariant: EVERY closure must show the copy moved. Zero `f` rows, always.
SELECT (result #>> '{revalidation,evidence,newest_component_update}')::timestamptz
         > (result #>> '{revalidation,evidence,item_filed_at}')::timestamptz AS copy_actually_changed,
       count(*) FROM site_work_items
WHERE item_type='claims_unverified' AND resolution_path='auto:revalidated' GROUP BY 1;
```

⚠ **Do not describe the gate as having prevented anything until that first count is non-zero.**
It is `[UNEXERCISED]`, not proven inert and not proven to bite — its four failure modes are pinned
by unit tests, not by observation.

---

## 1. State — verified 2026-08-10 14:39Z, chassis `v1.0.1279`, both replicas

| thing | state |
|---|---|
| `bugs_closed/168` (asset path) | CLOSED, live since `v1.0.1229` |
| **`voice_tells` revalidator** | **LIVE + PROVEN** — closed its first item unattended 08-09 08:38:54Z |
| **`claims_unverified` revalidator** | **LIVE + PROVEN** — 8 items closed 08-10, all with genuinely-edited copy |
| **Owner's copy-changed gate** | **LIVE, HELD on 8/8 closures, FIRED 0 times.** §0b |
| Pod-grep on `v1.0.1279` | `claims=1 ownergate=1 voice=1 control=2`, both replicas (baseline `0/0/0` on `v1.0.1270`) |
| Latest sweep (08-10 08:44Z) | `scanned 243 · cap_binding false · resolved 37` |
| Covered types | **6** — unresolved_cta, required_fields_missing, needs_section_data, needs_page, voice_tells, claims_unverified |
| Council | rounds 1–3 REVISE, all answered; **round 4 in flight** |
| Concept register | CQ-021 added (index 1,800 → 1,801); CQ-020 updated to LIVE+proven |
| Consumers told | `bugs_open/033` + `/083` carry dated notices for BOTH types |
| Owner's follow-on idea | `features_open/031_FEATURE_pages_carry_a_last_checked_date.md` — captured, not built |
| Decision 2's dedup half | OPEN, blocked on an owner judgement. **47 pairs / 168 rows, NOT drifting up.** §3.1 |
| Status-ceiling diagnosis | 2 runs, both UNVERIFIABLE on tooling; mechanism confirmed, membership first-hand. §3.3 |

## 2. What is left, in order of value

### 2.1 Tighten the gate from COMPONENT-granular to CLAIM-granular ← the best next job

Two seats named the same real limitation, independently, and **it is not answered**:

> `compliance` (HIGH): *"the underlying timestamp is component-granular, not claim-granular — any
> unrelated edit to the same slot after filing satisfies the gate even if the actual disputed
> sentence was untouched."*
> `debug_historian` (MEDIUM): *"`updated_at` can be bumped by unrelated rebuilds/rerenders that
> touch content_data without changing the claim text."*

**Tested, and the obvious failure did not occur:** if a bulk rerender were satisfying the gate, the
closures would share a timestamp. They span **ten days across several sites** (`2026-07-31 10:10`,
`08-02 10:24`, `08-04 20:54`, `08-05 14:06`, `08-09 01:02`, `01:07`, `16:46`, `19:35`). Two rows
five minutes apart on 08-09 share a filing instant and could be one wave — 2 of 8.

**The limitation stands regardless.** The shape of a fix: compare the specific flagged text
(`spec.findings[].matched` / `snippet`) against the current component, rather than comparing
timestamps. The finding already stores what it objected to; nothing currently reads it back.

> ⚠ **Read this first** — a neighbouring lane recorded, on 2026-07-31, exactly why this matters:
> ```sql
> SELECT body FROM doc_notes WHERE categories ? 'landmine' AND subject_key LIKE '%evidence_base%';
> ```
> *"A registered fact makes a GREEN claims gate meaningless as evidence of truth … the register is
> also the **writer whitelist**, so a false fact is **SELF-RATIFYING**: the platform instructs the
> writer to state it, then vouches for it."* Live case: `bugs_open/161` — gamesdesign.co.uk asserts
> "10,000 Monte Carlo trials per query" against tool JavaScript containing no randomness, and every
> gate passes it, correctly.

### 2.2 Resolve the two-standards asymmetry before a seventh type arrives

`voice_tells` has the identical moving-standard hole (loosen a site's `voice_gate` and findings
retract with the copy unchanged) and is **deliberately not gated** — it is live, council-approved,
and its surface is style rather than truth, which is the distinction the seats themselves drew.

But `architecture` (LOW) and `bug_historian` (MEDIUM) both flagged the consequence: **one shared
registry now has two different closing standards**, which is 016b §9's recurring shape *"one call
site of a shared judgement gets the rigorous fix; the sibling stays heuristic"*. Disclosed is not
resolved. Decide it before adding a seventh type.

### 2.3 Pin the newly-exported scan to its intended callers

`guardian` (LOW): nothing stops a future caller reusing `ScanDeployedClaims` against a different
predicate and drifting from the emit side — *the exact property the extraction exists to preserve*.
A `*_coverage_test.go`-style assertion on its call sites would close it.

### 2.4 The invisible backlog — mechanism confirmed, membership unreadable by the loop

The sweep only ever selects `needs_human_review` and `unresolved`. **467 rows sit across six other
statuses**, and `reportUncoveredBacklog` filters by the same list, so the coverage report cannot see
them either. It reports 625 when the parked population is nearer 1,100 — and it does not only omit
whole types: `undeployed_asset` is **listed at 50 while 46 more of its rows are invisible**.

Two diagnosis runs (`f3d18013-…`, `a174b184-…`), both **UNVERIFIABLE**. Run 2 confirmed the
mechanism — *"both `loadParkedReviewItems` and `reportUncoveredBacklog` demonstrably share this same
status-list filter (static fact, confirmed)"* — and could not read the list's **membership**,
because **`code_symbols` indexes no package-level vars or consts at all** (`func` 3,592 · `method`
1,114 · `struct` 973 · `alias` 40 · `interface` 36, and nothing else). Membership is first-hand
verified: `work_items_common.go:140-143` is literally `{"needs_human_review", "unresolved"}`.

**Do not file a third run into the same wall.** Widening that list is architecture-scope regardless:
it is interpolated in three places, and per its own comment widening the selection alone selects
rows the write-time CAS guards then silently refuse to update.

### 2.5 Decision 2's dedup half — still blocked on the owner

**47 colliding `(site_id, item_key)` pairs across 168 rows** (2026-08-09, PROPOSED predicate).
Needs a *"which copy do I keep, and does discarding the rest lose a true finding?"* judgement.

⚠ **READ THE INDEX, DO NOT RECONSTRUCT IT** — `SELECT pg_get_indexdef(oid) FROM pg_class WHERE
relname='idx_swi_dedup'`. Writing the exclusion list from memory produced 75/227.
⚠ **The "drifts upward ~2 pairs/day" line in every older doc is WITHDRAWN** — four points bounce
(48/135, 53/180, 55/184, 47/168); it fell 8 pairs in 14 hours.

### 2.6 More sweep coverage — the census needs a status filter or it lies

Zero-closer types that are actually **selectable**, measured 2026-08-09: `lock_blocked_change` 23 ·
`image_source_unsatisfiable` 18 (ties to the §2.7 tripwire) · `needs_sprite_css` 10 ·
`dead_control` 8 · `stale_evidence` 5. `cta_names_unknown_destination` (123) is **owned by
`bugs_open/023`** and a test asserts it stays out of the registry.

```sql
SELECT item_type,
       count(*) FILTER (WHERE status IN ('needs_human_review','unresolved'))          AS selectable,
       count(*) FILTER (WHERE status IN ('complete','verified'))                      AS closed_ever,
       count(*) FILTER (WHERE status IN ('complete','verified') AND result ? 'deploy_result') AS closed_by_deploy
FROM site_work_items GROUP BY 1
HAVING count(*) FILTER (WHERE status IN ('needs_human_review','unresolved')) > 0 ORDER BY 2 DESC;
```
`closed_ever > 0` ⇒ something already drains it. A `deploy_result` block ⇒ a real fix pipeline owns
it and retraction is the wrong tool (that disqualified `content_rewrite`). **Omit the status filter
and you will nominate types the sweep structurally cannot see** — that is how `image_url_404`
(26 open, 0 closed ever, flag-only, DB-answerable) looked obvious and turned out unreachable.

### 2.7 The armed-but-inert cap — a tripwire, not a task

`discovery_checks/check_image_source_unsatisfiable.go:167` is still `return result, nil` inside its
per-pass cap and still populates `Resolved` **0** times. Correct today, untouched. The commit that
adopts the retraction seam there is the commit that must change it to `break`.

## 3. Traps specific to this lane

- ⚠ **`uncovered_backlog` CANNOT confirm an adoption.** It was flat at **625 before and after** the
  `voice_tells` roll that worked perfectly — the type left `uncovered_types` (25 → absent) while
  others grew by exactly 25. Confirm at the **per-type map** and at `scanned` **decomposed by type**.
- ⚠ **The revalidation stamp key is `at`, NOT `checked_at`.** A wrong key returns 0 rows and reads
  exactly like "nothing was scanned".
- ⚠ **A short grep needle is someone else's string.** This lane greped `no longer exists` and got
  **6** on a binary without the change. Verify every needle returns **0** on a pre-change build, and
  `N=${N:-0}` is not optional (`grep -c` prints nothing and exits 1 on zero).
- ⚠ **A DISPATCH OF THIS SWEEP CANNOT BE SCOPED.** Both filters read from the **step config**; the
  live `sweep` step has no `input_mapping`. Filters in `input_data` are INERT and the run goes
  fleet-wide over six types while looking scoped.
- ⚠ **The council gate refuses a plan with >8 edits** (*"wider than 8 files is architecture-shaped"*).
  If you must add one, **merge related edits and say so** — a silently vanished edit is what that cap
  exists to surface. It also takes only `modify|add|remove|config_change`; `create` is refused.
- ⚠ **An owner ruling recorded only in markdown is INVISIBLE to the council.** `compliance`
  objected that it had no independent record of one. The delivery path that exists today is
  `LANDMINES.md` → `scripts/landmines-sync.py --apply` → `doc_notes` where `categories ? 'landmine'`.
- **`scheduled_tasks.input_data` is INERT for this action**; `last_triggered_at` is written at
  publish time and is NOT proof an agent ran. `orchestration_states` retention is **~24h**.
- **Registering a revalidator is the whole wiring** — `coveredItemTypes()` derives from
  `reviewRevalidators`. `TestRevalidatorCoverageIsDeliberate` pins the set and fails on any change,
  telling you to update it and say why. That is deliberate.
- **`platform/orchestration/actions` has FAILING TESTS THAT ARE NOT THIS LANE'S.**
  `TestEveryCheckProducedItemTypeIsClassified` (wants `decision_regression` registered or
  acknowledged) fails at clean HEAD from `e1628f7df` (RFC_015). Others come and go from other
  sessions' uncommitted working-tree edits. **Reproduce against `git archive HEAD` before
  attributing any failure to your own change.**
- **`/tmp` is a near-full 16G tmpfs**; use `TMPDIR=/home/ant/.cache/buildtmp go build ./...`.
- **Use `git commit -F <file>`**; backticks in `-m` execute.

## 4. Correlations and commits

| what | id |
|---|---|
| **claims_unverified council (r1–r3 REVISE, r4 in flight)** | `b67eb26a-14ef-45d7-b755-3e489fd57ef0` · r4 run `a7b35edc-…` |
| status-ceiling diagnosis (both UNVERIFIABLE) | `f3d18013-…` · `a174b184-…` |
| voice_tells council (APPROVED r1) | `4d430ca8-7e34-479a-95f3-71fdc12fdef6` |
| selection filter council (APPROVED r1) | `f64da546-e1a4-42e4-98c0-d94cf42af71c` |

`ef80216be` voice_tells revalidator · `4030cadb9` claims_unverified + CQ-021 + 2 landmines ·
`6ab7ff594` the producer-count correction · **`9a9fef332` the OWNER'S copy-changed gate +
`features_open/031`** · `c9519c061` the 08-10 summary · `c70c8e1de` round 4.
