# HANDOFF — 2026-09-02 (updated 13:3xZ) — `bugs_open/394`, render-audit coverage cursor

**Supersedes `HANDOFF_2026-08-26_continue_here.md`** (same directory). Read this one; that one's
open list is four-fifths out of date.

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_394_render_audit_rotation_cursor/`
**Bug:** `bugs_open/394_HANDOFF_2026-08-25_webdesign_render_audit_tail_is_71_pages_and_growing_and_the_truncation_row_has_no_reader.md`
**Council:** `f67593f5-90cb-4a35-9cc0-926254645192` — **APPROVED** (round 3). Trailer already on `ea08c831d`.
**Ownership 2026-09-02:** still this lane; nobody else has touched the action or the lane dir in the week.

---

## 1. State in one paragraph

**The cursor half is FIXED, LIVE and ACCEPTED.** The bug's own acceptance test passes outright:
over three consecutive **scheduled** runs the audit covered **151 distinct pages of 151 live —
zero missed**. The identity fix is confirmed live by a discriminating test. **Exactly one thing
remains: the commissioned reader has no CronJob**, and that is why `bugs_open/394` is still open
rather than closed. It is roughly one kustomize service of work.

## 2. Closed — with the evidence, not an assumption

| item | status | evidence `[MEASURED 2026-09-02]` |
|---|---|---|
| **the acceptance test** (bug §2) | **PASSES** | union of `audited_paths` over the 3 scheduled cursor runs = **151 distinct pages**; live pages = **151**; **missed = 0**. Graded against the site, not against itself. |
| cycle completion | **PROVEN in production** | run 3 (09-02 13:09Z): final window 37 pages, last page `tool-llm-cost-calculator` (`nav_order` 201), `cursor_cleared = true`, and the cursor row was then **gone** — `deleteAuditCursor` fired unattended |
| identity fix `faf4872ce` live | **PROVEN by a discriminating test** | the two hypotheses predicted opposite windows; observed `window_first = index` + a NEW `render-audit-agent` row, while the `generic` row was untouched. "Not live" refuted. |
| audit step TIMEOUT | **CLOSED — was transient** | `contrast_failure` rows created 08-27 **28**, 08-28 6, 08-29 2, 08-30 1, 09-01 2 |
| optional-key overlay | **CLOSED — already applied** | live ConfigMap `optional-key-budget-check-script-9b89gcmd8g` line 179: `"request_render_audit": 7` |
| mode split | **PROVEN in production** | `design-critique-agent` → `prefix`; `render-audit-agent` → `cursor`, every post-roll row |
| a second site cycling | **PROVEN** | `loanandmortgagecalculator.co.uk` (61 pages): window 1 on 08-27, 2-page final window on 08-30 |

**What this replaced**, for contrast: the same first 60 pages for ever, with 91 never audited —
including all 45 `tool-*-guide` pages, unreachable at any cap below 98.

## 3. ⚠ OPEN — ONE thing

### The commissioned reader has no CronJob — and that is the whole remaining task

`cmd/config-key-audit --render-truncation` is **built, tested, mutation-proven on all four arms**,
carries its acks file (`docs/agent_docs/docs024_key_docs_latest/architecture_review/render_truncation_acks.json`,
`design-critique-agent` acknowledged at birth with its reason), and is recorded in
`finding_code_registry.json` as the **`consumed`** reader for `RENDER_AUDIT_TRUNCATED`.

**Nothing runs it.** `kubectl -n ai-persona-system get cronjob` shows `component-render-check`,
`optional-explicit-wires-check`, `optional-key-budget-check` — and no `render-truncation-check`.

So the registry's `consumed` claim is **true about the code and false about the estate**, which is
precisely the state `DBG-075` exists to prevent, and it is why this bug is not closeable: the owner
commissioned two things on 2026-08-25 and only one of them is driven.

**How:** clone the `ungraded-completions-check` kustomize service — same shape, same acks
discipline, same `--report` mode that writes one `doc_notes` row per run whether clean or not.
Remember `RELEASE_IMAGES` / `AGENT_DEPLOY_SERVICES` membership, and that the check binary needs a
rebuild to carry `rendertruncation.go`.

Verify it the way the reader itself insists: a synthetic prefix-mode row from `render-audit-agent`
must go RED, and an empty `agent_error_log` must REFUSE rather than pass.

### Housekeeping already done, do not redo

The orphaned `generic` cursor row from my 2026-08-26 hand-runs was **deleted** on 2026-09-02 once
it was provably dead (nothing writes or reads that key any more). The table now holds exactly one
row: `webdesign.co.uk | render-audit-agent | 100 | tool-head-architect`, i.e. window 1 of a fresh
cycle started by the confirmation run.

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
