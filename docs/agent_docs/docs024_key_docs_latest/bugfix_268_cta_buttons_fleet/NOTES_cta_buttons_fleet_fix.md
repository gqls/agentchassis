# NOTES — bugfix_268_cta_buttons_fleet (append-only, newest at the bottom)

## 2026-08-13, session 1 (cold start from HANDOFF_2026-08-13_start_here.md)

**Coordination check (per the handoff and MEMORY):**
- `who-owns.py 268` → owning workstream is this directory itself
  (`bugfix_268_cta_buttons_fleet`, 1 commit/14d = the handoff). Filing lane
  `ai_site_selling_automation` appended §10 to `bugs_open/268` today
  (`9bcd02caf`) — pointer only; its "gap-fills" line is that lane's own FAQ
  work, not this fix.
- Live-transcript grep (12 most recent `.jsonl`): one file with hits
  (`581eb30a…`), all three hits are its SessionStart LANDMINES banner matching
  the `bugs_open/` footprint plus a directory listing — **no other session is
  working 268.** Misstep to keep: my first conclusion from `hits=3` was "a
  lane is on it"; reading the hit contexts reversed that. The check is the
  CONTEXT of the hit, not the count.
- Diagnosis queue: no 268-shaped `needs_diagnosis` item open — the 090 re-run
  the handoff asks for has not been fired by anyone. Noted in passing: recent
  `needs_diagnosis` rows (08-08..08-12) all show `status='failed'` —
  handshake-race class per MEMORY (2 COMPLETED / 2 FAILED all-history); check
  `diagnosis_artifacts` by correlation before believing a `failed`.

**Chassis rolled mid-session (owner message):** both agent-chassis pods on
`v1.0.1295`, started 2026-08-13T13:53Z, provenance stamp `69612d692` (asked the
pod, not git). Handoff was written against `v1.0.1291`/`da5a7eb8f`;
`da5a7eb8f` IS an ancestor of the stamp. Only ONE commit in the window touches
the 268 code paths: `0c8e08ccb` (fix 253 — per-slot component floor in
`save_page_sections`, scope-gated to slots ≥10 class attributes; hero/CTA slots
are likely below it [INFERRED]). Adjacent, not protective, not an obstacle.

**Code-read at HEAD `a3fee59b8` (all three cited symbols verified present):**
- `plan_sections_action.go:622-626` — `resolve()` returns `(nil, true)` for
  `""|llm|renderer|static`, and `:637-639` same for `renderer.*`/`static.*`.
- `:2362-2369` — **the actual bypass [INFERRED, 090 pending]**: a
  renderer/static-sourced field takes this early branch, writes only a declared
  `fallback` (the four CTA url fields declare none) and `continue`s — so it
  never reaches `resolver.resolve` (:2372), never reaches `handleMissingField`
  (:2382), and the 238 carry never gets a look.
- `:2374` — if the early branch did NOT exist, `found && value != nil` would be
  false for `(nil, true)` and the field WOULD fall into `handleMissingField`.
- `:2123-2126` — `carryStored` guard excludes only `""`/`"llm"`. **The
  handoff's caution ("it may also skip renderer-sourced fields") is INVERTED**
  — a renderer field reaching the carry would be carried. Recorded as a
  correction in PLAN; the handoff file itself is the filing lane's, correction
  noted here rather than editing their file.

**Next act:** author + fire the 090 (mechanism as a question, pointed at
`page_component_history` windows 16:37–17:23Z and 20:20–20:45Z, live rows
repaired+locked stated). Then, while it queues (~30 min): falsifier checks
(locks=8, census moved?) and read `save_page_sections` replace semantics vs
`rerender_page_sections` merge semantics to have the fix sketch ready.
