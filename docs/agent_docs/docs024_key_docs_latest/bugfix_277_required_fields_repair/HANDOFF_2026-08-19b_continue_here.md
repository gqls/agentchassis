# HANDOFF — 2026-08-19b, evening. `v1.0.1316` is live, `301` is APPROVED and being closed by another session, and the day's real finding is that a re-routed producer STRANDS its own backlog

**Supersedes `HANDOFF_2026-08-19_continue_here.md`.** That file stays authoritative for everything it
records about `277`, `083`, `300` and `314` — **read it second, it is not redundant.** What is stale in
it is named in §1 below. **Read this from disk, then `NOTES_required_fields_repair.md` from the
bottom** (two new entries today, `~16:20Z` and `~20:40Z`).

---

## 0. STATE TABLE

| bug | state | what blocks the close |
|---|---|---|
| **`bugs_open/277`** | router live, approved, doing its job | **its own verify clause 1 — the worked example must be REPAIRED.** Classified, not repaired. Nothing repairs `no_content_data` (27 of 30 parked rows). Unchanged today |
| **`bugs_open/083`** | fix live + artefact-proven; Tier 1 behaviourally proven | the door soak, **~2026-08-25** (owner decision 5). ⚠ `479`'s reclaim arm **has still never fired — 0 all-history, re-checked 20:45Z tonight**, live+archive. The close must SAY that, not imply it works |
| **`301`** | **✅ COUNCIL APPROVED r2 · fix live on `v1.0.1316`** | **being CLOSED right now by ANOTHER SESSION — do not touch the file (§3)** |
| `bugs_open/300` | fix live, council APPROVED r1 | **demand, not a defect.** No `page_component_status_drift` dispatch since 08-18 09:49. Nothing to verify against. Honest resting state — leave it |
| `bugs_open/314` | filed, unfixed | owner's call between four candidates |

---

## 1. WHAT IS STALE IN THE PREVIOUS HANDOFF — three things, all dated

1. **§1's roll is superseded.** It verifies `v1.0.1315` (12:15Z). **`v1.0.1316` rolled at 17:13Z** and
   is re-probed below. ⚠ My own NOTES entry at 16:20Z said *"no new roll, so the probe was not
   re-run"* — **true when written, stale within four hours, and caught by the owner rather than by
   me.** A deploy fact has a shelf life of hours on this tree; write one with its timestamp or not
   at all.
2. **§7f's `literal_markdown` numbers moved a long way, in the good direction** — see §4.
3. **§10.A's escalation table: the DATES are all confirmed, one ROW COUNT is stale, and its
   "first real escalation" line needs one word of precision** — see §5.

---

## 2. THE ROLL — `v1.0.1316`, probed on both replicas, and this time proven for every pod

Pods `86nqf` (17:13:39Z) and `8jlqh` (17:14:01Z). The `build provenance` startup line had **already
scrolled** — this service writes ~3.7MB to `--tail=400` — which means **"not in range", never
"unstamped"**. So the binary probe, which has no shelf life:

```sh
kubectl -n ai-persona-system exec <pod> -- sh -c \
 "grep -aoE 'owned_page_refusal_status|resolveStatusRepairComponent|refuse_owned_page|OWNED_PAGE_GUARD|ZZQQ_NEEDLE_THAT_MUST_NOT_EXIST' /proc/1/exe | sort -u"
```

`owned_page_refusal_status` **PRESENT** · `resolveStatusRepairComponent` **PRESENT** ·
`refuse_owned_page` **PRESENT** (`301`'s opt-in key) · `OWNED_PAGE_GUARD` **PRESENT** (control: the
probe works) · nonsense needle **ABSENT** (control: it discriminates). **Both replicas.**

⚠ **Budget ~420s per pod.** The probe scans the whole binary; a 2-minute timeout kills it mid-scan and
the truncated output is indistinguishable from a clean negative.

**AND THE FLEET-WIDE HALF, which is new practice and better than what we were doing:** all pods
running the chassis image resolve to **one** `imageID` — `distinct digests: 1`,
`sha256:2d0d3def…`. That upgrades *"I probed 2 replicas"* into *"those 2 replicas' bytes are every
pod's bytes"* without exec'ing the fleet, and closes the same-tag-cached-binary trap in the same
command. Recipe now in `LANDMINES.md` under the `-l app=agent-chassis` entry.

