# HANDOFF — 2026-08-08b — `voice_tells` adopted and APPROVED. One baseline left to spend.

**Read this file only.** It supersedes `HANDOFF_2026-08-08_continue_here.md` for state; that file
stays for its reasoning and its traps, all of which still hold. Working and missteps:
`NOTES_deployed_asset_path.md` (newest at the bottom).

**One thing is owed — the post-roll grep in §0b — and it is cheap. Everything else is finished,
blocked-by-design, or waiting on an owner call.**

---

## 0. THE OPEN ITEM — do this first

### 0a. ✅ DONE — council **APPROVED r1**, and the five objections are answered

> **APPROVED, 13 seats, 5 advisory objections, none high, `gated_by_truncation: false`.**
> Answers with the checks run: **`OBJECTIONS_2026-08-08_voice_tells_council.md`**. Committed
> `23e5b6721` with `Council-Reviewed: 4d430ca8-7e34-479a-95f3-71fdc12fdef6`.
>
> **Two objections changed something.** (1) `debug_historian` was **right about the mechanism**:
> `pages.status` takes only `active` (585) and `archived` (29) fleet-wide — **`'deployed'` never
> occurs**, so half the shared `p.status IN ('active','deployed')` disjunct is dead. Feared
> consequence refuted (all items sit on `active`, so the revalidator is not inert); the literal is
> inherited from the emit side and **stays**, recorded rather than silently narrowed.
> (2) **The population is 32, not 25** — see the correction below.
>
> One objection is **owner-facing and deliberately unactioned**: `bug_historian` wants the
> *registration contract* constrained (every `HandlerAgent: ""` item_type must have a revalidator or
> a documented exemption). That is a shared-contract change at architecture scope, and the
> `architecture` seat recorded `point_fix` in the same round — **seats disagreeing is the signal it
> needs a human.** Filed into `bugs_open/033`, whose ~175 figure is its evidence.

<details><summary>Original §0a text, kept — how to read a verdict that is still running</summary>

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='4d430ca8-7e34-479a-95f3-71fdc12fdef6' AND kind='council_report'
ORDER BY created_at;
-- human-readable:
SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
-- still running? this is latency, not a dropped dispatch — do NOT resubmit:
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '4d430ca8-7e34-479a-95f3-71fdc12fdef6';
```

**The code is already on the shared branch** (`ef80216be`), committed with `Council-Submitted:` —
which asserts nothing and is credited automatically by `098` once the verdict turns approved, with
no amend. **So a REVISE or REJECTED is real work, not a formality.** Do not write
`Council-Reviewed:` anywhere unless you have read an APPROVED verdict yourself.

</details>

### 0b. ⚠ THE PRE-ROLL BASELINE IS TAKEN. SPEND IT AFTER THE NEXT ROLL.

The change is **purely string-additive**, so there is **no valid negative control** — nothing was
removed that can be greped for 0. For an additive change the dated **0 → 1 transition is the whole
proof, and it only exists because the 0 was taken before the roll.** Here it is:

> **BASELINE 2026-08-08T17:13:45Z, `v1.0.1267` (predates the commit), BOTH replicas:**
> `opting out is not evidence the copy was fixed` = **0** ·
> `the scan read nothing` = **0** ·
> `an unserved page is not evidence the prose was fixed` = **0** ·
> positive control `auto:revalidated` = **2**

```bash
for POD in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | sed 's|pod/||'); do
  A=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 'opting out is not evidence the copy was fixed'" 2>/dev/null|tail -1); A=${A:-0}
  B=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 'the scan read nothing'" 2>/dev/null|tail -1); B=${B:-0}
  C=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 'an unserved page is not evidence the prose was fixed'" 2>/dev/null|tail -1); C=${C:-0}
  P=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 'auto:revalidated'" 2>/dev/null|tail -1); P=${P:-0}
  echo "$POD optout=$A read_nothing=$B unserved=$C positive_control=$P"   # want 1/1/1/non-zero
