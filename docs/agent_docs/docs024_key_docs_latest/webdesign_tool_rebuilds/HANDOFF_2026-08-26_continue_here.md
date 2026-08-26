# HANDOFF — webdesign tool rebuilds. START HERE. Written 2026-08-26 ~15:45Z.
Supersedes `HANDOFF_2026-08-25_continue_here.md` (which had accumulated nine stacked STATE lines).

## STATE: 43 of 63 serve-confirmed. NOTHING IN FLIGHT. THE GRIND IS BLOCKED — do not file yet.

43 `removed` + 20 `deployed` = 63, verified 2026-08-26 15:38Z, with **zero pages carrying both a
live ported slot and a live native slot**. Nothing is part-done: no open `add_tool`, no pending
retire, no unwatched rerender, no owed serve-grade.

**Why blocked, and the ONE query that tells you when it is not.** `webdesign.co.uk` is queue-starved:
328 triaged build items, draining ~6 claims/hour, and a new `add_tool` at priority 60 sorts *after*
the 03:46 batch — so it would queue behind **~107** items. A build you cannot attend must not be
filed (an unattended build publishes a page serving BOTH tools). **Run GATE ZERO before anything:**

```sql
SELECT count(*) AS ahead_of_me FROM site_work_items
WHERE site_id='6b49db8e-d447-4467-8277-4f3018af9897' AND pipeline='build' AND status='triaged'
  AND (priority, created_at) < (60, now());
SELECT max(claimed_at) AS site_last_served FROM site_work_items
WHERE site_id='6b49db8e-d447-4467-8277-4f3018af9897';
```
More than a handful ahead, or a stale last-served, means **wait**. ⚠ **Do NOT bump `priority` and do
NOT re-file** — the per-site pickup is `priority ASC` (LOWER first), so a bump moves you BACKWARDS,
and `LANDMINES.md` forbids it. On 2026-08-26 a 12-hour census showing every claimed `add_tool` at
priority 120/130 and none at 60 looked like starvation-by-priority and was the causation inverted.

**#44 `tool-monolith-splitter` was filed and WITHDRAWN** (`910ea037`, cancelled under guards 10:25Z;
ported slot `e134edb7` verified still `deployed`). Its analysis is banked below — reuse it, do not
redo it.

## ⚠ STANDING HAZARD: 41 confirmed-false findings are queued against the 43 rebuilds

`tool_acceptance` has filed **41 `improve_tool` items against tools this lane rebuilt**, every one
reading *"interaction anchor #X absent from deployed page"*. **They are wrong, and the cause is
CONFIRMED** — `needs_diagnosis` `91228c39-8980-42bf-95cd-bd16bb43de0a`, complete 10:59:05:

- criteria are the tool's **authored PLAN document** (`check_tool_acceptance.go:loadCurrentCriteria`
  → `SELECT body FROM doc_plans WHERE subject_type='tool' AND subject_key=$1 AND is_current`),
  written with **bare** ids;
- the renderer rewrites every id: `ConvertTemplateToInstanceScope` does
  `strings.ReplaceAll(out, 'id="'+id+'"', 'id="'+instancePrefix+id+'"')`, with
  `InstanceToken` = `"c-" + s`.
- So `#ring-copy-button` is sought on a page carrying `id="c-tool-focus-ring-ring-copy-button"`.
  **The tell is not the matching names — it is that `boots` PASSES** (its selector `.tool-container`
  is a *class*, which nothing prefixes) **while every id-anchored check fails on the same page.**

**Fleet-wide: 110 anchor-absent findings against 2 of every other kind; 32 already `complete`.** Each
becomes an LLM rewrite: an observed `tool-improver` note reads *"Root cause: unknown. Fix: Rebuilt
tool HTML to restore the #sessions-per-day-input element"* — regenerating something never missing.

**Status: 41 still `triaged`, 66 ahead, NONE TOUCHED.** Owned by `staged_component_build`
(`scripts/who-owns.py tool_acceptance`), who hold the CONTRIB + verdict:
`docs/agent_docs/docs024_key_docs_latest/staged_component_build/CONTRIB_2026-08-26_from_webdesign_tool_rebuilds_tier2_anchors_are_unscoped_while_the_renderer_scopes_them.md`.
**Do not dispatch, promote or "fix" these rows, and do not let a queue-drain effort proceed without
reading this** — the stall is currently the only thing standing between them and the 43 rebuilds.
Raised with the owner as a decision; he asked for a re-check rather than choosing, so the contested
hold stays untaken. **If you take it, it is a reversible status flip on webdesign's 41 ONLY** — the
other 69 belong to other lanes' sites.

## The recipe (proven 43 times) — unchanged except GATE ZERO

1. GATE ZERO (above). Then: fetch the live page cache-busted and read the ported slot IN FULL.
2. Gates: library-claim (0 rows or pin fork identity), local active component (0), open `add_tool`
   (0), adopt flag `true`, margin on the page's queued rerender. `related_pages`: 1–3 EXISTING
   non-tool `pages.name` picked by TOPIC.
3. **Write the brief as a SPECIFICATION, not a bug report.** Describe the tool to BUILD; keep the
   defect archaeology in NOTES. Two builds died at `max_tokens` on 08-25 because the brief carried
   history — golden-ratio went 4,431 → 2,701 (still died) → **1,551 chars (built in 4m28s)**.
   Character count is NOT the predictor; surface area is.