> ⚠ **AND DO NOT QUOTE A POD COUNT.** Three sessions measured **22, 57 and 85** within half an hour
> tonight, all correctly: **81 of 85 were `Job`-owned**, per-work-item pods that spawn and age out
> continuously; only **4** were `ReplicaSet`-owned, and the `app=agent-chassis` label shows **2**.
> The durable claims are `distinct digests: N` and the `ReplicaSet` count. The total is scrollback.
> (I got this wrong first, in a landmine bullet, and corrected it 20 minutes later — `3239e2bb0`.)

---

## 3. `301` — APPROVED, AND ANOTHER SESSION IS CLOSING IT. HANDS OFF THE FILE.

**Council round 2 APPROVED** — `RESUBMIT_CORR=c7bc1b9e-97c8-4f3e-8a4f-b3a7029505ee`, orchestration
`6469c138`, `complete_approved` **16:19:28Z, 11 minutes end to end** (CLAUDE.md budgets ~30).
`gated_by_truncation: false`, so the architecture seat's truncation landmine did not fire; its verdict
is whole and reads `ARCHITECTURE_SIGNAL: point_fix`, calling this *"the RFC_022 exception applied
correctly, not just claimed"*.

⚠ **THE FILE IS MID-MOVE BY ANOTHER SESSION.** At 20:44Z: on disk it is at
`bugs_closed/301_…`; **at HEAD it is still at `bugs_open/301_…`** — the rename is *staged*, with
content edits *unstaged* (`git status` reads `RM`). **Do not edit, do not commit those paths.** A
pathspec commit naming one side ships half the move — the `git mv` landmine. Verify with
`git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 301`, never with `ls`.

**No trailer is owed on the code.** `6be66bceb` carries `Council-Submitted: c7bc1b9e` and `098`
resolves the correlation at report time, crediting it automatically. No amend (forward-only forbids
one).

> ⚠ **My slip, so it does not puzzle the coverage report:** I put `Council-Reviewed: c7bc1b9e…` on a
> **LANDMINES docs commit** (`895df4e2f`). Not a false claim — I read the full verdict and every
> objection first — but the trailer exists to join *platform-code* commits, and docs are refused by
> the gate client-side. It stands; **do not repeat it.**

### The three MEDIUM advisories were checked, not waved through — and TWO ARE WRONG

The other session reached the same three conclusions independently, which is corroboration rather
than duplication. Recorded here because a close-out that repeats a false premise carries it for ever.

- **`diagnosis_guardian` — REFUTED AT SOURCE, and the seat's standing discipline is STALE.** It says
  the coordinator reads *only* `step.config.error_step` and a step-level one is *"parsed but silently
  inert"*, making our routing possibly coincidental. `platform/orchestration/coordinator.go:3667`
  (`routeToErrorStepOrFail`) checks `step.ErrorStep` **FIRST**, with the comment *"Check step-level
  first (parallel to NextStep) — **preferred location**"*, and falls back to `step.Config["error_step"]`
  *"for backward compatibility"*. **The precedence is exactly inverted.** ⚠ **Worth telling that
  seat** — a stale discipline mis-fires on every future submission that does the right thing, which
  is the pathology RFC_022 was narrowed to fix.
- **`bug_historian` — PREMISE VOID.** It rests on *"an OPEN case: `bugs_open/086`
  step_level_error_step_dropped_by_the_plan_converter"*. **086 is in `bugs_closed/`, closed
  2026-07-27** — and on exactly the evidence the objection wanted: *"the persisted plans show a clean
  0 → 10 step across the roll boundary"*. **Their layer distinction survives and is worth keeping**:
  the *coordinator* honouring step-level `ErrorStep` and the *plan converter* preserving it are two
  different questions. Both are now answered, by different evidence.
- **`debug_historian` — VALID, and already `LANDMINES.md` line 5909.** I had not read it: the
  SessionStart hook matches entries against files already **dirty in the tree**, and a `kubectl`
  footprint matches no path. **Grep LANDMINES by command and table too, not just by path.** Answered
  better than asked, via the digest check in §2.

