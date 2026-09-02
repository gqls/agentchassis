# HANDOFF — 2026-09-02 — `bugs_open/394`, render-audit coverage cursor

**Supersedes `HANDOFF_2026-08-26_continue_here.md`** (same directory). Read this one; that one's
open list is four-fifths out of date.

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_394_render_audit_rotation_cursor/`
**Bug:** `bugs_open/394_HANDOFF_2026-08-25_webdesign_render_audit_tail_is_71_pages_and_growing_and_the_truncation_row_has_no_reader.md`
**Council:** `f67593f5-90cb-4a35-9cc0-926254645192` — **APPROVED** (round 3). Trailer already on `ea08c831d`.
**Ownership 2026-09-02:** still this lane; nobody else has touched the action or the lane dir in the week.

---

## 1. State in one paragraph

The fix is **live, running unattended on the scheduled rotation, and working**. Over the six days
I was away it produced correct cursor-mode windows on **two** sites, cycled one of them to
completion, and kept `design-critique-agent` in prefix mode as designed. Four of the five open
items from the last handoff are **closed**. What remains is one acceptance measurement (due today),
one CronJob that was never built, and one honest uncertainty about whether a specific commit rolled.

## 2. Closed since the last handoff — with the evidence, not an assumption

| item | status | evidence `[MEASURED 2026-09-02]` |
|---|---|---|
| (b) audit step TIMEOUT | **CLOSED — was transient** | `contrast_failure` rows created 08-27 **28**, 08-28 6, 08-29 2, 08-30 1, 09-01 2. Audits complete and findings are written again. |
| (c) optional-key overlay | **CLOSED — already applied** | live ConfigMap `optional-key-budget-check-script-9b89gcmd8g` line 179 reads `"request_render_audit": 7`. The cluster has the new literal. |
| mode split in production | **PROVEN** | `design-critique-agent` → `prefix`; `render-audit-agent` → `cursor`, on every row since the roll. |
| cycle completion on a real site | **PROVEN** | `loanandmortgagecalculator.co.uk` (61 live pages) ran window 1 on 08-27, then a **2-page final window** on 08-30 with `window_first == window_last` — the `cursor_cleared` branch firing in production. |

## 3. ⚠ OPEN — three things, in order

### (a) The acceptance union — ONE RUN AWAY, and it is due TODAY
The bug's own acceptance is "the union of audited pages reaches the whole site over a cycle".

`[MEASURED 2026-09-02]` union over the **scheduled** cursor runs on webdesign = **117 of 151**,
across 2 runs. webdesign is due again at **2026-09-02 12:35Z** (`last_selected_at` 08-30 12:35 +
3 days). Window 3 should close the cycle.

```sql
WITH sched AS (
  SELECT context FROM agent_error_log
  WHERE error_code='RENDER_AUDIT_TRUNCATED' AND domain='webdesign.co.uk'
    AND agent_type='render-audit-agent' AND occurred_at >= '2026-08-27'
    AND context->>'coverage_mode'='cursor')
SELECT count(DISTINCT p) AS union_pages,
       (SELECT max((context->>'pages_total')::int) FROM sched) AS total_now,
       (SELECT count(*) FROM sched) AS runs
FROM sched, jsonb_array_elements_text(context->'audited_paths') p;
```
⚠ **Grade it honestly:** `pages_total` was 151 and the site GROWS (131 → 146 → 151 in a week). A
page added mid-cycle may lawfully wait one cycle, so "union < total_now" is not automatically a
failure — compare the union against the pages that were live for the WHOLE cycle.

### (b) Is the identity fix `faf4872ce` actually rolled? Strong evidence, NOT proof
Running image `v1.0.1351`, pods started **2026-09-01 21:00Z**; the fix committed **2026-08-26
23:28+01:00**; `make build-*` builds from committed HEAD. The 08-30 scheduled run wrote its cursor
under `render-audit-agent`. So: almost certainly in.

**Why that is not proof:** `faf4872ce` changed a call site and comments and added **no new string
literal**, so no binary probe can discriminate it. `runningStepProvenance` being present only shows
the pre-existing function is there. And the 08-30 key is equally consistent with the OLD code if
`Sender.AgentType` already equalled `render-audit-agent` on the scheduled topic.

**The decisive probe — but run it AFTER window 3, or it eats the acceptance window:**
```bash
./docs/leopardessconsulting/scripts/orchestrate_safe.sh render-audit-agent \
  '{"site_id":"6b49db8e-d447-4467-8277-4f3018af9897","domain":"webdesign.co.uk"}'
