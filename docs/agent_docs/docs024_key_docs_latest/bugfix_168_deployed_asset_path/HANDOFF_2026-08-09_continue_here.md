# HANDOFF — 2026-08-09 — `voice_tells` PROVEN LIVE by a real closure; `claims_unverified` adopted and OWNER-GATED; council round 3 in flight

**Read this file only.** It supersedes `HANDOFF_2026-08-08b_continue_here.md` for state. That file
stays for its reasoning and traps, **and it now carries four dated corrections** — read those if
you are about to repeat one of its instructions, especially §0b's verification recipe, which was
wrong in a way that would have hidden a success.

**Nothing is half-applied.** The code is committed; two asynchronous runs are outstanding and
neither blocks you.

---

## 0. OPEN ITEMS — none of them is code

### 0a. ⏳ Council — REVISE at r1 AND r2, both answered; **owner has ruled; ROUND 3 in flight**

**Full read-out: `OBJECTIONS_2026-08-09_claims_unverified_council.md`.** r1 gated by `editquality`
(the producer-count error), r2 gated by `compliance` (the HITL policy question). 15 seats,
2 abstained, `gated_by_truncation: false` both rounds. Corrections `6ab7ff594`; owner's gate
`9a9fef332`; **round 3 running under the same trail** (`RESUBMIT_CORR=b67eb26a-…`, run
`0f8ce5a8-…`).

