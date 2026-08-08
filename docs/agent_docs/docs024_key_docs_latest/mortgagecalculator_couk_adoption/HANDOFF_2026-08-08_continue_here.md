# HANDOFF — mortgagecalculator.co.uk COMPLETE ADOPTION — cold start, read this first

**Written 2026-08-08 ~10:30 UTC.** Supersedes `HANDOFF_2026-08-05_continue_here.md`
as the entry point (read it second — its §§2–4 checker-chain facts, §3 site-control
facts and §7 compare verdicts all still stand; only its §6 next-actions and §7 "three
produced nothing, unfinished, needs a fresh look" are superseded by this file).
Chassis at writing: **v1.0.1263** (pods 2026-08-08 08:54 UTC).

## 0. What changed since the 08-05 handoff — one paragraph

The three zero-output recreations are DIAGNOSED (they were the only open mystery):
two were falsely convicted by the placeholder validator reading JavaScript as prose,
one was held by the bug-020 fabrication gate. The validator is FIXED at source —
committed `201350e23`, council-submitted `a9ffed15`, **INERT until a
post-`201350e23` roll** (v1.0.1263 predates it). Full case: **`bugs_open/218`**
(also carries defect B: `validate_tool`'s failure routing is dead config, so failed
validations discard the recreation and complete the item — needs a deliberate
decision, do NOT hot-fix). Everything else — site locked, originals intact, 9 tools
live, queue drained — re-verified 08-07 and unchanged.

## 1. State right now

| thing | state |
|---|---|
| Live original site | intact — §10f sweep 2026-08-07: only `robots.txt` differs (Cloudflare, expected). If `data/latest-news.json` also differs, `git pull --rebase` in `~/projects/sites` first — it auto-commits daily ~08:10 UTC and a stale checkout reads as a diff |
| Site lock | **HELD** since 08-05 21:19 (`locked_by` = this lane). Keep held; recreation re-runs need a §10c unlock window |
| Queue | drained — 0 triaged/claimed on the site (08-07) |
| 9 rebuilt tools | live at new URLs, chrome present, wire-checked 08-07. `pages.build_status` still stale (`needs_rebuild`) for 7 — known, cosmetic |
| tool-overpayment, game-fact-finder | 404, 0 components — recreations were discarded by the validator false positive (`bugs_open/218` defect A). **Re-run AFTER the fix is live** |
| tool-portfolio | 404, 0 components — held by the bug-020 fabrication gate 08-05 12:55:38 (`needs_human_review` item on the site, still open). Signals unrecoverable; re-run and READ `check_fabrication` output from the live orchestration row before it purges |
| Validator fix | **round 2 landed** `b75f36601` + gofmt `f51ac6af8` (round 1 `201350e23` drew a council REVISE — a second stripper where `ExtractAssertionText` already existed; corrected same day, `WRONG_CALLS.md`). Scan now reads assertion-text blocks; `<no value>` exempted onto raw HTML (would otherwise go inert). **Round-2 verdict UNREAD** — `SELECT left(body,600) FROM diagnosis_artifacts WHERE correlation_id='a9ffed15-8e27-42a4-8ecd-e48f919470c9' AND kind='council_report' ORDER BY created_at DESC LIMIT 1;` — act on REVISE/REJECTED, the code is on the shared branch |
| 090 diagnosis | run `86721efd`: UNVERIFIABLE at iteration-cap; the two missing evidence items are supplied first-hand in `bugs_open/218` (stated-substitution per the 07-31 ruling) |
| Arithmetic verification | **still 0 of 12 proven** — unchanged from 08-05 §7; the id-renaming problem and path (a) stand |

## 2. Next actions, in order

1. **Confirm the fix is live** before anything else touches the tools. A roll is not
   evidence, and the tell is INVERTED because round 2 REMOVED round 1's symbol:
   `kubectl exec -n ai-persona-system <chassis-pod> -- sh -c
   'strings /app/agent-chassis | grep -c stripScriptAndStyle'` must be **0** on EVERY
   replica of an image that post-dates `f51ac6af8` (a 1 means the image carries
   round 1 only). Until then, re-running the two convicted recreations just
   re-convicts them.
2. **Council: DONE — round 4 APPROVED 2026-08-08** (trail: 3× REVISE, each
   answered same day, full read in `bugs_open/218`). Code ends at `35889819c`.
   The `Council-Submitted` trailers resolve to the approval automatically —
   nothing owed. **Still check the defect-B diagnosis verdict (run `c56b691d`,
   was `diagnosing` at hand-over)** — its outcome changes step 3's expectations
   for what a failed validation does, and its fix plan (not this lane) owns the
   save-anyway-vs-cannot-complete design call.
3. **Re-run the three recreations** (§10c unlock + backstop pattern — kill the
   backstop the moment the batch completes, §10g). Their `needs_tool_recreation`
   items are terminal-`complete`, so file fresh items (same spec shape — copy from
   the complete rows) or reset the rows; watch: `validate_tool` outcome for
   overpayment/fact-finder, `check_fabrication` output for portfolio (read it from
   `orchestration_states.collected_data` DURING/just after the run — it purges
   within ~a day). If portfolio trips the gate again, the signals name the reason;
   judge true/false positive THEN. Afterwards: components >0, pages deploy, wire 200,
   §10f sweep, re-lock.
4. **Then the id-alignment batch (08-05 handoff §7 path a)** — the 9 live rebuilds
   diverge from goldens on wholesale ID renaming; per-tool fix items stating "carry
   the original input/output ids verbatim; give the button an id"; re-run
   `acceptance/compare_rebuilt.py`; stamp-duty's £0-after-press stays UNRESOLVED
   until its compare comes back clean — do not adopt it as verified.
5. **Fences** re-emitted from id-complete rebuilds → hand to
   `staged_component_build` (they own PLAN/fence authoring; update the CONTRIB).
6. **Owner README** entry with outcomes (08-08 entry already covers the diagnosis).

## 3. Landmines active on this exact work

- **`complete` work item ≠ artefact** — bit this lane twice now; `bugs_open/218`
  defect B is the mechanism on the recreation path.
- **`site_work_items.result` on the recreation items describes the WRONG artefact**
  (cross-wired payloads; noted unowned in 218's tail). Don't trust it.
- The lock holds the QUEUE, not the site (08-03 handoff §3). Direct orchestrate
  publishes bypass it.
- `orchestration_states` purges completed runs within ~a day. Evidence you need from
  a run must be read the same day or it's gone (this cost us portfolio's signals).
- `agent_error_log`: query domain with `COALESCE(domain,'')`, never `IS NULL`; and
  the `agent_type` on validate rows can name the SHARED label, not the workflow that
  ran — join `work_item_id` to be sure.
- Chassis rolled twice during the diagnosis (1262, 1263); neither carries the fix.
  **Grep the binary, not the version number.**

## 4. Files of record

This dir: `NOTES` (08-08 entry = the diagnosis detail) · `README_where_we_are`
(owner log; 08-08 entry written) · `RUNBOOK` §10 (rebuild chain; §10f sweep, §10c/§10g
backstop rules) · `SUMMARY_2026-08-06_*` (milestone; still current — no new one
warranted until the re-runs land) · `acceptance/` (goldens, criteria,
compare_rebuilt.py). Bugs: **`bugs_open/218`** (this diagnosis; defect B open),
`bugs_open/191` (fixed live, file unmoved). Council runbook:
`fixloop_eg_dartsonline/RUNBOOK_council_gate.md`.
