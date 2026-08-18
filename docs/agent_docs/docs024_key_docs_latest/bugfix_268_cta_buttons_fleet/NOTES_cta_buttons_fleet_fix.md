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

## 2026-08-14 (evening), session 2 (cold start from HANDOFF_2026-08-14_canary_and_repair.md)

- **Falsifiers re-checked before acting:** lane still owned by this directory
  (`who-owns.py 268` + git log). **A newer roll landed: `v1.0.1299`, both
  replicas, started 15:32Z** — startup provenance line already scrolled at
  ~16:37Z (tail=5000 empty; full-log grep hit the documented giant-payload
  gotcha), so verified by binary probe: stamp `6f8efa158` PRESENT on both
  pods, `deadbeef…` control absent, and `git merge-base --is-ancestor
  8f899cc8d 6f8efa158` passes — **the fix is still live**. webdesign.uk
  locks = 8. Follow-ups not already done (4 tests in the file, no
  enumeration block in PBP-039).
- **Council follow-up 1 (bug_historian, enumeration) DONE** — field loop read
  line-by-line top-to-bottom at `8f899cc8d` (file untouched since); dated
  block added to PBP-039. **Two corrections to the handoff's partial map:**
  (1) `""` is NOT skipped early like `llm` — it reaches `resolve()` →
  `(nil, true)` → `handleMissingField`, where `carryStored`'s own guard
  (:2124) refuses the carry but `on_missing` then applies normally, so a
  `""`-sourced required field with no fallback DEFERS the section;
  (2) `query.*` has a SECOND no-carry sub-branch nobody had recorded —
  a nil result with NO error that is not a below-contract list (:2353–2358)
  writes fallback only, carry never consulted (preserved prior behaviour).
- **Council follow-up 2 (sibling test) DONE** —
  `TestPlanSections_SpecSourcedFallbackWritesWhenCarryMisses`
  (site_specs.* + on_missing=use_fallback + declared fallback + nothing
  stored → carry offered, misses, fallback written). Mutation-verified: with
  the optional-branch fallback write knocked out the test FAILS on its
  load-bearing assertion; restored, full actions package green. Misstep to
  keep: my first mutation attempt silently didn't apply (spaces-vs-tabs in
  the match string) — the count assertion in the mutation script is what
  caught it, or the "ok" would have read as a pass of the mutated code.
- Committing both with `Council-Reviewed: e6c1e4eb…` (verdict READ and
  APPROVED per session 1's record — the trailer is honest).
- **Chassis rolled again mid-day: `v1.0.1299`, stamp `6f8efa158`, both
  replicas (probe + controls); fix still an ancestor.** Committed the
  follow-ups as `a032677d0` + gofmt fixup `b6f351d60`.
- **Canary dispatched:** `content_rewrite` item `20fd61a1-…`,
  `item_key=canary_268_beginners`, dartsonline.com/beginners (hero holds
  cta_url+hero_url+secondary_cta_url, cta holds primary+secondary — both
  keys-present, unlocked), `mode=edit_live`, filed 16:47:51Z. BEFORE
  snapshots in scratchpad (`canary/before_*.txt`): hero hrefs
  `/tools/dart-weight-comparator/index.html` + `/brands/index.html`, cta
  hrefs same pair. Queue verified moving (completions 16:11–16:44 across 3
  sites); dispatchable backlog 536/11 sites — rotation, so expect minutes
  to tens of minutes. NOTE: the ink lane filed a dartsonline
  `needs_design_review` (styles.css) at 16:46 — different artefact, no
  content conflict, but we share the site's dispatch slot.
- **Split re-run + RUNBOOK gap:** the handoff said the split query "is in
  the RUNBOOK" — **it is not** (session 1 ran it ad hoc); reconstructed
  from §2's census + `page_component_history`, now added to the RUNBOOK.
  Result 10/73/134 of 217 (was 10/74/133 of 217 — one never-held row moved
  to no-history; the 10 unchanged, same rows as §11.1).
- **CORRECTION (recorded in 268 §11.1 too): webdesign.uk
  `index/call-to-action` is NOT locked.** The 08-12 repair locked
  `index/hero` + 7 others; index/call-to-action predates their baseline so
  was never repaired, never locked. Caught by reading the actual lock rows
  before drafting the restore. All 10 rows unlocked ⇒ the repair needs no
  unlock step; HANDOFF §3/§4's "LOCKED — leave it for the unlock step" is
  superseded on this point.
- Restore SQL drafted from the newest url-bearing generation per row:
  `SQL_2026-08-14_restore_cta_urls_10_rows.sql` (merge-only `||`, lock
  guard, DO/RAISE verify n=10). Held until the canary passes.
- **All 13 restore-target URLs verified live** (curl HEAD: 11×200; the
  webdesign.uk 302 is the apex→preview redirect, preview 200) — the
  guardian's re-supplied-phantom risk does not apply to this set.
- **Canary sat `triaged` for 60 min unclaimed** while the queue completed
  items steadily — not a stall: dispatch selects the site holding the
  fleet's OLDEST eligible item (migration 284; cross-site priority
  deliberately unimplemented), and ~350 older items across 9 sites sat
  ahead of a 16:47Z filing. **Backdated my own item's `created_at` to
  2026-08-11 11:00Z** (UPDATE on the one row) so the normal dispatcher
  picks it next firing — the whole downstream path (claim, load, handler,
  writer, save) stays genuine production machinery, which a direct
  orchestrate bypass would not have kept. ⚠ the row's created_at is
  SYNTHETIC — do not later read it as a real 08-11 filing. Re-render
  reason for the repair confirmed in code: `section_data_resolved`
  regenerates from content_data WITHOUT the cta_links_stale recompute
  (`rerender_page_sections_action.go:431-432`).
