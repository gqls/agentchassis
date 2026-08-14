# HANDOFF 2026-08-14 — bug 268: the fix is LIVE; you are the canary/repair session

Written by the session that built and shipped the fix. Read order: this file →
`NOTES_cta_buttons_fleet_fix.md` (the full evidence trail) → `bugs_open/268`
§11–§11.1 (what changed about the bug's own numbers). The original cold-start
(`HANDOFF_2026-08-13_start_here.md`) is superseded on repair sizing — its §5
assumed a 214-row history restore; §11.1 of the bug file corrects that.

## 0. State, with evidence

- **Fix committed `8f899cc8d`** — `carryStored()` inside the renderer/static
  branch of `plan_sections_action.go` (carry first; declared fallback only
  when nothing stored; early continue preserved). 4 tests in
  `plan_sections_renderer_carry_test.go`, mutation-verified.
- **Council APPROVED round 1**, corr `e6c1e4eb-69d5-4b02-93c4-742cc47315b2`
  (commit carries `Council-Submitted:`; 098 credits it automatically).
- **LIVE on `v1.0.1298`, both chassis replicas, stamp `bc39e7bf5`**
  (= descendant of the fix; verified 2026-08-14 ~14:00Z by binary probe with
  negative control — the startup log line scrolls after hours, so probe
  `/proc/1/exe` with CANDIDATE shas from the roll window, never a discovery
  grep).
- **Register:** PBP-039 extended in place (renderer/static extension block).
  **LANDMINES:** stored-beats-fallback entry added + synced. **WRONG_CALLS:**
  the "expect ~0" over-attribution logged.
- **Damage census** [MEASURED 08-14]: 217 label-without-URL components /
  20 sites = **10** real regeneration losses (recoverable from history; list
  in 268 §11.1) + **74** never-held-a-URL across archived generations +
  **133** no archived generation [INDETERMINATE]. Most of the fleet count is
  the old `unresolved_cta` never-resolved class — NOT this bug's damage.

## 1. First: the two council follow-ups (advisory, cheap, do them before canary)

1. **bug_historian's enumeration**: read planSection's field loop TOP to
   BOTTOM (`plan_sections_action.go` ~:2270–2385 at `8f899cc8d`) and record,
   in PBP-039's entry, every distinct `source` branch and its carry status.
   Partial map (verify, don't trust): `llm`/`""` → skipped by design, never
   carried · `query.*` → carry runs via `handleMissingField`, EXCEPT a query
   ERROR which writes fallback only (deliberate, bugs_open/054) ·
   `renderer`/`static`/dotted → the fixed branch, carry runs · all other
   sources → `resolve()` miss → `handleMissingField`, carry runs.
2. **The sibling test**: site_specs-sourced field with
   `on_missing=use_fallback`, declared fallback, nothing stored → carry
   misses, fallback written. Add to `plan_sections_renderer_carry_test.go`
   or the 238 file (same harness). Commit both with trailer
   `Council-Reviewed: e6c1e4eb-69d5-4b02-93c4-742cc47315b2` (verdict READ
   and APPROVED — that trailer is now honest).

## 2. Canary (prove the fix on a live regeneration)

- Pick ONE keys-present, UNLOCKED hero on a low-stakes site (webdesign.uk is
  locked — do not use). Selection query: hero/call-to-action rows where
  `content_data ? 'cta_url'`, site NOT webdesign.uk, page active; prefer
  dartsonline.com (its index/grip-styles rows are among the 10 damaged, so a
  keys-present page there pairs naturally with the repair proof).
- Take the invariant diff (RUNBOOK) BEFORE. Dispatch a real `content_rewrite`
  with **`mode=edit_live`** (without it the writer guts prose, bugs_open/178)
  scoped to that page. Take the invariant diff AFTER, as a matched pair.
- **Pass**: url keys survive in `content_data`, hrefs unchanged, prose
  actually rewritten, and the chassis log shows
  `plan_sections: non-llm field carried from stored content_data` for the
  cta fields (or `carried_fields` on the plan item).
- **Landmines**: no orchestration dispatch within ~300s of a chassis pod
  restart · fix 253's floor now REFUSES some saves — a refused save writes
  NOTHING and files a queue item; that is a guard, not the bug · a `failed`
  work item is not failed work (verify at the artefact both directions).

## 3. Repair (only after the canary passes)

- **The 10 history-recoverable rows** (list + dates in 268 §11.1; the split
  query is in the RUNBOOK — re-run it, the fleet moves): restore
  `content_data` from the last history generation carrying the url (crib
  `ai_site_selling_automation/SQL_2026-08-12d_restore_cta_urls.sql`), then
  re-render. Never patch `rendered_html`. **webdesign.uk
  `index/call-to-action` is LOCKED** — leave it for the unlock step.
- **Prove permanence on one repaired row**: after restore + re-render,
  dispatch one more `edit_live` rewrite and confirm the key SURVIVES — that
  is the end-to-end proof the fix + repair compose (repair-before-fix was the
  original trap; you are proving fix-then-repair holds).
- **The ~207 never-resolved rows are NOT yours to repair from history** —
  nothing to restore. They are the `unresolved_cta` class
  (`resolve_internal_links` filed items; check
  `site_work_items WHERE item_type='unresolved_cta'`). Scope it, then put
  the decision to the owner: re-run resolution per site (candidate 2 of the
  original handoff), accept label-only, or a new lane. Do not silently adopt
  it into this one.

## 4. Then, in order

1. Unlock webdesign.uk (filing lane's RUNBOOK in
   `ai_site_selling_automation/` has unlock/edit/relock) — restore its
   locked `index/call-to-action` url from history while unlocked; re-verify
   locks... then decide with the owner whether locks stay off (the fix now
   protects it) or go back on.
2. Re-run the census + the 10-row split; append figures to 268.
3. `bugs_open/268` moves to `bugs_closed/` only when fixed AND live AND
   the 10 repaired (the ~207 unresolved_cta rows are a separate deliverable
   and must not block closure — but say so explicitly in the closing note).
4. 016b §9: consider one entry — "a symptom census conflates causes; ask
   history whether each damaged row EVER held the value" (WRONG_CALLS
   2026-08-14 has the incident).

## 5. Falsifiers (check before trusting this file)

- `git log` on this directory + `bugs_open/268` (another session may have
  moved it); `who-owns.py 268` + live-transcript grep (a hit can be just a
  LANDMINES banner — read contexts).
- Chassis stamp: re-verify per SERVICE (`kubectl … grep -aq <sha>
  /proc/1/exe` with controls) — a newer roll may have landed; the fix cannot
  UN-ship (forward-only) but line numbers cited here can drift.
- webdesign.uk locks still 8 (query in RUNBOOK).
- Council follow-ups may already be done — grep this directory and the
  register entry before redoing them.
