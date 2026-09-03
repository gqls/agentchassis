> ## ⚠ SUPERSEDED — the bug is CLOSED. Read `HANDOFF_2026-09-03_closed.md` in this directory.
>
> Nothing in this file's open list remains. What is left is three DECISIONS, laid out there.

# HANDOFF — 2026-09-02b — `bugs_open/394`: ONE verification from closing

**Supersedes `HANDOFF_2026-09-02_continue_here.md` and `HANDOFF_2026-08-26_continue_here.md`.**

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_394_render_audit_rotation_cursor/`
**Bug:** `bugs_open/394_HANDOFF_2026-08-25_webdesign_render_audit_tail_is_71_pages_and_growing_and_the_truncation_row_has_no_reader.md`
**Councils:** `f67593f5` APPROVED (cursor, r3) · `f49da30d` APPROVED (CronJob, 2 advisories, both answered)

---

## 1. WHAT IS LEFT — the whole answer

**One thing: the CronJob must fire ON ITS SCHEDULE once, at 07:50 UTC.** Everything else is done,
live and proven. That is the only outstanding item on this lane.

`[MEASURED 2026-09-02 16:18Z]` `render-truncation-check` — `SCHEDULE 50 7 * * *`,
`SUSPEND false`, **`LAST <none>`**. It has never fired on the clock, because it was deployed after
today's 07:50 had passed.

### The check, tomorrow (one query)
```sql
SELECT created_at, left(body, 300) FROM doc_notes
 WHERE categories ? 'render-truncation' ORDER BY created_at DESC LIMIT 3;
```
There is **one row now**, from the manual proving run at 16:17:57Z. **A SECOND row dated ~07:50Z is
the thing still owed.** The job writes one row per run, clean included, so an absent row means it
did not fire and must never read as "nothing is wrong".

Also confirm the slot did not collide once live:
```bash
kubectl -n ai-persona-system get cronjob -o custom-columns='NAME:.metadata.name,SCHEDULE:.spec.schedule,LAST:.status.lastScheduleTime' | sort -k2
```

### Then close it — and name BOTH paths
Both commissioned halves are then live AND exercised, which is the `/bugs_closed/` bar.
```bash
git add bugs_closed/394_HANDOFF_2026-08-25_....md
git commit bugs_open/394_HANDOFF_2026-08-25_....md bugs_closed/394_HANDOFF_2026-08-25_....md -m "close(394): ..."
git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 394   # must be exactly ONE line
```
A pathspec commit naming only the new path ships a **copy** and leaves the old file at HEAD; `ls`
cannot tell you, because the file is gone from disk either way.

## 2. Everything that IS done — with the evidence

| half | status | evidence `[MEASURED 2026-09-02]` |
|---|---|---|
| **cursor: fixed** | LIVE | binary probed with the prescribed `tr '\0' '\n' \| grep -Fc` instrument, both controls through the same pipeline |
| **cursor: accepted** | **PASSES** | union of `audited_paths` over 3 scheduled runs = **151 distinct pages**; live = **151**; **missed = 0** |
| cursor: cycle completion | PROVEN in production | run 3 (09-02 13:09Z) final window 37 pages → `tool-llm-cost-calculator` (`nav_order` 201), `cursor_cleared=true`, cursor row then deleted |
| cursor: identity fix | PROVEN by a discriminating test | two hypotheses predicted OPPOSITE windows; observed `window_first=index` + a new `render-audit-agent` row |
| a second site cycling | PROVEN | `loanandmortgagecalculator.co.uk` (61 pages): window 1 on 08-27, 2-page final window on 08-30 |
| mode split | PROVEN | `design-critique-agent` → `prefix`, `render-audit-agent` → `cursor`, every post-roll row |
| **reader: built** | in `v1.0.1354` | 4 alarm arms + dormancy + a not-live report line, all mutation-proven |
| **reader: deployed** | CronJob live, unsuspended | `render-truncation-check`, image `v1.0.1354` |
| **reader: proven at the pod** | **RAN, exit 0** | manual Job **from the CronJob**: 19 rows / 4 sites / 0 findings / 1 dormant named; `acks=/app/...json` (the in-image path); 61,872 rows read; `doc_notes` row written 16:17:57Z |
| registry `consumed` | done 08-26 | `41b03241d` — disposition/reader/reader_sink all set |

**What this replaced:** the same first 60 pages of webdesign.co.uk for ever, with 91 never audited —
including all 45 `tool-*-guide` pages, unreachable at any cap below 98.

## 3. Why the manual run does not already close it

It proves the **container**: the image pulls (no `ImagePullBackOff`, which this fleet reports as
RUNNING, never FAILED), the acks file really shipped in-image, the DB env is wired, the logic is
right against live data, and the durable row lands. It does **not** prove the **schedule** — cron
expression, timezone, and that nothing else contends for 07:50. That is a small residue, but it is
the difference between "the job works" and "the job runs", and this lane has spent two weeks on the
principle that those are different claims.

## 4. Traps this lane paid for — read before touching these

- **`snapshot_agent` has TWO overloads writing to DIFFERENT TABLES.** A bare literal is ambiguous
  and aborts the migration; the two-arg form writes `agent_definitions_backup` and returns the
  SOURCE id, so verifying in `agent_definitions` reads 0 and looks like a no-op.
- **NEVER parse a `contrast_failure` `item_key`** — a selector may contain `#` and so may a page URL
  (`idea.uk` has `/tools.html` and `/tools.html#audience-check` both active). Match forward with
  `workItemKey(...)` + `HasPrefix`, longest path first.