done
```

**After the roll all three must be ≥1 on every replica, positive control still non-zero.** A 0/0/0/0
means the probe broke, not that the change did not ship. `N=${N:-0}` is not optional (`grep -c`
prints nothing and exits 1 on zero) and the needles are **ASCII-only on purpose**.

⚠ **My first baseline attempt used a BAD NEEDLE and I nearly recorded it.** I grepped
`no longer exists` — a fragment of one of my new reason strings — and got **6** on a binary that
does not contain my change at all. It matches six unrelated strings in the binary. A needle short
enough to be convenient is long enough to be someone else's. **Every needle above was verified to
return 0 on a build that predates the commit**, which is the only thing that makes the later 1
mean anything.

**Then prove it by effect**, which the strings cannot — the next scheduled sweep should scan the
`voice_tells` rows too, so `scanned` rises and `uncovered_backlog` falls by the same amount:

```sql
SELECT collected_data #>> '{revalidation_result,scanned}'           AS scanned,
       collected_data #>> '{revalidation_result,cap_binding}'       AS cap_binding,
       collected_data #>> '{revalidation_result,resolved}'          AS resolved,
       collected_data #>> '{revalidation_result,uncovered_backlog}' AS uncovered_backlog
FROM orchestration_states WHERE orchestration_name ILIKE '%reval%' ORDER BY created_at DESC LIMIT 1;
-- and the closures themselves, which outlive the ~24h orchestration retention:
SELECT completed_at::date, count(*) FROM site_work_items
WHERE resolution_path='auto:revalidated' GROUP BY 1 ORDER BY 1 DESC;
```

`scanned` should rise by roughly the live `voice_tells` count (**32 and growing** — re-count it,
do not carry this number); `uncovered_backlog`
should FALL by the same. **`cap_binding` must stay `false`** — if it ever goes `true`, judgeable work
is being left behind and the starvation this lane fixed is back.

---

## 1. State — verified 2026-08-08, chassis `v1.0.1267`, both replicas

| thing | state |
|---|---|
| `bugs_closed/168` (asset path) | CLOSED, live since `v1.0.1229` |
| Sweep starvation fix | **LIVE, PROVEN UNATTENDED.** Re-greped 1/1/2 on `1264`, `1266`, `1267` |
| Latest scheduled run (`1ac359c4`, 08:38Z) | `scanned 151 · cap_binding false · resolved 3 · uncovered_backlog 625` |
| Closures by day | 3 (08-08) · 1 (08-07) · 21 (08-06) · 33 (08-04) — steady state |
| **`voice_tells` revalidator** | **COMMITTED `ef80216be`, council APPROVED r1, INERT until the next roll.** §0b |
| Council objections | **ANSWERED** — `OBJECTIONS_2026-08-08_voice_tells_council.md`, commit `23e5b6721` |
| `voice_tells` population | **32 and GROWING** (25 filed 07-17, 7 filed 08-08). NOT a fixed backlog |
| Concept register | CQ-020 added, index 1,794 → **1,795** (re-greped, 1,795 unique ids) |
| Consumers told | `bugs_open/033` and `bugs_open/083` carry a dated CONSUMER NOTICE |
| **Decision 2's dedup half** | **OPEN, blocked, and drifting.** §3.1 |
| Armed-but-inert cap in a sibling check | OPEN as a tripwire. §3.2 |
| RFC_010 Q1 (two-strike) | Owner-ruled accept-as-is, tracked, not work |

## 2. What was built today, and the one thing to understand about it

The sweep covered four item_types; `voice_tells` is the fifth. **Nothing had ever closed one** —
the CLOSER census (`item_type='voice_tells' AND status IN ('complete','verified')`) returns **zero
rows**, its producer sets `HandlerAgent: ""` by design, and the items sit at `needs_human_review`
indefinitely on leopardessconsulting.co.uk.

> **CORRECTED — the population is 32, not the 25 I first recorded.** Seven more were filed on
> **2026-08-08** by `quality-discovery-agent` while the revalidator was being built, found only
> because a council objection sent me back to the provenance table. **The check is actively filing,
> so this type GROWS.** Do not size anything off 25. Churn over all 32 is still 13.
>
> ⚠ Two `created_by` values (`generic`, `quality-discovery-agent`) look exactly like two producers
> and are **two agents running one check** — `created_by` is `dctx.AgentType`, not the filing code.
> The producer question is answered by the Go call-site census (ONE, `check_voice_tells.go:114`),
> never by `created_by`.

**Retraction is not the auto-rewrite the check forbids.** `check_voice_tells.go`'s stored `fix` text
says *"never an unreviewed auto-rewrite"*, and `bugs_open/033` quotes it as evidence the type was
filed correctly for human review. That governs the **fix** path. Retraction never edits copy and
never dispatches a rewrite — it withdraws a claim the current page no longer supports. `083`
classifies the type as *advisory / machine-fixable in principle*, not *needs a human ANSWER*, so no
open owner decision was pre-empted.

**The design choice that matters, and the trap inside it.** `actions` imports `discovery_checks`
one-way, so the shared scan lives in `discovery_checks` as exported `ScanVoiceTells`, called by
**both** the check and the revalidator — the two ends of an item's life cannot answer *"does this
read machine-written?"* differently. And:

> **`len(Findings) == 0` collapses THREE states into one:**
> 1. components examined and clean → the prose was fixed — **the only state that may close an item**
> 2. **nothing was read at all** — page deleted, not `active`/`deployed`, no rendered components
> 3. **only human-LOCKED components were read** — the emit side has always skipped those
>
> So `VoicePageScan` reports `ComponentsExamined` and `ComponentsSkippedLocked`, and the ladder
> refuses on 2 and 3. **The no-op case is the dangerous one here, not the damage case** — a wrong
> `still_holds` costs a human glance, a wrong `resolved` closes a live finding.

**[UNEXERCISED, not proven unreachable]** All rows today have `has_locked_components 0` and
`no_unlocked_components 0`, so the two locked arms are reasoned and unit-tested, never observed.
**[MEASURED]** 13 pages have a component updated since filing, so this judges real change rather
than running inert — that was the disconfirming check, and a 0 would have killed the change.

⚠ **LANDMINE — the voice gate is a MOVING STANDARD.** If a site loosens its `voice_gate` thresholds,
previously-failing copy scans clean and the item retracts **although the prose never changed**.
Arguably correct (the site redefined its own standard) but **a `resolved` stamp is not proof the
copy was rewritten.** To tell them apart, compare `page_components.updated_at` against the item's
`created_at`. In CQ-020 and the submission's risks block.

## 3. What is left

### 3.1 Decision 2's dedup half — still the only substantive work, still blocked, now worse

Owner ruled `unresolved` is OPEN, so it must leave `idx_swi_dedup`'s exclusion list. Re-measured
**2026-08-08 against the PROPOSED predicate**: **55 colliding `(site_id, item_key)` pairs across 184
rows**, up from 53/180 this morning and 48/135 on 08-03. `undeployed_asset` 48, `improve_tool` 30,
`needs_internal_links` 29. `CREATE UNIQUE INDEX` fails against this population; the cleanup needs a
*"which copy do I keep, and does discarding the rest lose a true finding?"* judgement. **It drifts
upward roughly two pairs a day, so this gets more expensive with every day it waits.**

⚠ **READ THE INDEX, DO NOT RECONSTRUCT IT** — `SELECT pg_get_indexdef(oid) FROM pg_class WHERE
relname='idx_swi_dedup'` is the only authority. Writing the exclusion list from memory of the phrase
"terminal statuses" produced 75/227 and nearly recorded a 56% growth that was an artefact of the
query. Ordering hazard unchanged and asymmetric — §4.3 of `HANDOFF_2026-08-04` for the `42P10`/`23505`
argument.

### 3.2 The armed-but-inert cap — a tripwire, not a task

`discovery_checks/check_image_source_unsatisfiable.go:167` is still `return result, nil` inside its
per-pass cap and still populates `Resolved` **0** times. **Correct today.** The commit that adopts
the retraction seam there is the commit that must change it to `break`.

### 3.3 More sweep coverage — and two of the three named candidates were wrong

`uncovered_backlog` is 625. Closing it means teaching the sweep more types, one at a time.
**The previous handoff's candidate list did not survive the CLOSER check**, so re-run it yourself:

| candidate | what the census said |
|---|---|
| `content_rewrite` (~34) | **REJECTED** — 51 `complete` rows carrying `deploy_result` payloads. A real fix pipeline already drains it |
| `needs_sprite_css` (10) | **DEFERRED** — zero closed, but all 10 are `unresolved` and its source comment says *asset-deployer's sprite_css mode re-runs*. Needs its own producer/closer pass |
| `voice_tells` (25) | **ADOPTED today** |
| `cta_names_unknown_destination` (~123) | **DO NOT TOUCH** — owned by `bugs_open/023`, and a test asserts it stays out of the registry |

```sql
-- run this BEFORE writing any code, every time
SELECT status, count(*), left(result::text,120) FROM site_work_items
WHERE item_type='<your type>' AND status IN ('complete','verified') GROUP BY 1,3 ORDER BY 2 DESC;
```
A `result.revalidation` block means `revalidate_review_queue` already owns the type — extend it
rather than writing a second closer. **A `deploy_result` block means a real fix pipeline owns it, and
retraction is the wrong tool.**

### 3.4 Or take the next unowned bug from `bugs_open/`

57 files. `scripts/who-owns.py <n>` **plus** a grep of live `.jsonl` transcripts — the script reads
commits and is blind to a session mid-fix.

## 4. Traps specific to this lane

- **`scheduled_tasks.input_data` is INERT for this action.** `max_items` *and* `item_type` come from
  the **step config**; the `sweep` step has no `input_mapping`. "Several scheduled rows, one per
  type" cannot work — they would all read the same config.
- **`last_triggered_at` is written at publish time and is NOT proof an agent ran.**
- **A row missing a stamp its siblings have may never have been LOADED.** Check reachability before
  theorising about a predicate.
- **`orchestration_states` retention is ~24h.** Record a payload the day you take it or it is gone.
- **Registering a revalidator is the whole wiring** — `coveredItemTypes()` is derived from
  `reviewRevalidators`, so the map entry widens the selection in the same edit. You do not have to
  think about the cap.
- **`TestRevalidatorCoverageIsDeliberate` pins the covered set** and fails on any change, telling you
  to update it and say why. That is deliberate; it fired today and it should.
- ⚠ **A short grep needle is someone else's string.** See §0b.
- ⚠ **`cmd | head -N && echo OK` prints OK on a FAILED command**, because `head` exits 0. I built
  that landmine and then read a truncated error from it. Gate on `${PIPESTATUS[0]}`.
- **The shared tree is not a build signal.** Another session's mid-write left `go build` failing on
  `undefined: verificationOutcome`; it passed minutes later untouched. Build against
  `git archive HEAD` plus your own files, and delete the extraction afterwards.
- **`/tmp` is a near-full 16G tmpfs**; use `TMPDIR=/home/ant/.cache/buildtmp go build ./...`.
- **Use `git commit -F <file>`**; backticks in `-m` execute.

## 5. Correlations and commits

| what | id |
|---|---|
| **voice_tells council (SUBMITTED, verdict pending)** | `4d430ca8-7e34-479a-95f3-71fdc12fdef6` |
| selection filter council (APPROVED r1) | `f64da546-e1a4-42e4-98c0-d94cf42af71c` |
| RFC_010 Decision 1 council | `846f4f3d-8958-4e4c-be81-d5f02e20852d` |

`b14609e05` mig 321 · `0e4e79124` selection filter · `246763083` docs/notices · `c21b7e216` live
proof · `45c37b4f8` landmine · `9e4caa2b3` SUMMARY 08-06b · **`ef80216be` the voice_tells
revalidator** · this commit: docs, CQ-020, consumer notices.
