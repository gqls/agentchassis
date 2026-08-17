# HANDOFF — 2026-08-17c (evening), fresh chat starts here: v1.0.1307 SHIPPED properly, 454 proven, and ONE OPEN RISK that undermines the promoter's own rules

**Supersedes `HANDOFF_2026-08-17b_continue_here.md`** (same day, earlier — its §0 "fleet is 252
commits behind" is RESOLVED). Measured 2026-08-17 17:00–18:30Z.

## 0. The build DID ship this time — and the pair of rolls is now evidence

`v1.0.1307`, OCI `revision=a6d1c53c0`, `created=16:50:06Z`, pods up 17:05Z. Verified at the binary
with both controls in one `exec`: `a6d1c53c0` **PRESENT**, superseded `6a782274b` **ABSENT**.
**296 commits shipped, 31 touching Go.** Now 48 behind HEAD (3 touching Go) — ordinary churn.

The **only** difference between the afternoon roll that shipped nothing and this one was whether
`IMAGE_TAG` was bumped. Same makefile, same operator, same day. Recorded in `bugs_open/153` as
candidate 2's argument stated as an experiment. **The trap itself is still unfixed.**

## 1. What is LIVE and PROVEN

| thing | state |
|---|---|
| council `05a3d1c8` (promoter) / `7b0e2833` (router) | both **APPROVED**; both trails closed |
| `detected-item-promoter` (430 + 444 + **454**) | live, 900s |
| **`454` PROVEN 16:43:05Z** | the corrected floor released `empty_section → page-build-handler` (2 rows) on the **first tick after applying**. Disconfirmable both ways and neither fired: rows didn't stay held (no-op), and `literal_markdown` stayed held (floor not disabled) |
| `held-pair-canary-escalation` (453) | live, daily; first tick escalated the 4 seven-day rows, correctly left the 1-day pair |
| both owner decisions (2026-08-17) | **ANSWERED** — convert arms LEAVE AS IS; owner-per-type + 3-day limit DONE. Do not reopen |

## 2. ⚠ OPEN RISK — read before trusting either of the promoter's success tests

**Work-item history SHRANK today and I did not find the actor.** `required_fields_missing` read
`complete 64` at 11:00Z and `complete 50` at 18:30Z, `needs_human_review` unchanged at 31. Not
re-statused (`verified` = 0, the type has two statuses) — the rows left the table. Stable on
re-measure, so a discrete event.

**Not diagnosed, and deliberately not patched blind.** No `DELETE FROM site_work_items` in any
scheduled task, no retention window in `platform/`/`internal/`/`sql_for_agents/`, no CronJob naming
the table. Oldest surviving row fleet-wide is **2026-03-15** and 89 pre-August completes survive, so
it does not look like a blanket sweep — but `[UNDIAGNOSED]` is the honest label.

**Why it bites this lane:** 430's known-good rule and 444/454's floor **both read lifetime history**.
If completed rows can leave, "lifetime" is "whatever survives", and a pair that worked well but has
been quiet can read as *never having worked* → held for ever. Same latent failure `454` fixed,
arriving by a second route.

**Next session, in this order:** (1) `090` on *"completed `site_work_items` rows disappear; 14
`required_fields_missing` completes left the table between 11:00Z and 18:30Z 2026-08-17 with no
status change"*; (2) only then decide whether the tests need a durable per-pair tally that survives
row deletion rather than a live `COUNT` over the table. Full write-up: `bugs_open/083`, last section.

## 3. Migration-number collisions (mine, already applied — do NOT rename)

`453` and `454` each exist **twice** on disk and in the ledger:

```
453_held_pair_canary_escalation.sql          (mine, 12:58)  | 453_tool_recreation_whole_site_context.sql        (16:21)
454_promoter_counts_verified_as_success.sql  (mine, 16:35)  | 454_page_content_writer_honest_stop_word.sql      (16:19)
```

The documented number race — I took the next free number at the time and two other sessions landed
between. **Nothing is lost**: `schema_migrations`' PK is `filename`, so all four rows coexist and all
four are applied. **Do not renumber mine.** The ledger keys on filename, so a rename orphans the
ledger row, the file reads as *pending*, and the runner would re-apply it — and `454`'s verify block
has a positive control that would now fail (the pair it names has since been promoted).

## 4. Owed work, in priority order

1. **The §2 open risk** — `090` first, then decide. It is the only thing that can quietly undo this
   lane's mechanism.
2. **Watch `453`'s second discriminating test**: `placeholder_contact → page-build-handler` (created
   08-16, 0 lifetime successes) should escalate ~**2026-08-19**. `empty_section` should NOT
   reappear — after `454` it passes the floor and was promoted.