**The pattern: every objection was cheap to check and none needed the author's word** — one `grep`
and one `ls` refuted two of them. That is the case for reading advisories on an *approved* verdict
rather than filing them.

---

## 4. `literal_markdown` — RESOLVED on the new route, and §7f's "one number to watch" is comfortable

[MEASURED 2026-08-19 ~16:16Z, live + archive]

| handler | complete/verified | failed | detected | % of outcomes |
|---|---|---|---|---|
| **`page-rerender`** (the new route) | **7** | **1** | 0 | **87.5%** |
| `page-build-handler` (the old route) | 3 | 34 | **7** | 8.1% |
| `page-content-writer` (older still) | 2 | 9 | 0 | 18.2% |

§7f left this at 1 complete / 2 failed and warned that `floor_ok` binds at the 5th outcome with 2
failures banked. **Eight outcomes at 87.5% — the floor cannot bind, and that worry is closed.** Seven
of those completions landed **between 16:00Z and 16:11Z**, while the previous handoff was being read.

⚠ **A THIRD handler exists** — `page-content-writer` (2/9), in neither §7e nor §7f, which both frame
this as a two-route story. The producer's header records the chain: `page-content-writer` →
`page-build-handler` (08-05) → `page-rerender` (08-18). **Three eras, three pairs, each keeping its
own record for ever.**

---

## 5. THE FINDING — a re-routed producer STRANDS its existing backlog, and nothing tells you

This is the day's transferable result. Full evidence in NOTES `~16:20Z`; the `LANDMINES.md` entry
*"Re-routing an `item_type` to REPAIR it creates a NEW pair…"* now carries both halves.

**The landmine had been retracted to a derived property with NO measured instance. This is the
instance.** `check_literal_markdown.go:402` began filing `HandlerAgent: "page-rerender"` at
`763bb5d55` (08-18 20:08), live only from today's roll — **so every row created before that was filed
`page-build-handler`.** Yet rows from **one detector run, identical to the microsecond**
(`created_at = 2026-08-18 07:23:16.545362+00`) now sit on **both** handlers: 7 `detected` on the old,
2 on the new. The producer cannot do that. **`handler_agent` was mutated on existing rows.**

**And the full prescribed sequence is visible in the artefacts:** 4 new-route rows carry
`pipeline='build'` + `spec.original_pipeline` — **migration `466`'s hand-canary recipe verbatim** —
and 5 carry the promoter's plain `pipeline='content'`, with the `build` ones ordered *first* by
`updated_at`. *Re-point → canary one → `known_good` flips → promoter takes the rest.* **Measured, not
reasoned.**

### The half that was not written down, and it is the one that costs something

**Re-pointing was PARTIAL.** Seven rows of that same batch still carry `page-build-handler`, with
`updated_at` **equal to `created_at`** — untouched for 33 hours while their siblings were repaired.
They are unreachable from both directions:

- the **promoter** will not dispatch them — their pair is 3/34 = 8.1%, held under the floor;
- the **producer** cannot re-file over them — `idx_swi_dedup` holds one open row per
  `(site_id, item_key)`, key `literal_markdown:<page_id>`, so re-detection is silently dropped while
  the old row is open. **[INFERRED from index semantics + the codebase's own "silently drop" language;
  I did not read the INSERT.]** **[MEASURED]** is only that they have not moved in 33 hours.
- **demand control** (because this lane keeps shipping zeros that mean nothing): the discovery
  machinery is emphatically alive — **60 rows across 11 item types filed since the 12:15Z roll**, most
  recent 16:14Z. **Zero of them `literal_markdown`.**

**On 08-21 those 7 rows escalate to `needs_human_review` carrying the reason *"the pair succeeds below
25%, the promoter has stopped feeding it"* — true of the route they are pinned to, and not the reason
they are actually stuck.**