4. File; attend with foreground poll loops (never a background watcher).
5. Grade the **RUN** (`page_adopted='true'`, no `already_exists`, no `__step_error` — an item reads
   `complete` with `error` NULL on a dead run), then **RETIRE IMMEDIATELY** (guarded txn, DO/RAISE
   pre- and post-asserts, md5 pinned, post-commit re-read).
6. Mechanism-grade the component at the DECIDING code arms — never the tool-doc header.
7. Serve-grade cache-busted: http=200 first, `last-modified > completed_at` (S3 lands 1–2 min AFTER
   the item completes), negatives 0, positives present. **Validate every negative BOTH ways** — it
   must count ≥1 in the ported bytes, or a 0 proves nothing.
8. Tombstone re-read. Dispatch the sidecar's dry-run retraction if the tool had one; record the
   orphan (`bugs_open/365`, 12 files across 11 tools).

## Next up — #44 `tool-monolith-splitter` (9,037), ANALYSIS ALREADY DONE

Page `05449406-4215-4c4a-9ffc-0fae8b83b7a0`, slot `e134edb7-60c9-4544-b392-0b00d9eb08dd`, md5
`79c32824…`, url `/tools/monolith-splitter/index.html`, self-contained (0 external scripts → no
orphan). Title `Monolith Splitter`. ⚠ Its ported slot carries a **hand-restored input column**
(comment dated 2026-07-29) — the port had dropped the whole left column, so the page said "define
your components" with nowhere to define them and `generateRefactorPrompt()` had no caller. The
rebuild replaces that patch. `related_pages`: `learn-operations-maintenance` +
`learn-operations-scaling`. **The 2,636-char brief is in `file_ms.sql` in the 08-26 scratchpad, or
re-derive from the defects below.**

Ported defects (sighting #22): **the instruction block goes stale silently** — it is built once on a
Generate press, so changing framework, file name or any component name afterwards leaves the panel
showing a prompt for values no longer on the page; the copy guard is
`text.includes("Define your components")`, i.e. keyed on placeholder PROSE, which returns silently
with no message and breaks if the copy is reworded; `copyPrompt` announces "Copied!" with no
`.catch`; its 2-second restore hardcodes `#333`/`#fff`, colours the button never had, so one copy
leaves it permanently mis-styled; `alert()` on the empty case; no way to REMOVE a component row;
4 inline `onclick` + 3 globals.

Then, smallest-first: head-architect 9,212 · asset-formatter 9,222 · layout-generator 9,223 ·
insight-injector 9,369 · … **re-run the census, do not trust this list.** The FIVE rich apps go
LAST, one at a time, owner-reviewed (standing ruling).

## Two items OWED — each wants its OWN `replace_existing` filing, not a fold-in

1. **cubic-bezier — keyboard access** (arrow-key nudge) for the two drag handles, cut to fit the
   token budget. A real gap on a site publishing `/learn/accessibility/focus-states.html`.
2. **golden-ratio — a REAL crop export.** The rebuild has **NO download at all**; the ported one had
   a "Download Crop" button that cropped nothing and burned the guides into the photograph. Wanted:
   cropped to the chosen ratio, centred on the overlay, guides NOT drawn on it.
   ⚠ **Both are `replace_existing:true` filings on rebuilt tool pages — the exact shape of the
   write-conflict spiral the noted lane found** (`bugfix_283_component_instance_scope/CONTRIB_2026-08-26_…`).
   Check strike history for the completed-then-overwritten pattern BEFORE either dispatches.

## Standing rules (load-bearing)

- ONE at a time (serial item key); file ONLY what you attend in-turn.
- Retire = status flip ONLY, never delete; revert handle = row id + length + md5 recorded pre-file.
- Grade the RUN, the COMPONENT (by mechanism), the SERVED page — never a status.
- **Any all-history claim about `site_work_items` must `UNION site_work_items_archive`** — the live
  table is a rolling window (cost me a published figure on 08-25: 13 `failed` was really 22).
- **Census by `spec->>'check'`, never by item_type name** — checks file under *other* type names by
  design, so an item_type census for `tool_acceptance` returns 0 and reads as "never ran".
- **`related_pages` mentions never land on this site** — `deferred` is a TERMINUS, not a gate: 0 of
  80 `tool_crosslink` rows have ever completed here, because every `/learn/` article is
  `rebuild_policy='owned'`. Keep filing the key (the finding is correctly targeted and is the raw
  material when an owned-page route exists); **never report a mention as delivered.**
- **Nothing may defer behavioural correctness to `tool_acceptance` while its findings are the
  anchor-absent class.** Grade at the mechanism and the served bytes, or say plainly it is ungraded.
- Counts carry the date they were counted (owner 08-22). A `[MEASURED]` figure about STATE expires.

## Cold-start dependencies

DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`
Site `6b49db8e-d447-4467-8277-4f3018af9897`. Tally: the RUNBOOK's `GROUP BY build_status` query
(expect `removed` = 43 + N). All per-tool ids/md5s: NOTES, per-tool entries (newest at bottom).
Chassis: rolled 2026-08-25 19:07Z, stamp `a7459a44b`; adopt path and the 360 tombstone guard both
verified present on both replicas (NOTES 08-25 20:30Z) — **re-verify after any further roll.**