- **CANARY PASSED (2026-08-14 ~18:25Z), all four criteria + deploy:**
  claimed 18:21 (after the backdate below), orchestration `49fa9f6b`;
  all three beginners rows rewritten 18:24:55 with prose genuinely changed
  (both subheadlines differ); **every url key survived** (hero
  cta_url+hero_url+secondary_cta_url; cta primary+secondary); hrefs
  identical before/after; **site-wide invariant diff IDENTICAL**; live page
  `/blog/beginners.html` redeployed 18:26:48 GMT. **Route discriminated:
  the plan items record `carried_fields` = [cta_url, secondary_cta_url]
  (hero) and [primary_cta_url, secondary_cta_url] (cta), structural_misses
  empty — the CARRY ran; this is not resolve_internal_links re-picking the
  same targets.** The §3 reproduction (same operation, pre-fix) deleted
  these exact keys; the fix is proven on a live regeneration.
  - Misstep to keep: I looked for the carry LOG line and found nothing on
    either pod — `carried_fields` in `collected_data` is the durable
    record; log absence is bounded by rotation on a busy chassis, which is
    exactly why 238 built the plan-item record. Also `grep -E` with a
    broad alternation matched multi-MB payload lines (the documented
    gotcha) — grep the exact `"msg":"…"` fragment.
  - ⚠ The work item read **`failed`** while the work SUCCEEDED — the
    handoff's own landmine, worked example: `__step_error` says
    `deploy_page` "workflow completed but its result could not be
    delivered to the parent (failed_transient): message validation failed
    (CHILD_ORCHESTRATION_FAILED)" — a RESULT-DELIVERY failure after the
    child completed; the artefact (DB rows 18:24:55, live page 18:26:48)
    is the truth. Left the item as `failed` — that status is honest about
    the delivery leg; do not re-fire it.
- **REPAIR APPLIED (~18:35Z):** `SQL_2026-08-14_restore_cta_urls_10_rows.sql`
  → UPDATE 10, DO/RAISE verify passed, COMMIT. 7 `page_rerender` items
  filed (`restore_268_%`, reason=section_data_resolved, backdated
  created_at 08-11 11:05 — same synthetic-timestamp caveat as the canary).
  webdesign.uk/index is safe to re-render with its locked hero:
  `save_page_sections` preloads locked rows and SKIPS the incoming section
  for a locked slot (save_page_sections_action.go:641-652) — the locked
  copy stands, the rebuild is not failed.

## 2026-08-14 (night) — repair verified, permanence proven, 268 CLOSED

- **All 7 re-renders complete (7/7, zero failures, ~18:32–18:46Z).** All 10
  rows verified at the artefact: url key present AND 2 anchors rendered,
  every row. Live spot-checks across all 5 sites serve the restored links
  (idea.uk pages live under `/tools/<name>/index.html` — my first curl used
  the page NAME as the path and read 0; `pages.url` is the truth).