# then:  SELECT agent_type FROM render_audit_page_cursor;
#   a NEW 'generic' row  -> the fix is NOT live
#   only 'render-audit-agent' -> it is
```
Housekeeping either way: the cursor table still carries an **orphaned `generic` row** from my
2026-08-26 hand-runs (`webdesign.co.uk | generic | 200 | tool-entropy-meter-guide`). It is honest
history, not damage. Delete it once (b) is settled.

### (c) The commissioned reader has no CronJob — the registry says `consumed` and nothing consumes it
`cmd/config-key-audit --render-truncation` exists, is tested, is mutation-proven on all four arms,
and `finding_code_registry.json` records it as the `consumed` reader. **But no CronJob runs it.**
`kubectl get cronjob` shows `component-render-check`, `optional-explicit-wires-check`,
`optional-key-budget-check` — and no `render-truncation-check`.

Clone the `ungraded-completions-check` kustomize service (same shape, same acks-file discipline).
Until then the registry's `consumed` claim is true about the code and false about the estate, which
is the exact shape `DBG-075` exists to prevent.

## 4. ⚠ Read before you re-probe a binary — a NEW landmine lands on last week's method

`LANDMINES.md` gained, 2026-08-24: **"BusyBox `grep` over `/proc/1/exe` reports FALSE ABSENCES —
and your present/absent controls PASS while it does it."** The fleet's images are BusyBox v1.37
(CLAUDE.md's "debian-slim" is stale); its grep is line-oriented and a Go binary's "line" can be
enormous, so a literal can read absent with a clean exit code.

**I used that instrument on 2026-08-26.** Assessed rather than waved away: the fault produces false
**ABSENCES**, and last week's claim rested on three **PRESENCES**, so the direction could only have
made me under-claim. The conclusion was also confirmed behaviourally minutes later. The negative
control may have been vacuous — that costs the control, not the conclusion.

Use the prescribed instrument from now on, both controls through the SAME pipeline:
```bash
kubectl -n ai-persona-system exec <pod> -- sh -c "tr '\0' '\n' < /proc/1/exe | grep -Fc '<literal>'"
```
Re-probed today on `agent-chassis-5bd89cf49-t4wdl`: `render_audit_page_cursor` 3 ·
`rotate_coverage` 2 · `runningStepProvenance` 1 · `selectAuditWindow` 2 · nonsense control 0.

## 5. Other traps this lane paid for (unchanged, still true)

- **`snapshot_agent` has TWO overloads writing to DIFFERENT TABLES**; a bare literal is ambiguous
  and aborts the migration; the two-arg form writes `agent_definitions_backup` and returns the
  SOURCE id, so verifying in `agent_definitions` reads 0 and looks like a no-op.
- **NEVER parse a `contrast_failure` `item_key`** — a selector may contain `#` and so may a page
  URL (`idea.uk` has `/tools.html` and `/tools.html#audience-check` both active). Match forward
  with `workItemKey(...)` + `HasPrefix`, longest path first.
- **`sql_for_agents` numbers are a shared sequence with no reservation** — mine was renumbered
  twice (646 → 649 → **660**). Take `max+1` at the moment you APPLY.
- **A test can only discriminate what its fixture varies** — `renderAuditParams` set
  `Sender.AgentType` and the running agent to the same literal, which made twelve tests blind to
  the key defect.

## 6. Facts, dated 2026-09-02

webdesign.co.uk **151** live pages (131 on 08-24 → 146 → 151; still growing). Bands: `0..90` 6 nav ·
`100` 94 tools · `200` 48 `tool-*-guide` · `201` 1. Callers: `render-audit-agent` cap 60
(rotating), `design-critique-agent` cap 8 (prefix by design, manual-only, acknowledged in
`render_truncation_acks.json`). Second truncating site: `loanandmortgagecalculator.co.uk`, 61 pages.
Cursor rows: 2 on webdesign (`render-audit-agent` live, `generic` orphaned).

## 7. Commit trail

`95a04168c` cursor · `72b16391b` R1 fix (forward match) · `41b03241d` reader + registry ·
`ea08c831d` R3 advisories + `Council-Reviewed` · `c71b46be0` migration 660 applied + landmine ·
`faf4872ce` identity fix · `a3610ea23` artefact proof · `99026097f` the week's evidence.
Migration: `docs/agent_docs/sql_for_agents/660_render_audit_coverage_cursor_HOLD.sql` (applied
2026-08-26 22:20:40Z, backup in `agent_definitions_backup`).

## 8. Also open, from this session's sweep (not 394)

- **`apis_uk_bees_homepage`** — their `SUBJECT_MISSING_ON_REPEATED_COMPONENT` registry entry puts
  its prose under `note` where the checker reads `why`, so `TestShippedRegistryIsSelfConsistent`
  was red at HEAD on 08-26. CONTRIB filed in their dir; worth re-checking whether it is still red.
- **`bugs_open/359`** — I validated it (7 of 39 archived pages serving) and yielded the lane; that
  session has since built the detector. Not mine.