> **The gating objection caught a REAL ERROR, and it narrows the change.** This lane claimed
> **TWO converging producers** and invoked the owner ruling of 2026-08-02 §1 as the authority for
> shipping without an RFC. **There is ONE producer, so that ruling never applied.**
> `check_unverified_claims_stats.go` registers no check, has no `init()`, emits no `WorkItemSpec`
> — its `scanStoredStatClaims()` is called from inside `ScanDeployedClaims`
> (`check_unverified_claims.go:385`, `:427`) and nowhere else in production. Corrected in all five
> places it had spread to, visibly. The seat's sharper question — is the *scan logic* shared, not
> just the item_key shape? — is answered favourably by that same call graph: ONE scan, both halves
> inside it.
>
> ✅ **THE `compliance` OBJECTION IS RESOLVED — OWNER RULED 2026-08-09, and the answer is in code.**
> It gated rounds 1 AND 2 (round 2: *"needs confirmed owner sign-off BEFORE the revalidator is
> wired live, not concurrent notification"*), with `bug_historian` and `architecture` routing the
> same question to a human independently. The owner was given four costed options and chose the
> **copy-changed gate**: `resolved` now requires an EXAMINED component whose
> `page_components.updated_at` is later than the item's `created_at`. **A register edit alone can
> no longer retract a factual-claim finding.** Committed `9a9fef332`; round 3 submitted under the
> same trail (run `0f8ce5a8-…`). Full reasoning: `OBJECTIONS_2026-08-09_claims_unverified_council.md`.
>
> ⚠ **Three things about that gate a future editor must not undo:**
> 1. **The zero-`filedAt` arm is SEPARATE from the comparison.** `x.After(time.Time{})` is true for
>    any real timestamp, so folding them makes an item with no filing date close on ANY component
>    edit however old — the inverse of the gate. It was briefly present *behind a doc comment that
>    already claimed the correct behaviour*; the property test caught it.
> 2. **`NewestComponentUpdate` tracks EXAMINED components only.** A locked component's timestamp is
>    a change to content the scan deliberately did not read.
> 3. **It is NOT applied to `voice_tells`, deliberately** — identical hole, but live,
>    council-approved, and a style surface rather than a truth one. Extending it is a separate
>    argued change, not a tidy-up.
>
> **Honest cost:** the gate narrows what can close, so a page fixed by an edit that did not touch
> `page_components.updated_at` now refuses instead of closing. Intended direction, real cost — the
> type will drain more slowly than 23 rows suggests.

Committed `4030cadb9` with `Council-Submitted:`, which asserts nothing and is credited
automatically by `098` once approved — **no amend needed, and forward-only forbids one.**

```sql
-- the decisions, one row per round, oldest first:
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='b67eb26a-14ef-45d7-b755-3e489fd57ef0' AND kind='council_report' ORDER BY created_at;
-- THE OBJECTIONS THEMSELVES live in `body`, NOT in `metadata` (which carries only
-- decision/reviewers/abstained). Read the newest round for this correlation:
SELECT body FROM diagnosis_artifacts
WHERE correlation_id='b67eb26a-14ef-45d7-b755-3e489fd57ef0' AND kind='council_report'
ORDER BY created_at DESC LIMIT 1;
-- ⚠ NOT `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC
-- LIMIT 1` (CLAUDE.md's documented command). On a tree this concurrent that returns whichever
-- lane finished last — it handed me bugs_open/228's verdict, which read as plausible until it
-- started discussing contact forms. Always query by YOUR correlation.
-- still running? latency, NOT a dropped dispatch — do NOT resubmit:
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = 'b67eb26a-14ef-45d7-b755-3e489fd57ef0';
```

**The code is already on the shared branch, so a REVISE or REJECTED is real work, not a
formality.** Do not write `Council-Reviewed:` anywhere unless you have read an APPROVED verdict
yourself — `098` buckets an unread claim as MISMATCH, which is the report's dishonesty surface.

### 0b. ⏳ Diagnosis on the sweep's STATUS CEILING — first run UNVERIFIABLE on TOOLING, re-filed as `a174b184-dac2-47a1-95ca-df2d192e183a`

> **RUN 1 (`f3d18013-…`) came back `UNVERIFIABLE — stopped: iteration-cap`, and it is NOT evidence
> against the premise.** I scoped it at `work_items_common.go:workItemRevalidatableStatuses`, and
> **the code index contains no package-level vars or consts at all** — `code_symbols.kind` takes
> only `func` (3,592), `method` (1,114), `struct` (973), `alias` (40), `interface` (36). The loop
> could never open the list, said so honestly (*"0 rows — per the index-staleness caveat this is
> **unknown, not proof**"*), and spent its cap trying. That is a broken lookup wearing the costume
> of a hard bug.
>
> **RUN 2 (`a174b184-…`) — also UNVERIFIABLE (`scope-not-narrowing`), BUT it confirmed the
> load-bearing half and the wall is now understood.** Function-scoping got much further:
>
> > *"Both `loadParkedReviewItems` and `reportUncoveredBacklog` **demonstrably share this same
> > status-list filter (static fact, confirmed)**, which logically means any item_type whose rows
> > sit entirely outside that list is invisible to both — but I cannot confirm this is actually
> > OCCURRING without knowing which statuses the list contains."*
>
> **Mechanism: loop-confirmed. Membership: unreadable by any scoping**, for the `code_symbols`
> reason above, and **first-hand verified instead** — `work_items_common.go:140-143` is literally
> `{"needs_human_review", "unresolved"}`. That is the declared substitution the owner ruling of
> 2026-07-31 asks for, with a named reason. **Do not file a third run into the same wall.**
>
> ⚠ **The loop got a FIGURE wrong — check its data, not just its reasoning.** It reported
> `undeployed_asset` as having *"zero rows in needs_human_review/unresolved/failed"*. It has **50
> `unresolved`** (`complete 55 · unresolved 50 · detected 17 · triaged 17 · deferred 12 ·
> cancelled 11`). Taking that at face value would have had me "correct" a correct number.
>
> **Its slip produced a sharper example than the one I filed with:** `undeployed_asset` is *not*
> absent from `uncovered_types` — it is **listed at 50 while 46 more rows of the same type sit
> invisible**. So the report does not only omit whole types (`image_url_404`), **it understates
> types it already lists**, here by roughly half. Harder to notice than a missing key.

Intake 1 `0c9b44d2-5c74-4322-aa78-7dd206f92689` (item now `complete`, verdict UNVERIFIABLE);
intake 2 item_key `needs_diagnosis:uncovered-types-omits-open-item-types`.

**The claim under test, which I did NOT assert:** `reportUncoveredBacklog` counts parked rows with
the same `workItemRevalidatableStatuses` list that scopes the selection, so a type whose rows sit
in `blocked`/`detected`/`triaged`/`deferred` is **absent** from `uncovered_types` rather than
reported as uncovered. Measured 2026-08-08: **467 rows across 6 such statuses** — so the 625 this
lane steers by understates the parked population by roughly 43%.

**A REFUTED verdict is a success**, and the scoping may well be deliberate. Read it before
building anything on this. Widening the list is architecture-scope regardless: it is interpolated
in three places, and per its own comment widening the selection alone selects rows the write-time
CAS guards then silently refuse to update.

⚠ **`UNVERIFIABLE` is neither confirm nor refute — do not read run 1 as a refutation.** The premise
is first-hand verified independently of the loop: `work_items_common.go:140-143` is literally
`{"needs_human_review", "unresolved"}`, and `image_url_404` has 26 open rows yet appears nowhere in
the live `uncovered_types` map. What is still owed is an *independent* read, which is what run 2 is
for.

### 0c. 👀 Confirm `claims_unverified` by EFFECT after the next chassis roll — **and not the way §0b told you last time**

The Go change is inert until a roll. Baseline is **already spent** (below); after the roll:

```bash
for POD in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | sed 's|pod/||'); do
  A=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 'withdrawing the register is not evidence the claims were substantiated'" 2>/dev/null|tail -1); A=${A:-0}
  B=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 'an empty finding list here means the audit examined nothing'" 2>/dev/null|tail -1); B=${B:-0}
  C=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 'an unbuilt page is not evidence the claims were removed'" 2>/dev/null|tail -1); C=${C:-0}
  P=$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c 'auto:revalidated'" 2>/dev/null|tail -1); P=${P:-0}
  echo "$POD register=$A examined_nothing=$B unbuilt=$C positive_control=$P"   # want 1/1/1/non-zero
done
```

> **BASELINE 2026-08-09T09:36:10Z, `v1.0.1270` (predates commit `4030cadb9`), BOTH replicas:
> `0 / 0 / 0`, positive control `2`.** Purely string-additive change ⇒ no valid negative control
> exists; the dated **0 → 1 transition is the whole proof**, and it only exists because the 0 was
> taken first. `N=${N:-0}` is not optional (`grep -c` prints nothing and exits 1 on zero), needles
> are ASCII-only on purpose, and every one was verified to return 0 on a build predating the commit.

**Then by effect — READ THIS PART, the obvious check does not work:**

```sql
-- 1. the type must have LEFT the map (absent, not merely smaller)
SELECT collected_data #>> '{revalidation_result,uncovered_types}' AS uncovered_types,
       collected_data #>> '{revalidation_result,scanned}'         AS scanned,
       collected_data #>> '{revalidation_result,cap_binding}'     AS cap_binding
FROM orchestration_states WHERE orchestration_name ILIKE '%reval%' ORDER BY created_at DESC LIMIT 1;

-- 2. scanned, decomposed by type — must sum to `scanned`, with claims_unverified at ~23
SELECT item_type, count(*) FROM site_work_items
WHERE result #>> '{revalidation,at}' >= CURRENT_DATE GROUP BY 1 ORDER BY 2 DESC;

-- 3. closures outlive the ~24h orchestration retention
SELECT completed_at::date, item_type, count(*) FROM site_work_items
WHERE resolution_path='auto:revalidated' GROUP BY 1,2 ORDER BY 1 DESC;
```

> ⚠ **DO NOT confirm at `uncovered_backlog`.** It is a sum over ~40 types and it stayed **flat at
> 625** across the `voice_tells` transition that worked perfectly. See §2.
> ⚠ **The stamp key is `at`, not `checked_at`.** A wrong key returns 0 rows, which reads exactly
> like "nothing was scanned".
> ⚠ **Do NOT hand-dispatch to force it** — §3.5 of the previous handoff still holds: a dispatch
> cannot be scoped, so it runs fleet-wide over six types now.

---

## 1. State — verified 2026-08-09, chassis `v1.0.1270`, both replicas

| thing | state |
|---|---|
| `bugs_closed/168` (asset path) | CLOSED, live since `v1.0.1229` |
| Sweep starvation fix | LIVE, proven unattended |
| **`voice_tells` revalidator** | **LIVE AND BEHAVIOURALLY PROVEN 2026-08-09** — 32/32 scanned, **1 closed unattended**. §2 |
| **`claims_unverified` revalidator** | **BUILT + COMMITTED `4030cadb9`, CQ-021 — inert until the next roll.** §0c |
| Council (claims) | **SUBMITTED, verdict pending** — `b67eb26a-…`. §0a |
| Diagnosis (status ceiling) | **RUNNING** — `f3d18013-…`. §0b |
| Latest scheduled run (08-09 08:38Z) | `scanned 186 · cap_binding false · resolved 2 · uncovered_backlog 625` |
| Covered types | **6**: unresolved_cta, required_fields_missing, needs_section_data, needs_page, voice_tells, claims_unverified |
| Concept register | CQ-021 added, index 1,800 → **1,801** (clean triple: rows = row ids = entry ids) |
| Consumers told | `bugs_open/033` and `bugs_open/083` — **notice for claims_unverified still OWED**, §3.0 |
| Decision 2's dedup half | **OPEN, blocked on an owner judgement.** 47 pairs / 168 rows — **not** drifting up. §3.1 |
| Armed-but-inert cap in a sibling check | OPEN as a tripwire, untouched. §3.2 |

## 2. The one thing to take from today: the check that hides a success

`voice_tells` worked on its first unattended post-roll run — 08:38:53Z, `scanned` 151 → **186**
(63+42+34+**32**+15 by type, so **all 32 rows scanned**, up from zero), `cap_binding` false, and
one item genuinely CLOSED: `voice:ecfd0bfd-bc5c-4ed4-9c45-7ba9143e72c8`, page `ai-readiness-quiz`,
`resolution_path='auto:revalidated'`, carrying this code's own resolved-arm reason string.

**And `uncovered_backlog` did not move — 625 before, 625 after.** The previous handoff told this
session to confirm by watching exactly that number fall by ~32. Following it literally would have
recorded *"the adoption did nothing"* on the day it worked. What actually happened: `voice_tells`
left `uncovered_types` **entirely** (25 → absent) while nine other types grew by **exactly 25**
(`claims_unverified` +5, `content_rewrite` +5, `lock_blocked_change` +5, `save_refused_incomplete`
+4, `empty_internal_href` +2, five more at +1). The coincidence is incidental — **any inflow makes
a ~40-type total uninformative about one type.**

Recorded in `LANDMINES.md` and in both register entries' `verify-later`, so the next adopter
cannot inherit the recipe. This is the second time in this lane that **a trap documented in one
paragraph was walked into in the next** — the first was §3.5's scoped-dispatch correction.

## 3. What is left

### 3.0 Owed from today: the consumer notice for `claims_unverified`

`bugs_open/033` and `bugs_open/083` carry dated CONSUMER NOTICEs for `voice_tells`. The same is
owed for `claims_unverified` — their parked-item tallies move again, and this type is a **factual**
claims surface, which is a higher-stakes close than voice's style findings. Per the owner ruling of
2026-07-29 §3: **name the consumers and TELL them; measuring that nothing breaks is not the same
as establishing they would have agreed.** Say what changed about their guarantee, not which keys
were added.

### 3.1 Decision 2's dedup half — the only substantive work, still blocked

47 colliding `(site_id, item_key)` pairs across 168 rows (2026-08-09, PROPOSED predicate).
`CREATE UNIQUE INDEX` fails against this population; the cleanup needs a *"which copy do I keep,
and does discarding the rest lose a true finding?"* judgement that is the owner's.

⚠ **READ THE INDEX, DO NOT RECONSTRUCT IT** — `SELECT pg_get_indexdef(oid) FROM pg_class WHERE
relname='idx_swi_dedup'` is the only authority. Writing the exclusion list from memory of the
phrase "terminal statuses" produced 75/227 and nearly recorded a growth figure that was an
artefact of the query.
⚠ **The "drifts upward ~2 pairs/day" line in every prior doc is WITHDRAWN** — four points bounce
(48/135, 53/180, 55/184, 47/168); it fell 8 pairs in 14 hours. Do not re-derive urgency from it.

### 3.2 The armed-but-inert cap — a tripwire, not a task

`discovery_checks/check_image_source_unsatisfiable.go:167` is still `return result, nil` inside its
per-pass cap and still populates `Resolved` **0** times. **Correct today, and untouched by this
commit.** The commit that adopts the retraction seam there is the commit that must change it to
`break`.

### 3.3 More sweep coverage — and the census now needs a status filter

**Re-run the census yourself, and include the status filter or you will nominate types the sweep
structurally cannot see** (this is how `image_url_404` — 26 open, 0 closed ever, flag-only, DB-
answerable — looked like an obvious candidate and turned out to be unreachable):

```sql
SELECT item_type,
       count(*) FILTER (WHERE status IN ('needs_human_review','unresolved'))          AS selectable,
       count(*) FILTER (WHERE status IN ('complete','verified'))                      AS closed_ever,
       count(*) FILTER (WHERE status IN ('complete','verified') AND result ? 'deploy_result') AS closed_by_deploy,
       count(DISTINCT site_id) FILTER (WHERE status IN ('needs_human_review','unresolved')) AS sites
FROM site_work_items GROUP BY 1 HAVING count(*) FILTER (WHERE status IN ('needs_human_review','unresolved')) > 0
ORDER BY 2 DESC;
```

`closed_ever > 0` ⇒ something already drains it. A `deploy_result` block ⇒ a real fix pipeline
owns it and retraction is the wrong tool (that is what disqualified `content_rewrite`). A
`result.revalidation` block on an UNCOVERED type is a legacy stamp from before the covered-type
selection filter — not evidence of a closer.

Zero-closer types that are actually selectable, measured 2026-08-09:

| candidate | selectable | sites | note |
|---|---|---|---|
| `lock_blocked_change` | 23 | 4 | ⚠ see LANDMINES — its items cannot distinguish an obeyed lock from a leaked one |
| `image_source_unsatisfiable` | 18 | 2 | ties to §3.2's tripwire — adopting it is the commit that flips the cap to `break` |
| `needs_sprite_css` | 10 | 1 | source comment says asset-deployer's sprite_css mode re-runs; needs its own producer/closer pass |
| `dead_control` | 8 | 4 | |
| `stale_evidence` | 5 | 5 | |
| `cta_names_unknown_destination` | 123 | 10 | **DO NOT TOUCH** — owned by `bugs_open/023`, and a test asserts it stays out of the registry |

### 3.4 Or take the next unowned bug from `bugs_open/`

`scripts/who-owns.py <n>` **plus** a grep of live `.jsonl` transcripts — the script reads commits
and is blind to a session mid-fix.

## 4. Traps specific to this lane

- ⚠ **`uncovered_backlog` cannot confirm an adoption.** §2. Use `uncovered_types` + decomposed `scanned`.
- ⚠ **The revalidation stamp key is `at`, not `checked_at`.** A wrong key returns 0 rows.
- ⚠ **A short grep needle is someone else's string.** This lane greped `no longer exists` and got
  **6** on a binary without the change. Verify every needle returns 0 on a pre-change build.
- ⚠ **A DISPATCH OF THIS SWEEP CANNOT BE SCOPED.** Both filters read from the **step config**; the
  live `sweep` step has no `input_mapping`. Filters in `input_data` are INERT and the run goes
  fleet-wide while looking scoped.
- **`scheduled_tasks.input_data` is INERT for this action**; `last_triggered_at` is written at
  publish time and is NOT proof an agent ran.
- **`orchestration_states` retention is ~24h.** Record a payload the day you take it.
- **Registering a revalidator is the whole wiring** — `coveredItemTypes()` derives from
  `reviewRevalidators`, so the map entry widens the selection in the same edit.
- **`TestRevalidatorCoverageIsDeliberate` pins the covered set** and fails on any change, telling
  you to update it and say why. It fired today; it should.
- ⚠ **The 097 council trigger takes `modify|add|remove|config_change`** — `create` is refused
  client-side. A new file is `add`.
- **Two tests in `platform/orchestration/actions` FAIL AT HEAD and are not this lane's**:
  `TestValidDocSubjectTypes_LockstepWithMigrationCheck` and
  `TestEveryCheckProducedItemTypeIsClassified` (wants `decision_regression` registered or
  acknowledged). Reproduced on a clean `git archive HEAD` extraction — they came in with
  `e1628f7df` (RFC_015). Do not attribute them to your own change, and do not "fix" them blind.
- **The shared tree is not a build signal.** Build against `git archive HEAD` plus your own files.
- **`/tmp` is a near-full 16G tmpfs**; use `TMPDIR=/home/ant/.cache/buildtmp go build ./...`.
- **Use `git commit -F <file>`**; backticks in `-m` execute.

## 5. Correlations and commits

| what | id |
|---|---|
| **claims_unverified council (SUBMITTED, pending)** | `b67eb26a-14ef-45d7-b755-3e489fd57ef0` |
| **status-ceiling diagnosis (RUNNING)** | `f3d18013-0b78-472f-b2cb-5bf5e4e893b8` (intake `0c9b44d2-…`) |
| voice_tells council (APPROVED r1) | `4d430ca8-7e34-479a-95f3-71fdc12fdef6` |
| selection filter council (APPROVED r1) | `f64da546-e1a4-42e4-98c0-d94cf42af71c` |
| RFC_010 Decision 1 council | `846f4f3d-8958-4e4c-be81-d5f02e20852d` |

`ef80216be` voice_tells revalidator · `23e5b6721` its objections answered · `a697ec4e1` the
08-08b handoff · **`4030cadb9` the claims_unverified revalidator + CQ-021 + 2 landmines** ·
this commit: corrections to the 08-08b handoff, NOTES, the owner log, and this file.