> ## ⚠⚠ CORRECTED 2026-08-19 21:00Z, BEFORE THIS HANDOFF WAS AN HOUR OLD — I HAD THE REMEDY BACKWARDS AND IT WOULD HAVE DONE DAMAGE
>
> This section originally said *"the remedy is proven in this very population — an explicit `UPDATE …
> SET handler_agent='page-rerender'`"*, and called the 7 rows a working repair away. **Both wrong.**
>
> **What caught it:** running `scripts/who-owns.py` before routing the finding — which said `184` is
> **CLOSED** (today, `0ca143c2d`) and that its close-out had **already routed this exact residual**
> (*"owned/ported → 301/tool-rebuilds"*). That made me ask what the 7 rows actually **are**, which I
> had not done in either file I had already written the claim into.
>
> **[MEASURED 21:00Z] All 7 are on `rebuild_policy='owned'` pages** — 6 of them `tool-*`
> (`tool-cubic-bezier`, `tool-grid-generator`, `tool-json-cleaner`, `tool-noise-generator`,
> `tool-text-extractor`, `tool-head-architect`, plus `learn-design-physics-of-ui`). **And the new
> route, split by the same axis:**
>
> | `literal_markdown → page-rerender` | rows |
> |---|---|
> | **`generic`, complete** | **8** |
> | **`owned`, failed** | **1** |
>
> **Every success on the new route is a generic page; the only owned attempt failed.** The new route
> is refused by the ownership guard on owned pages — which is **this lane's own §7b warning**, now
> confirmed at n=1. So re-pointing the 7 would have produced **7 more failures**, dragged a healthy
> 8/1 pair toward its floor, and repaired nothing.
>
> **They are not accidentally stranded. They are the owned-page residual, deliberately routed here.**
> `184` closed correctly and sent exactly this class to `301`/tool-rebuilds — i.e. **to us**.
>
> **What survives, and it is still worth having:** the *mechanism* is real and measured — existing
> open rows keep the old `handler_agent`, cannot be dispatched (old pair held) and cannot be re-filed
> (dedup). **What was wrong is the inference** that therefore a working repair exists and only routing
> separates them. **Whether the new route can SERVE the old rows is a separate question, and I
> assumed it.** The corrected two-part check is in the `LANDMINES.md` entry (`50b8c65cf`).
>
> **And the corrected finding is sharper for us, not weaker:** these 7 rows are `277`'s subject, not
> `184`'s. An owned page with a real, mechanically-repairable defect has **no route at all** — the
> generic repair refuses it and nothing else claims it. That is the same hole as `no_content_data`
> (§0), reached from a different direction.

> **THE GENERAL FORM THAT SURVIVES: re-routing a producer fixes only FUTURE findings.** Every open row
> filed under the old literal keeps the old `handler_agent` for ever, and nothing warns you. **But the
> backlog is not automatically re-pointable** — first ask whether the new route would even accept
> those rows, split by whatever its guard keys on. If it would refuse them, the residual belongs to
> whoever owns the **blocker**, not the item type.

---

## 6. THE ESCALATION CLOCK — today's tick verified, and the INSTRUMENT has two defects

**Today's tick fired 12:58:16Z: `escalated=0, reclaimed=0, watching=13`.** §10.A predicted
`escalated=0, watching=15`. **The load-bearing half held** — *"ZERO IS CORRECT, not a failed
migration"*. The count differed because the pile had already dropped 15 → 13 (recorded in §7e), and it
was **12** by 16:15Z. The held set is not stable between a tick and your query.

### ⚠ `watching_detail` cannot be read the way §10.A reads it

```sql
string_agg(DISTINCT item_type||'->'||handler_agent||' ('||hold_kind
           ||', day '||(now()::date - created_at::date)||' of 3)', ', ') FROM classified