3. **`277` → `bugs_closed/`** ~**2026-08-22**: churn guard (day 2: one new row, born-detected →
   promoted → routed → parked, zero `unresolved`) + the two cancelled conversions re-raising.
   **Both paths on the commit** — LANDMINE.
4. **`083` → `bugs_closed/`** — blocked only by §2 now; criteria 1 and 3 hold, criterion 2 is met but
   **non-discriminating** (already true 6 days before the fix — do not bank it).
5. **Start the `router_engine` lane (RFC_030)** — NOT started; phase 1 (measurement) is done and
   recorded in that lane's NOTES. **Its phase 2 is a council design round on shape A vs B, BEFORE
   building** — that is the lane's own instruction and it is architecture scope.
   ⚠ **Its PLAN's guarantee 8 is STALE**: it calls RFC_022's accumulation counter "unbuilt", but
   CLAUDE.md now records RFC_022 **CLOSED**, the counter live since 2026-08-13 (`cmd/config-key-audit
   --optional-key-budget`, register WFA-013), owner-ruled `N=10`, with a daily CronJob
   (`optional-key-budget-check`). **This materially affects the A-vs-B choice** — shape B's per-type
   optional keys would be counted against that budget. Fix the PLAN before submitting.
6. **The council gate's config blind spot** — `097` scopes on `platform/`/`internal/`/`pkg/`, so a
   mechanism shipping as `agent_definitions`/`scheduled_tasks` config cannot be submitted. Round 2
   needed `FORCE=1`; round 5 passed only by being anchored on a Go file. Another lane filed a
   LANDMINE (`landmine(297)`); it wants a real fix.

## 5. Landmines this lane hit (all in LANDMINES.md / WRONG_CALLS.md)
- **`status` has TWO terminal success states** (`complete`, `verified`) — `GROUP BY status` before
  filtering on it.
- **`failed` rows carry NO `completed_at`** — a pair-health query keyed on it returns a uniform 100%.
- **The row set itself is not stable** (§2).
- **All three are ONE class**: the population measured was not the population assumed. None was
  *unmeasured* — each was **measured against an incomplete definition**, which no marker and no
  council round detects (12 seats approved `444`).
- **A same-tag rebuild ships the cached image** (§0); negative control must be a **real but
  different sha**, never a zeros run.
- **Check whether YOUR OWN guard strands YOUR OWN lane** — `444`'s allow-list passes
  `(build, content, design)`; this lane's producer files `content` (`:163`). Clean by measurement.
- **A verify block's control can be a tautology** — `453`'s first draft asserted
  `EXISTS(X) AND NOT EXISTS(X) = 0`. Replaced with an intersection of two independent sets.
- **A pathspec commit still takes a same-file passenger** (`b6bca52fc` carried another session's
  LANDMINES improvement; nothing lost, named in NOTES).

## 6. Session-start checklist
`git log --oneline -10` · re-read this file from disk · **verify the chassis revision before
trusting any Go-code claim** · `scripts/who-owns.py 277` / `083` (by SLUG — 083 is ambiguous) ·
re-measure §1 · then §4 item 1.