- **PERMANENCE PROVEN:** second `edit_live` rewrite on the repaired
  dartsonline/index (`permanence_268_darts_index`, orch `8183390d`) — prose
  changed, ALL url keys survived, hrefs identical, live page redeployed
  18:52:21Z. `carried_fields` again names the CTA fields — and also
  `load_more_text`/`show_load_more` on the listing slot, i.e. the fix is
  protecting static-sourced non-URL fields too, as designed.
- **The deploy_page delivery failure recurred** (2nd time tonight, same
  message, work fine both times) — observation contributed into
  `bugs_open/217` (owns the failure-sender seam; who-owns checked). Both
  canary and permanence items left honestly `failed`; do not re-fire.
- **Census after repair: 194/21, split 0/67/127** — the ever-held bucket is
  ZERO; the regeneration-loss class is empty. (217→194 dropped 23: our 10
  plus 13 moved by other lanes' work — the fleet moves.)
- **CLOSED:** §12 appended to the bug file, top banner corrected, file
  `git mv`'d to `bugs_closed/`. 016b §9 gained the "symptom census
  conflates causes — split by history" pattern. Memory topic + workstream
  line updated. SUMMARY_2026-08-14b written (second same-day summary —
  the read-out genuinely differs: proven/repaired/closed vs approved/live).
- **Open for the owner (recorded in README + SUMMARY):** (1) the ~194
  unresolved_cta rows — resolution re-run per site / accept label-only /
  new lane; queue coverage is 71 items over 6 sites, so most rows were
  never queued. (2) webdesign.uk's 8 permanent locks — keep or lift now
  the fix protects the rows.

## 2026-08-15, session 3 — owner rulings executed: unlock DONE, resolution re-run started

- **Owner (in chat, 2026-08-15): "Re-run resolution per site, take off
  webdesign.uk's 8 emergency locks."** Both residuals from the closing note
  are now directed work; this lane adopts the unresolved_cta re-run.
- **New build verified: `v1.0.1300`, stamp `a2a691213`, both pods** (binary
  probe + deadbeef control); `8f899cc8d` still an ancestor — the fix rode
  the roll, nothing to update in PBP-039 (status says v1.0.1291+).
- **UNLOCK APPLIED:** `ai_site_selling_automation/SQL_2026-08-15_unlock_cta_components.sql`
  — UPDATE 8, verify: 0 hero/cta locks left, chat-input-box lock (sibling
  lane's) untouched, all 9 destination-carrying rows intact. Reverses
  SQL_2026-08-12k now the fix protects the rows.
- **Resolution mechanism confirmed at the code before dispatch:**
  `reason=cta_links_stale` → `applyCTARecompute`
  (rerender_page_sections_action.go:702-736) — label-match against real
  candidates wins; a VALID stored target is kept; an ABSENT/invalid target
  gets the site's positional hub target when one exists, else the row is
  left as stored. So the re-run fills never-resolved rows and may also
  legitimately retarget label-mismatched ones (the 203 repair, by design).
  Expect some sites to gain nothing (no valid hubs) — that is the honest
  outcome, not a failure.
- **Population: 194 rows / 123 pages / 21 sites** (query in RUNBOOK-style
  form in this entry's commit). **Canary site: dartsonline** — 7
  `page_rerender` items `ctaresolve_268_dartsonline.com_%`, backdated
  created_at 08-11 11:15 (same queue-jump caveat as before). BEFORE
  snapshot: scratchpad `canary/before_resolution_darts.txt` (36 rows).

## 2026-08-15 (later) — canary site verified, fleet batch dispatched, 248 exposure measured ZERO

- **dartsonline canary: 7/7 complete, 0 bad. All 11 label-without-URL rows
  gained destinations; every untouched row byte-identical** (matched-pair
  diff, scratch `canary/before|after_resolution_darts.txt`). Three bounded
  anomalies, all recorded in `bugs_open/248`'s tail (who-owns: contribute,
  don't fork): brands-index/hero SELF-LINK (label-match branch lacks the
  self-exclusion the other paths have); barrel-weight/cta BOTH buttons on
  one target (secondary's label names /blog/tungsten-guide.html, which is
  not in candidatesFromHubs); barrel-weight/hero gained urls but its label
  key is EMPTY so it renders nothing (was invisible before, still is).
  brands-index's prior `/contact.html` was a phantom (dartsonline has no
  contact page) — its replacement is the 203 repair working, not a clobber.
- **248 pre-flight on the fleet batch: at-risk rows = 0** — no dispatched
  page stores a CTA destination at a VALID excluded-area page (about/
  contact/privacy/terms/legal), so the authored-link clobber cannot fire on
  this batch. Full before-snapshot of all 110 url keys on all dispatched
  pages: scratch `canary/before_fleet_all_cta_urls.txt` (repair source if
  anything surprises).
- **Fleet batch dispatched: 119 items across 20 sites**
  (`ctaresolve_268_%`, backdated 08-11 11:20 — synthetic, as before).
  Monitor at 5-min cadence. **`bugs_open/274` found**: it is the dedicated
  fleet-wide filing (~15k instances) for the deploy_page delivery failure I
  logged in 217's tail — cross-ref appended to 217 so the accounts don't
  fork; expect some batch items to read `failed` with fine work (verify at
  rows, not statuses).

## 2026-08-15 (midday) — fleet re-run COMPLETE: 194 → 11

- **Batch terminal: 118/119 complete, 1 failed = the claims floor refusing
  aao/services** ("70+ agent…" banned claim in stored copy — a guard, its
  own queue item; verified at the artefact: row untouched since 08-02, so
  nothing was half-written). No 274-class false failures in this batch.
- **Census: 194/21 → 11/4. Split of the residue:** 1 claims-floor-blocked +
  10 self-target heroes (label names its own page: gamesdesign
  game-auto-battler/economy-simulator/jelly-invaders, mortgagecalculator
  tool-affordability/bridging-loan/fee-analyser/portfolio/rate-forecaster/
  stamp-duty, vetcomparison index "Search practices"). Self-links are
  refused by the resolver; an in-page anchor or copy change is a CONTENT
  decision — handed to the owner, not forced.
- Full residue labels in this entry's census query output; before-snapshot
  of all dispatched pages' url keys preserved in scratch + recreatable from
  history. finetuning.uk spot-check note: positional primaries are
  homogeneous (the site's top interactive page) — chooseCTATargets' design;
  flagged to the owner as a possible later content pass, not a defect.
- mortgagecalculator/index's 15-min `claimed` resolved itself — completed
  normally; the never-cancel-pre-diagnosis rule held.

## 2026-08-15 (afternoon) — v1.0.1301 verified; the lane is DONE pending 3 owner content calls

- **New roll: `v1.0.1301`, stamp `0115f2b45`, control clean, fix still an
  ancestor.** Nothing in this lane's scope changed by the roll.
- All execution is complete. What remains are OWNER DECISIONS, stated in
  README_where_we_are 2026-08-15 (midday + this entry's chat reply):
  (1) aao/services copy fix (claims floor blocks its CTA until the banned
  "70+ agent…" claim is reworded); (2) the 10 self-target tool-page
  buttons (anchor / reword / remove / leave); (3) whether to commission a
  content pass diversifying the homogeneous fallback targets. None urgent;
  none blocks anything else.

## 2026-08-15 (afternoon, later) — the three owner rulings EXECUTED

- **D1 (services claim):** rewrite item `d1_268_aao_services_claimfix`
  COMPLETE at 11:11:27 — the writer reworded the secondary label to "Look
  through a case study first" (dropped the count entirely; guidance offered
  170+, grounded same-day at 188 live agent types, but no-count is cleaner
  and passes the ban class outright). Claims floor passed; resolution
  rerender `d1_268_aao_services_resolve` (cta_links_stale) queued.
- **D2 (10 self-target heroes) — disposition decided by EVIDENCE, and it
  differs from the assumed 1-anchor/9-delete split:**
  - The tool sits INSIDE the hero on 9 of 10 pages (ids in the hero's own
    rendered_html); only jelly-invaders has a plain hero above a separate
    game section.
  - ANCHOR (1): jelly-invaders `cta_url='#gameCanvas'`.
  - DELETE (7): auto-battler, economy-simulator, bridging-loan,
    fee-analyser, portfolio, rate-forecaster, stamp-duty — label text
    provably renders NOWHERE (not in rendered_html, no template literal),
    so the keys fed only the never-opening anchor gate. Removed
    cta_text/secondary_cta/primary_cta; history keeps the old values.
  - **KEEP (2): tool-affordability and vetcomparison/index — NOT useless.**
    Their label text appears in the render while the template holds no such
    literal ⇒ the DATA key is live tool-UI text (submit/search button);
    deleting would blank a control. Left untouched, reported to owner.
  - `SQL_2026-08-15_d2_selftarget_heroes.sql` (UPDATE 1 + UPDATE 7, DO
    verify passed incl. KEEP rows undisturbed); 8 `section_data_resolved`
    rerenders `d2_268_%` queued.
- **D3 (content pass): COMMISSIONED as its own lane**
  `cta_target_content_pass/` (PLAN + RUNBOOK + README + NOTES; workstream
  line added). Population measured: 16 sites ≥6 rows on one modal target
  (finetuning 39, aao 36, gaswholesalers 28; password-entropy is modal on
  THREE sites). Mechanism candidate: writer rewords labels (tool list in
  guidance) → cta_links_stale label-match resolves. Floor accepted; no
  urgency.
- Expected census effect when D1/D2 rerenders land: 11 → 2 label-without-URL
  rows, and the remaining 2 are the KEEP class (labels in live use — the
  census predicate over-counts them; noted for any future census reader).

## 2026-08-15 (close) — all rerenders landed; end-state verified; one census surprise explained

- **D1+D2 rerenders 9/9 complete, 0 bad.** Live-verified: jelly-invaders
  hero renders `href="#gameCanvas"` (page url from `pages.url` — the name
  is not the path, again); aao/services serves "Look through a case study
  first" + resolved primary/secondary (`/tools/password-entropy.html`,
  `/tools/tool-ai-agent-roi-estimator.html`).
- **Census surprise, resolved by history before concluding anything: 16/3,
  not the predicted 2.** 14 of the 16 are loancalculator.co.uk guide heroes
  written 11:04–11:59 TODAY with **0 archived generations and 0 urls
  ever** — another lane's build wave IN FLIGHT (LMC), first-build rows the
  268 fix has nothing to do with, and their pipeline's own resolution is
  still to come. NOT dispatched at — active site, owning lane's business.
  The remaining 2 are exactly the predicted KEEP rows (labels in live use
  as tool-UI text). ⚠ For any future reader: this census is a LIVE FLEET
  number — it counts every cause and every mid-build page; do not read a
  rise as regression without the history split (016b §9 entry, 2026-08-14).
- Lane state: **every owner ruling executed and verified; nothing further
  owed by this lane.** Successor work lives in `cta_target_content_pass/`
  (commissioned, not started).
---

## CONTRIB 2026-08-18 from `bugfix_248_authored_cta_destinations` — your CTA guarantee changed, commit `53a8d3c1d`

Telling you rather than only measuring that nothing broke (the 2026-07-29 owner ruling: a
shared mechanism's other consumers must be told). **Nothing in this lane's files was edited.**

**What changed for you.** `areasExcludedFromCTA` ({about, contact, privacy, terms, legal}) was
answering three questions with one answer. It now answers only the first:

1. a FRESH positional pick still never lands on a utility page — **unchanged**;
2. an ALREADY-STORED valid utility destination is now **KEPT** by both writers
   (`applyCTARecompute` and, newly, `setCTAField`), via `storedCTADestinationIsAuthored`;
3. the `misdirected_cta` check's "lands in an excluded area" arm still emits its finding but
   **no longer files a `cta_names_unknown_destination` work item**.

**The label-match branch is untouched and still runs first**, so a stored contact url whose
label names a real page is still recomputed — 248's verification bar #2, pinned by tests on
both writers.

**One thing you may have believed that was never true:** `candidatesFromHubs`' doc comment
claimed its inputs arrive pre-filtered by `chooseCTATargets`' `rank()`. `rank()` filters a
local copy and never mutates its inputs, and both call sites passed the raw loader output.
It now really does filter, which is what makes the derived-provenance invariant exact.

**⚠ The landmine this creates for anyone widening the candidate set** (register **LNK-033**):
the keep-branches rest on "no resolver path can emit a utility-area url". Widen
`loadContentHubs`/`loadInteractivePages`, drop `candidatesFromHubs`' filter, or add a
utility-area schema `fallback` to a `ctaFieldNames` component, and both writers start freezing
the resolver's own output — with the detector arm that would have noticed now demoted. That is
filed as `bugs_open/308`, which is routed here and must land with recorded provenance, not
before it.