```

- **(i) `DISTINCT` is applied to a string containing the PER-ROW `created_at`, while the clock runs on
  `min(created_at)` per PAIR.** A pair spanning two dates prints as two entries at two day-counts.
  Confirmed: `count(DISTINCT created_at::date)` per pair is `2,1,1,1` and the readout shows **5 entries
  for 4 pairs**. **You cannot count held pairs off this line**, and the lower entry is a lie about the
  pair — `overdue` joins *every* row of the pair, so "day 1 of 3" rows escalate in the same tick as
  "day 3 of 3" ones.
- **(ii) The day counter is DATE arithmetic; the predicate is TIMESTAMP arithmetic.** So **"day 3 of 3"
  does NOT mean "fires this tick"** — `placeholder_contact` printed *day 3 of 3* at 12:58Z with a clock
  that did not expire until **19:17:45Z**. This is §10.A's own off-by-a-tick landmine **living inside
  the instrument**, which is why writing the warning did not prevent it.

### The real clocks — read these, not the readout [MEASURED 16:15Z]

| pair | kind | rows | oldest | clock expires | escalates at tick |
|---|---|---|---|---|---|
| `placeholder_contact → page-build-handler` | canary | **3** | 08-16 19:17:45 | **08-19 19:17:45Z** | **08-20 12:57** |
| `missing_conversion_path → content-gap-planner` | canary | 1 | 08-17 22:21:46 | 08-20 22:21:46Z | 08-21 12:57 |
| `dead_fragment_link → page-build-handler` | canary | 1 | 08-18 01:38:47 | 08-21 01:38:47Z | 08-21 12:57 |
| `literal_markdown → page-build-handler` | floor | **7** | 08-18 07:23:16 | 08-21 07:23:16Z | 08-21 12:57 |

Re-derive with the `pre_query`'s own `classified` CTE (`SELECT pre_query FROM scheduled_tasks WHERE
name='held-pair-canary-escalation'` — the column is **`name`**, not `task_name`).

**Every DATE in §10.A is confirmed.** The `literal_markdown` row count is stale in all three places
(§10.A says 10, §7f says 8, it is **7**) — **the population drains; re-derive at the tick.**

> ⚠ **§10.A's "first real escalation this mechanism has produced" needs one word.** All-history
> (live+archive), the escalation arm has fired **exactly once**: `page_component_status_drift →
> component-template-fixer`, **2026-08-17 21:51:33Z, 7 days waiting — under migration `453`**, which
> `466` superseded. So 08-20 would be **`466`'s** first, not the family's first. **And the good news
> nobody has written down: that escalated row went on to `complete`** — the escalate → human → repair
> loop has closed once, end to end.

⚠ **Do not canary `missing_conversion_path → content-gap-planner`** — `bugs_open/255` owns it.

---

## 7. STILL OWED, unchanged from the previous handoff

**§7a: the `copy_edit_proposed` exclusion in the promoter's `pre_query`, citing owner decision D2
(2026-08-12).** Not done, deliberately. It is a config change to a live shared scheduled task; needs a
numbered migration with a guard + `_ROLLBACK.sql`, exercised in a rolled-back transaction, **and it
goes past the owner — not in on a peer's say-so.** Full argument in the `LANDMINES.md` entry.

---

## 8. WHAT I WOULD DO NEXT

1. **The 7 `literal_markdown` rows before the 08-21 12:57 tick** (§5) — **and read §5's correction
   box first: they are OURS, not `184`'s, and they must NOT be re-pointed.** All 7 are owned-page
   items; the generic repair refuses owned pages. `184` closed today having explicitly routed this
   residual to `301`/tool-rebuilds. The open question is not routing but **what repairs a
   mechanically-fixable defect on an owned page at all** — the same hole as `no_content_data`. This
   is the only item with a clock; at minimum the 08-21 escalation should not read as a routing
   failure when it is a missing-repair one.
2. **Tell the `diagnosis_guardian` seat its `error_step` discipline is inverted** (§3). It will
   mis-fire on every correct submission until someone fixes it.
3. **~2026-08-25: close `083`** once `444`/`458`'s doors have held a week. Move with **both paths on
   the commit** and verify at HEAD with `git ls-tree`. **The close must state that `479`'s reclaim arm
   has never fired** — re-verified 0 all-history at 20:45Z tonight.
4. **`277`'s remaining half** — the `no_content_data` repair. Answer already in: **different agent**,
   do not design around `copy-editor`, and `473`'s deterministic route does not cover this class.
5. **`314`** — owner's call.
6. **`300`** — leave it. Demand-blocked, honest resting state; the first new drift finding proves it,
   and it will carry `result->'resolved_by'` when it does.

---

## 9. Session-start checklist
`git log --oneline -10` · re-read this from disk · `scripts/who-owns.py` **by slug** for `277`, `083`,
`300`, `314` (**not `301` — another session is closing it**) · re-probe §2 on whatever tag is live
**and check `distinct digests`** · re-derive §6's clocks from the `pre_query` · then §8.
