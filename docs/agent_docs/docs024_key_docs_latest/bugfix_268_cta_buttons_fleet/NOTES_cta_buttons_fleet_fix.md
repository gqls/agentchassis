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

## 2026-08-13 (later) — falsifiers, trace, owner decision

- **Falsifiers re-checked:** webdesign.uk locks HOLD (8 permanent on
  hero/call-to-action — query in RUNBOOK). **Census MOVED: 216/19 → 217/20**
  (label-without-URL components / sites, §2 query) — the leak is active;
  damage accrues while unfixed.
- **Code-path trace (Explore agent, full citations in the session; the five
  findings):**
  1. The early branch (`plan_sections_action.go:2361-2369`, today's numbers)
     is ORIGINAL design (`abf1c308a`, 2026-03-09, same commit that created the
     file) — predates `handleMissingField` (bugs_closed/044 drift extraction)
     and `carryStored` (238) by months. An optimisation, not a guard. TWO of
     its properties are retroactively load-bearing, documented in SQL not Go:
     (a) renderer/static short-circuit BEFORE required-field handling —
     `098_broaden_cta_recompute_source_unlock.sql:26-30` relies on
     `required:true` never blocking section readiness; (b) declared fallbacks
     write unconditionally, bypassing required/on_missing —
     `181_class_e_live_cta_url_integrity.sql:16-18` (097b's deliberate pin).
     So the fix must KEEP the branch, not route through `handleMissingField`.
     The SQL docs cite the branch at ~1187/1210-1218 — same branch, older
     line numbers; grep confirms exactly one such branch today.
  2. Carried-value data flow confirmed end-to-end: `resolvedData` →
     `item.ResolvedData` (:2445) → writer workflow → RenderComponentAction
     `merge_with` (`v3_site_actions.go:2020-2052` — writes BOTH
     `sectionContentData` and renderCtx) → `save_page_sections` INSERT
     (:964-967; DELETE at :787-790 — replace wholesale, no jsonb merge). A
     carried `cta_url` reaches the anchor AND the persisted row via existing
     machinery (PBP-014). No new seam.
  3. Rerender contrast confirmed in the file's own comments
     (`rerender_page_sections_action.go:482-489, :528-535, :496-499` "cannot
     LOSE a key"). `planSection`+`carryStored` already run there (no-op —
     reads the same rows it merges).
  4. Test coverage at the branch is ZERO — no test in the repo has ever
     declared a renderer/static-sourced field. Regression test from scratch.
  5. The four CTA url fields were flipped to `source:renderer` by migrations
     091/098 PRECISELY to keep them out of `resolved_data` (schema-derived
     recompute merges last and reverts authored edits — 089's Gauntlet
     retargets). **This looks like a blocker for the carry and is not:** the
     carry re-supplies the page's OWN last stored value (cannot fabricate —
     `storedFieldValue` refuses empties via `isEmptyContentValue`; cannot
     revert an edit — it IS the edit). Argument goes in the fix comment and
     council rationale. All other consumers checked and clear
     (`applyCTARecompute` reads stored not resolved; conflict log fires only
     on stored != fresh; envelope guard already reasons about merge_with).
- **OWNER DECISION (2026-08-13, in-session):** the carry extension ships
  **default ON** (unconditional in the branch, no opt-in flag). Divergence
  from the 2026-08-02 opt-in-OFF shape is stated in the council submission's
  risks for reviewers to judge — fix 253's precedent on this same pipeline.
- **090 FIRED (2026-08-14):** intake `95df3483-0291-48d3-992f-6453b5e8324f`,
  claimed by `diagnose-dispatch-loop`; **run correlation
  `38e53a03-ddcd-46c6-8533-d48510747758`** (artifacts key). Coverage check was
  CLEAR (no FORCE needed — 54 older open items on the site are >2h, FYI only).
  Commit of the fix is GATED on this verdict.

## 2026-08-14 — fix built, tests mutation-verified, register extended, council submitted

- **Fix written** (`plan_sections_action.go`, renderer/static branch): carry
  first, declared fallback only when nothing stored, early `continue`
  preserved. One line of logic + the block comment carrying the 091/098
  safety argument.
- **Tests** (`plan_sections_renderer_carry_test.go`, 4, reusing the 238
  harness): motivating carry · correct-or-absent + 098 readiness contract
  (required:true renderer field, no error-log write asserted by ABSENT
  sqlmock expectation) · 181/097b fallback pin · stored-beats-fallback (the
  one deliberate precedence change). **Mutation-verified both ways:** with
  the fix, all 4 pass; with the one line reverted, all 4 fail on their
  load-bearing assertions. Full actions package green; whole platform builds.
- **Composition with the owning writer verified** (`setCTAField`,
  `resolve_internal_links_action.go:308-336` + caller :191-193): writes
  unconditionally on a valid target, no skip-if-present guard — fresh
  resolution beats carried; unresolvable leaves the carry standing and files
  `unresolved_cta`. So the carry is the floor, never a block on recompute.
- **Register:** PBP-039 extended in place (dated block) + its stale
  status corrected visibly (said INERT; live since v1.0.1291 per the
  handoff's merge-base proof). Index row updated. Same-commit rule observed.
- **238 ownership re-checked before touching carryStored's surroundings:**
  last commit 2026-08-11, no live transcript activity (hits in
  `c9499322` are directory listings). Contribution routed into PBP-039's
  entry, pointer to be added to `bugs_open/238` tail in the fix commit.
- **Council SUBMITTED before commit:**
  `SUBMISSION_CORR=e6c1e4eb-69d5-4b02-93c4-742cc47315b2`. Commit will carry
  `Council-Submitted:` (verdict unread at commit time by design).
- 090 status at writing: iteration 3, `route`/EXECUTING_STEP 07:49Z. Waiting.

## 2026-08-14 (later) — 090 broke (no verdict); verification substituted; the census splits 10/74/133

- **090 outcome: infrastructure failure, not a verdict.** Verdict step died
  iteration 5 on `stop_reason=max_tokens` (32k output cap; bundles
  77→141KB); `max_attempts=1` → item `failed`. The loud refusal is
  `bugs_closed/076`'s guard working. Not re-filed on one occurrence.
- **Substituted first-hand verification (2026-07-31 ruling escape hatch),
  full statement in `bugs_open/268` §11.1:** in-vitro test reproduction +
  mutation; independent trace (branch predates carry, `abf1c308a`); history
  delete-archives read WITH the trigger's semantics (prosrc: archives OLD ⇒
  a delete row is the pre-rewrite state) — webdesign.uk index/hero carried
  its urls into BOTH rewrites (16:50:19, 20:34:58 archives).
- **Fleet split [MEASURED, disconfirmable — could have come out 200/17]:**
  of 217 damaged, **10** ever held a URL in history (real regeneration
  losses, all 08-11/12, listed in §11.1 — incl. webdesign.uk
  index/call-to-action, lost BEFORE the repair baseline, still locked), **74**
  never held one across archived generations, **133** have no archived
  generation [INDETERMINATE]. §2's census conflates the 268 mechanism with
  the older `unresolved_cta` never-resolved class. PLAN steps 5–6 corrected
  (repair is ~10 rows from history; census will NOT trend to ~0);
  WRONG_CALLS row added for my "expect ~0" claim.
- **Same-file passenger, incoming direction:** my PBP-039 index-row edit was
  swept into `504758a91` (RFC_022 lane, 08:46) while uncommitted. Nothing
  lost; noted in the fix commit message.
- Committing now: fix + tests + register + LANDMINES + WRONG_CALLS + workstream
  docs + both bug files, by pathspec, `Council-Submitted: e6c1e4eb…`.

## 2026-08-14 (afternoon) — committed, APPROVED, and LIVE

- **Committed `8f899cc8d`** (8 files — LANDMINES + WRONG_CALLS dropped out of
  the pathspec because OTHER sessions' commits `bb8bce65c`/`7252c2856` had
  already swept my appended entries in as same-file passengers; verified both
  present in HEAD; third passenger event this session, both directions now).
- **Council APPROVED round 1** (corr `e6c1e4eb…`, 13 reviews, 4 abstained,
  4 advisory objections, none high). Verdict landed 08:04Z — before the
  commit; `Council-Submitted:` was still the honest trailer (unread at commit
  time); 098 credits it automatically. **Advisory asks owed** (next session):
  - `bug_historian` (medium ×2): (1) enumerate EVERY distinct `source`
    branch in planSection's field loop and state each one's carry status —
    partial answer from reads so far: `llm`/`""` skipped early by design
    (never carried); `query.*` block → `queryListBelowContract` →
    `handleMissingField` (carry runs), except a query ERROR which
    deliberately bypasses on_missing and writes fallback only (054's trap);
    `renderer`/`static` (+dotted) → fixed branch (carry now runs); everything
    else → `resolve()` → `handleMissingField` on miss (carry runs). NOT yet
    verified line-by-line from the loop top — finish before recording as
    settled. (2) add a sibling-branch test: site_specs-sourced field,
    `on_missing=use_fallback`, declared fallback, nothing stored → carry
    misses, fallback written (note the pre-existing asymmetry: resolver
    fallbacks apply only under `use_fallback`; renderer/static fallbacks
    unconditional).
  - `guardian` (medium): default-ON fleet-wide on next roll — known,
    owner-decided, recorded. (low ×2: precedence-flip population check;
    tests fine.)
- **LIVE: `v1.0.1298`, both replicas, stamp `bc39e7bf5`** — descendant of
  `8f899cc8d` (`git merge-base --is-ancestor` passes; binary probe with
  negative control `deadbeef…` no-match; startup provenance line had
  scrolled after ~5h — the documented landmine — so the probe method was
  the grep of CANDIDATE shas from the roll window against `/proc/1/exe`).
- Next phase (fresh session): canary → 10-row history repair →
  unresolved_cta scoping → unlock webdesign.uk. See
  `HANDOFF_2026-08-14_canary_and_repair.md`.