- **BusyBox `grep` over `/proc/1/exe` reports FALSE ABSENCES while both controls pass.** Use
  `tr '\0' '\n' | grep -Fc`, both controls through the SAME pipeline.
- **`sql_for_agents` numbers are a shared sequence with no reservation** — mine was renumbered twice
  (646 → 649 → **660**). Take `max+1` at the moment you APPLY.
- **A test can only discriminate what its fixture varies** — `renderAuditParams` set
  `Sender.AgentType` and the running agent to one literal, blinding twelve tests to the key defect.
- **Prove a mutation is not vacuous BEFORE trusting it**: build the mutated tree first (a build
  failure reads just like the red you wanted), and predict which assertion moves. Two of mine were
  vacuous on 09-02, eight days after I wrote the row about it.

## 5. Facts, dated 2026-09-02

webdesign.co.uk **151** live pages (131 on 08-24 → 146 → 151). Bands: `0..90` 6 nav · `100` 94 tools ·
`200` 48 `tool-*-guide` · `201` 1. Callers: `render-audit-agent` cap 60 (rotating, opted in by
migration **660**), `design-critique-agent` cap 8 (prefix by design, manual-only, acknowledged).
Second truncating site: `loanandmortgagecalculator.co.uk`, 61 pages. Dormant group:
`loancalculator.co.uk` (28 live pages, cannot truncate again). Cursor table: one row,
`render-audit-agent | 100 | tool-head-architect` — window 1 of a fresh cycle.

## 6. Do NOT chase these in the shared tree

1. **`go test ./cmd/config-key-audit/` may fail on `TestBudgetCronCountsLiteralMatchesTheRegistry`.**
   Another session has **uncommitted** work adding a 5th optional key (`max_image_dimension`) to
   `execute_vision_prompt`, plus untracked `vision_image_downscale.go`/`_test.go`. Their commit will
   hit that guard — that is the guard working. **Do not fix it; do not commit those files.**
   Confirm: `git status --short -- platform/orchestration/actions/ | grep -i vision`.
2. **`TestShippedRegistryIsSelfConsistent` is GREEN again.** It was red at HEAD for seven days on
   another lane's entry (prose under `note`; the human-evidence arm reads `why`). CONTRIBed 08-26,
   unacted for a week while blocking a clean package run, so fixed **additively** on 09-02 — new
   `why` field, their `note` untouched.

## 7. Commit trail

`95a04168c` cursor · `72b16391b` R1 fix · `41b03241d` reader + registry flip · `ea08c831d` R3
advisories + `Council-Reviewed` · `c71b46be0` migration 660 applied · `faf4872ce` identity fix ·
`a3610ea23` artefact proof · `667973351` acceptance passes · **CronJob commit** (dockerfile, base,
overlay, makefile, wrapper, dormancy) · `3749132e0` registry `why` · `ad8eb5f4f` pod proof.
Migration: `docs/agent_docs/sql_for_agents/660_render_audit_coverage_cursor_HOLD.sql` (applied
2026-08-26 22:20:40Z; backup in `agent_definitions_backup`).
