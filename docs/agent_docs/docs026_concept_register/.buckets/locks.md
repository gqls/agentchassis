
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Section locking with lock types and expiry (design vs implementation gap)
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 031_LOCKS (investigation): "The columns don't exist in schema… expiry mechanism specced never built" while pass-reset half IS implemented; 004/007 describe lock_type permanent/timed/review as if landed
- **what:** Design: components that pass verification lock; lock_type permanent/timed(default 90d)/review(HITL on expiry) with query filter expansion `(locked_at IS NULL OR lock_expires_at < NOW())`. Reality: only plain locked_at/locked_by exists; auto-lock-on-deploy fires on every dashboard edit, so lock proliferation monotonically shrinks the improvement loop's surface (three documented failure modes). Recommended: timed default for routine edits, permanent opt-in.
- **sources:** 031_LOCKS_should_locks_expire.md; 004#Section Locking; 007#Lock lifecycle
- **relations:** audit pass auto-reset; lock coherence debt
- **verify-later:** lock_type/lock_expires_at columns exist?; discovery query filters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Lock semantics: hard gate for discovery, soft gate for execution, read-only rerender
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** 013 Phase 1 ✅ with the four amended checks; behaviour tables
- **what:** Lock means "human controls this", not read-only: edit refreshes locked_at without unlock; unlock is a separate deliberate act. Discovery checks skip locked rows (hard gate); execution agents process explicit items regardless (soft gate); rerender reads everything. locked_by vocabulary: admin/admin-removed/checkpoint (human-only unlock) vs deploy (agents may clear). Three lock levels: component, site component, whole site (site lock stops all automation via LoadWorkItemsAction gate + pre_query filter).
- **sources:** 013#Three Levels of Lock, #How Agents Behave; 031(3)#rules
- **relations:** growth budget; suppression
- **verify-later:** lock_helpers.go; four discovery checks' filters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Lock patterns A/B, Pattern B (pinned) is dead, and lock transfer across plan rebuilds
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 031(3) verification 2026-05-19: "Pattern B is unenforced in the current code — treat it as dead"; lock transfer specced for Phase 1 site_plan_directives
- **what:** Pattern A locked_at/locked_by (+partial index) is the dominant per-row pattern (sites, page_components, site_components, site_plan_directives). Pattern B pinned boolean on site_specs was never wired (no reads/writes; every spec write is supersede-then-insert with no guard) — new tables must use A. Lock transfer: only the rewriting agent (write_site_plan) copies locks onto matching new rows by composite key; locked text beats LLM rewrite; unmatched locks drop with a log. Locks and snapshots are orthogonal (prevention vs restore); open question whether revert respects locks.
- **sources:** 031_locks(3).md; 030 Q1 directives schema; 013 (pinned column added Phase 4 — UI-level only)
- **relations:** plan-domain tables; spec pin/propagate UI
- **verify-later:** \d site_specs pinned; write_site_plan lock-transfer code

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Adoption faithfulness via 90-day timed locks
- **category:** locks
- **status-signal:** partial
- **status-evidence:** "Status: design agreed (Option A, 90-day window). Schema migration written (053). Go follow-on pending" (2026-05-19); convergence layer marked [done]
- **what:** Adopted sites stay faithful to source for 90 days then develop normally — enforced as timed locks, not a permanent flag. Deliberately timed despite being user-initiated (a faithful starting point, not a frozen final value — documented so nobody "fixes" it to permanent). Because site_plan_directives are plan-scoped and adoption writes no plan, the lock originates at the FIRST write_site_plan (no-current-plan + pages-exist uniquely identifies adopted first plans): page-scoped preserve directives locked adoption/timed/90d; convergence (ValidateSitePlanAction) preserves whatever the 054 query flags adoption_locked; transferDirectiveLocks carries expiry across re-plans; after expiry everything is a no-op. Coexists with 30-day deploy locks at component scope (different questions, no contention).
- **sources:** FOCUS_adoption_faithfulness_via_locks(2).md (whole)
- **relations:** lock policy table; lock transfer; FOCUS_planner_ignores_adopted_state (the duplication this protects against)
- **verify-later:** 053/054 applied; write_site_plan first-plan lock branch; v3_site_actions.go convergence

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Lock policy table and the improvable-row predicate
- **category:** locks
- **status-signal:** partial
- **status-evidence:** "Approved policy table (with adoption added)" (2026-05-19); filter sweep of "11 locked_at IS NULL callsites" still pending
- **what:** Canonical lock semantics: human-set locks (admin/manual/checkpoint) permanent; auto-locks timed (deploy +30d on page_components; auditors +90d; adoption +90d on plan directives); audit_pending is not a lock. The improvable predicate — `locked_at IS NULL OR (lock_type='timed' AND lock_expires_at < NOW())` — must replace the 11 bare locked_at checks; CheckComponentLock to gain LockType/LockExpiresAt; expired review locks become needs_lock_review HITL items. Coherence rule: all four Pattern-A tables migrate in one migration, no partial state.
- **sources:** FOCUS_adoption_faithfulness_via_locks(2).md#policy, #predicate, #implementation-plan
- **relations:** adoption faithfulness; asset locking; Tension #3 candidate (lock-model coherence debt)
- **verify-later:** the 11 callsites; check_component_lock.go; expired_review_locks check existence

<!-- SOURCE: U05_content_quality_linking.md -->
### page_components locking subsystem + non-functional adoption re-plan window
- **category:** locks
- **status-signal:** partial
- **status-evidence:** HANDOFF_2026-06-15(2) schema corrections (lock columns + trigger exist, all NULL on index); running_notes_14(26) 14i: "90-day RE-PLAN window is non-functional".
- **what:** page_components (and assets, site_components, site_plan_directives) carry locked_at/locked_by/lock_type(permanent|timed|review)/lock_expires_at plus a trigger_auto_lock_on_deploy — but observed unlocked in practice on the investigated pages, and 013 doctrine says execution agents process explicit items regardless of locks. The adoption-faithfulness design's 90-day timed re-plan lock is non-functional: transferDirectiveLocks copies only locked_at/locked_by (no type/expiry) and nothing creates the adoption/timed lock — only the first-plan convergence branch works. Open question recorded: does save_page_sections honor locked_at (locking a tool section as a zero-code clobber mitigation — probably not).
- **sources:** HANDOFF_2026-06-15(2).md#schema-corrections; running_notes_14(26).md#part-14h-14i; NOTES(44) open sub-questions
- **relations:** adoption faithfulness convergence; interactive clobber mitigations; locks doc 031/053/054.
- **verify-later:** auto_lock_on_deploy function; transferDirectiveLocks; write_site_plan lock creation.

<!-- SOURCE: U09_adoption.md -->
### Adoption faithfulness via timed locks (90-day window)
- **category:** locks
- **status-signal:** partial
- **status-evidence:** "Verified landed state (2026-06-05)… 053 schema — applied… 054 partially live (first-plan branch only)… write_site_plan Changes 1-3 — not deployed… Consequence: the 90-day re-plan window is non-functional. The only working faithful↔normal boundary today is the first-plan branch."
- **what:** The faithful first pass after adoption is protected by a timed lock (`locked_by='adoption'`, `lock_type='timed'`, `lock_expires_at=NOW()+90d`) on page-scoped preserve-directives, self-releasing so the site later develops normally. Approved policy table adds `adoption` alongside deploy (+30d) and auditor (+90d) timed locks; human locks stay permanent. As landed, only the first-plan branch works; re-plans within 90 days rely on the LLM "preserve existing pages" prompt, not locks.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md, PLAN_lock_coherence.md, HANDOFF_2026-05-25#larger-work
- **relations:** first-plan branch detection; preserve-directive lock origination (pending); lock-model coherence plan; 031_locks docs
- **verify-later:** 053_lock_expiry.sql applied state (lock_type/lock_expires_at on assets, page_components, site_components, site_plan_directives); live `load_existing_pages` query; `transferDirectiveLocks` in write_site_plan_action.go (still copies locked_at/locked_by only?)

<!-- SOURCE: U09_adoption.md -->
### Adoption-side lock origination (superseded design)
- **category:** locks
- **status-signal:** superseded
- **status-evidence:** "REVISED 2026-05-19 after schema check: `site_plan_directives` is plan-scoped… adoption writes pages + specs but not plans or directives. So the lock cannot originate at adoption time" (FOCUS_adoption_faithfulness_via_locks(5)); the old2 base version still describes "Adoption writes a per-page preserve directive… locked locked_by='adoption'".
- **what:** The original design had adoption itself writing locked preserve-directives into site_plan_directives. Superseded because directives are keyed by plan_id and adoption creates no plan; the lock now originates at the planner's first `write_site_plan` (detected by `prevPlanID == uuid.Nil` AND existing pages present). There is no adoption-side Go change.
- **sources:** old2/FOCUS_adoption_faithfulness_via_locks.md, FOCUS_adoption_faithfulness_via_locks(5).md#how-this-drives
- **relations:** replaced by write_site_plan first-plan lock origination
- **verify-later:** confirm no adoption-side directive writes exist

<!-- SOURCE: U09_adoption.md -->
### write_site_plan preserve-directives + lock transfer patch (Changes 1–3)
- **category:** locks
- **status-signal:** aspirational
- **status-evidence:** "write_site_plan Changes 1-3 — not deployed. `transferDirectiveLocks` (verified) still copies locked_at/locked_by only… nothing emits page preserve directives or creates an adoption/timed/+90d lock" (2026-06-05).
- **what:** Three coordinated changes written as a patch doc but never deployed: (1) emit a page-scoped `preserve` directive per plan row; (2) on the first plan after adoption, lock those directives adoption/timed/90d; (3) extend `transferDirectiveLocks` to carry `lock_type` + `lock_expires_at` and skip already-expired timed locks (so expired locks release rather than chain forward). Needed only to protect re-plans within the window; re-prioritised low after the first-plan branch proved sufficient for the faithful first pass.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md#implementation, old2/write_site_plan_adoption_patch(1).md
- **relations:** adoption faithfulness via timed locks; lock coherence plan step 2
- **verify-later:** `write_site_plan_action.go` transferDirectiveLocks SELECT/UPDATE column lists

<!-- SOURCE: U09_adoption.md -->
### Lock-model coherence plan (one pattern, one lifecycle column, one predicate, one policy function)
- **category:** locks
- **status-signal:** aspirational
- **status-evidence:** "Status: PLAN ONLY — NOTHING IN THIS PLAN HAS BEEN APPLIED… held deliberately" (PLAN_lock_coherence, 2026-05-19).
- **what:** Collapse the accreted lock model: Pattern A everywhere (`locked_by` identity, `lock_type` permanent|timed|review, `lock_expires_at`), one improvable predicate, a single `LockPolicyFor(lockedBy)` policy function; retire Pattern B (`site_specs.pinned`, functionally dead in chassis code but exposed via core-manager pin/unpin HTTP endpoints) and the hard/soft `locked_by` string-switch in `check_component_lock.go` (`IsHard = lock_type=='permanent'`). Also resolves the snapshot×lock interaction (does revert_site_to_snapshot clobber human locks?). A fourth `lock_class` column was considered and dropped as redundant.
- **sources:** PLAN_lock_coherence.md, old2/PLAN_lock_coherence(2).md
- **relations:** 031_locks target model; adoption faithfulness runs on the current model without waiting for this
- **verify-later:** `check_component_lock.go` switch; `site_specs.pinned` column existence; `server.go` HandlePinSpec/HandleUnpinSpec; `\sf take_site_snapshot`/`revert_site_to_snapshot`; the 6 improvable-filter callsites vs locked-row finders (three distinct predicate semantics)

<!-- SOURCE: U10_imagery.md -->
### Asset locking (2A) and hard-vs-soft lock semantics
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** "2A — assets.locked_at + locked_by ✅ delivered 2026-05-09"; docstring "v3 (final, applied 2026-05-08)".
- **what:** `assets` gains `locked_at timestamptz` + `locked_by text` + partial index, mirroring `page_components` exactly. Canonical lock model (settled after three docstring iterations): detection via `locked_at IS NULL`; classification hard (admin/admin-removed/checkpoint) vs soft (deploy/manual/auditor names) via `locked_by`; NO time-based expiry exists in production. Human uploads/locked assets are excluded from auditor queries and regeneration.
- **sources:** STATUS_imagery_2026-05-08.md#Phase-2A, PLAN_imagery_loop_closure.md#2A, old/README.md
- **relations:** logo permanence (D5) is the first real consumer; timed lock-expiry project (deferred).
- **verify-later:** `check_component_lock.go`; assets table DDL; the store-asset lock guard `WHERE assets.locked_at IS NULL`.

<!-- SOURCE: U10_imagery.md -->
### Timed lock-expiry project (deferred)
- **category:** locks
- **status-signal:** aspirational
- **status-evidence:** "Approved policy (2026-05-08): implement timed expiry as a focused future project… Sequenced after the imagery loop work completes."
- **what:** One migration adding `lock_type` + `lock_expires_at` to all four Pattern A tables (page_components, site_components, site_plan_directives, assets); auto-lock writers default from a policy table ('admin' permanent, 'deploy' timed/30, auditor approvals timed/90); ~8–10 callsite filter expansions; CheckComponentLock extended; new `expired_review_locks` discovery check. Restores the rhythm doc 004 v4 designed, of which only the audit-pass-counter-reset half shipped.
- **sources:** old/README.md, STATUS_imagery_2026-05-08.md#Lock-expiry-investigation, PLAN_imagery_loop_closure.md#Decisions
- **relations:** references LOCKS_should_locks_expire.md (outside this unit); asset locking 2A.
- **verify-later:** whether lock_type/lock_expires_at columns exist on any Pattern A table.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Locks — HITL durability across the platform
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 031_locks(2) "This doc is the canonical reference for lock semantics"; "Tech debt: lock-model coherence (target model) … Status (2026-05-19): the lock model has accreted"
- **what:** Two per-row lock patterns protect human-edited data: Pattern A (`locked_at`+`locked_by`, dominant) and legacy Pattern B (`pinned` boolean on site_specs, don't use for new tables). Every writer must read lock state before writing and preserve it when superseding. A coherence cleanup to three orthogonal columns under the invariant permanent⟺human is recorded as deferred tech debt.
- **sources:** WM/031_locks(2).md#the-two-patterns-in-use, WM/031_locks(2).md#lock-transfer-across-rebuilds, WM/031_locks(2).md#tech-debt-lock-model-coherence-target-model, WM/030_phase1_plan_and_reconciler(4).md#lock-transfer-across-plan-rebuilds
- **relations:** human direction/lock lifecycle (007); adoption faithfulness via locks; site plan directives
- **verify-later:** migration 053; check_component_lock.go; FOCUS_adoption_faithfulness_via_locks.md; PLAN_lock_coherence.md

<!-- SOURCE: U18_sql_for_agents.md -->
### Section/component locking with timed expiry
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 115 header: "the project doc 004 v4 designed and docs 031_locks... approved (2026-05-08). Implemented now (Option A)... This migration is SCHEMA + BACKFILL only. The Go follow-on... lands as separate code changes."
- **what:** Locking is the improvement loop's termination and protection mechanism: verified/human-edited rows get locked_at set; auditors exclude locked rows (086). 115 adds lock_type + lock_expires_at to all four Pattern A lock-bearing tables (page_components, site_components, site_plan_directives, +1) in one transaction for coherence. Policy: admin/manual/checkpoint = permanent; deploy = timed +30d; visual-design-auditor / imagery-quality-auditor / adoption (new, faithful-first-pass) = timed +90d. Unlock predicate: `locked_at IS NULL OR (lock_type='timed' AND lock_expires_at < NOW())`. Go-side sweep of 11 callsites still pending at write time.
- **sources:** 115_locks.sql; 086_visual_design_auditor.sql
- **relations:** adoption faithfulness (FOCUS_adoption_faithfulness_via_locks.md); expired_review_locks discovery check (planned)
- **verify-later:** the 11 `locked_at IS NULL` callsites; CheckComponentLock extension; whether expiry sweep landed

<!-- SOURCE: U19_sql_tables_components.md -->
### Pattern A lock convention (locked_at / locked_by, hard vs soft)
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** 041 Phase 2A codifies: "a row is locked if locked_at IS NOT NULL. No time comparison... timed expiry is documented design intent (004 v4, 007 v4) but not implemented"; canonical classifier named (check_component_lock.go CheckComponentLock → IsHard).
- **what:** Uniform HITL/agent lock across four tables (page_components, site_components, assets, site_plan_directives — plus site_plan_imagery): locked_at timestamp + locked_by identity. Hard locks ('admin', 'admin-removed', 'checkpoint', 'manual' upload) only humans clear; soft locks ('deploy', auditor names, 'audit-pending') agents may clear when a work item references the row. Discovery skips both; execution skips hard. locked_by vocabulary is convention, not CHECK, to allow new identifiers without migration. A future lock-expiry project would add lock_type/lock_expires_at across all Pattern A tables in one migration.
- **sources:** docs/agent_docs/sql_for_tables/041_assets.sql#Phase2A; docs/agent_docs/sql_for_tables/012_site_components.sql#phase-7a; docs/agent_docs/sql_for_tables/040_site_plans_schema.sql#directives
- **relations:** 031_locks.md canonical doc; site-level lock; imagery/directive lock transfer.
- **verify-later:** CheckComponentLock consumers; lock-expiry project status.

<!-- SOURCE: U19_sql_tables_components.md -->
### Site-level lock (sites.locked_at)
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** "Phase 7: Site-level lock — prevents all automated agent activity" (012 tail); scheduled-task pre_query patched to exclude locked sites (020 site-lock section).
- **what:** locked_at/locked_by on sites acts as a master switch: when set, no automated agent activity (discovery, dispatch, improvement) touches the site. Scheduler pre_queries filter locked sites out of candidate selection.
- **sources:** docs/agent_docs/sql_for_tables/012_site_components.sql#phase-7; docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#site-lock
- **relations:** Pattern A locks; scheduler pre_query gating.
- **verify-later:** all dispatch/discovery entry points honour sites.locked_at.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Auto-lock on deploy (page_components lock trigger)
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** Schema captured 2026-07-01: "trigger_auto_lock_on_deploy auto-locks on deploy... lock_type permanent|timed|review"; lock check run pre-rebuild (all 4 index rows unlocked).
- **what:** page_components carries locked_at/lock_type/locked_by with a trigger that auto-locks components on deploy (fires on UPDATE). Operational consequence observed: deployed components MAY be locked, so rebuilds/re-renders must check lock state (a lock could block re-render of a target or protect neighbours); on the vonc index all rows were NULL-locked so the behaviour never actually bit in this corpus.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-01-~13:25 + #2026-07-01-~13:55; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§7
- **relations:** locks category (031); save_page_sections (does it honour locks? open question in 016b Part 4)
- **verify-later:** trigger_auto_lock_on_deploy definition; save_page_sections lock handling

<!-- SOURCE: U25_leopardess_social.md -->
### auto_lock_on_deploy trigger and the stillborn strict-mode subsystem
- **category:** locks
- **status-signal:** abandoned
- **status-evidence:** RUNNING_NOTES_minilobby (2026-07-09 record): "the strict-mode subsystem it belonged to was stillborn — no Go code reads schema_mode … snapshot columns never created … fired exactly once in the system's history"; dropped via migration 009 with saved reversal.
- **what:** A BEFORE UPDATE trigger stamping schema_mode='strict' + lock fields when a row reached deployed on first_deploy sites. Never functional as designed: save_page_sections INSERTs rows already deployed (trigger never fires), its companion snapshot columns were never created, and nothing reads the lock. It nearly sabotaged the section-editor fix (every edit would have locked its row) and was dropped 2026-07-10 with the function body backed up. schema_mode/strict_mode_trigger columns and the orphaned lock/unlock functions deliberately retained.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-09-the-dropped-trigger; docs/social001_vonc_tiktok_social/minilobby_task/auto_lock_on_deploy.FUNCTION_BACKUP.sql; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#8.2
- **relations:** locks category (031 lock semantics); build_status defect (the near-collision)
- **verify-later:** trigger absence on page_components; 009_drop_auto_lock_on_deploy.sql; leftover lock functions

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Section locking with lock types and expiry (design vs implementation gap)
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 031_LOCKS (investigation): "The columns don't exist in schema… expiry mechanism specced never built" while pass-reset half IS implemented; 004/007 describe lock_type permanent/timed/review as if landed
- **what:** Design: components that pass verification lock; lock_type permanent/timed(default 90d)/review(HITL on expiry) with query filter expansion `(locked_at IS NULL OR lock_expires_at < NOW())`. Reality: only plain locked_at/locked_by exists; auto-lock-on-deploy fires on every dashboard edit, so lock proliferation monotonically shrinks the improvement loop's surface (three documented failure modes). Recommended: timed default for routine edits, permanent opt-in.
- **sources:** 031_LOCKS_should_locks_expire.md; 004#Section Locking; 007#Lock lifecycle
- **relations:** audit pass auto-reset; lock coherence debt
- **verify-later:** lock_type/lock_expires_at columns exist?; discovery query filters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Lock semantics: hard gate for discovery, soft gate for execution, read-only rerender
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** 013 Phase 1 ✅ with the four amended checks; behaviour tables
- **what:** Lock means "human controls this", not read-only: edit refreshes locked_at without unlock; unlock is a separate deliberate act. Discovery checks skip locked rows (hard gate); execution agents process explicit items regardless (soft gate); rerender reads everything. locked_by vocabulary: admin/admin-removed/checkpoint (human-only unlock) vs deploy (agents may clear). Three lock levels: component, site component, whole site (site lock stops all automation via LoadWorkItemsAction gate + pre_query filter).
- **sources:** 013#Three Levels of Lock, #How Agents Behave; 031(3)#rules
- **relations:** growth budget; suppression
- **verify-later:** lock_helpers.go; four discovery checks' filters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Lock patterns A/B, Pattern B (pinned) is dead, and lock transfer across plan rebuilds
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 031(3) verification 2026-05-19: "Pattern B is unenforced in the current code — treat it as dead"; lock transfer specced for Phase 1 site_plan_directives
- **what:** Pattern A locked_at/locked_by (+partial index) is the dominant per-row pattern (sites, page_components, site_components, site_plan_directives). Pattern B pinned boolean on site_specs was never wired (no reads/writes; every spec write is supersede-then-insert with no guard) — new tables must use A. Lock transfer: only the rewriting agent (write_site_plan) copies locks onto matching new rows by composite key; locked text beats LLM rewrite; unmatched locks drop with a log. Locks and snapshots are orthogonal (prevention vs restore); open question whether revert respects locks.
- **sources:** 031_locks(3).md; 030 Q1 directives schema; 013 (pinned column added Phase 4 — UI-level only)
- **relations:** plan-domain tables; spec pin/propagate UI
- **verify-later:** \d site_specs pinned; write_site_plan lock-transfer code

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Adoption faithfulness via 90-day timed locks
- **category:** locks
- **status-signal:** partial
- **status-evidence:** "Status: design agreed (Option A, 90-day window). Schema migration written (053). Go follow-on pending" (2026-05-19); convergence layer marked [done]
- **what:** Adopted sites stay faithful to source for 90 days then develop normally — enforced as timed locks, not a permanent flag. Deliberately timed despite being user-initiated (a faithful starting point, not a frozen final value — documented so nobody "fixes" it to permanent). Because site_plan_directives are plan-scoped and adoption writes no plan, the lock originates at the FIRST write_site_plan (no-current-plan + pages-exist uniquely identifies adopted first plans): page-scoped preserve directives locked adoption/timed/90d; convergence (ValidateSitePlanAction) preserves whatever the 054 query flags adoption_locked; transferDirectiveLocks carries expiry across re-plans; after expiry everything is a no-op. Coexists with 30-day deploy locks at component scope (different questions, no contention).
- **sources:** FOCUS_adoption_faithfulness_via_locks(2).md (whole)
- **relations:** lock policy table; lock transfer; FOCUS_planner_ignores_adopted_state (the duplication this protects against)
- **verify-later:** 053/054 applied; write_site_plan first-plan lock branch; v3_site_actions.go convergence

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Lock policy table and the improvable-row predicate
- **category:** locks
- **status-signal:** partial
- **status-evidence:** "Approved policy table (with adoption added)" (2026-05-19); filter sweep of "11 locked_at IS NULL callsites" still pending
- **what:** Canonical lock semantics: human-set locks (admin/manual/checkpoint) permanent; auto-locks timed (deploy +30d on page_components; auditors +90d; adoption +90d on plan directives); audit_pending is not a lock. The improvable predicate — `locked_at IS NULL OR (lock_type='timed' AND lock_expires_at < NOW())` — must replace the 11 bare locked_at checks; CheckComponentLock to gain LockType/LockExpiresAt; expired review locks become needs_lock_review HITL items. Coherence rule: all four Pattern-A tables migrate in one migration, no partial state.
- **sources:** FOCUS_adoption_faithfulness_via_locks(2).md#policy, #predicate, #implementation-plan
- **relations:** adoption faithfulness; asset locking; Tension #3 candidate (lock-model coherence debt)
- **verify-later:** the 11 callsites; check_component_lock.go; expired_review_locks check existence

<!-- SOURCE: U05_content_quality_linking.md -->
### page_components locking subsystem + non-functional adoption re-plan window
- **category:** locks
- **status-signal:** partial
- **status-evidence:** HANDOFF_2026-06-15(2) schema corrections (lock columns + trigger exist, all NULL on index); running_notes_14(26) 14i: "90-day RE-PLAN window is non-functional".
- **what:** page_components (and assets, site_components, site_plan_directives) carry locked_at/locked_by/lock_type(permanent|timed|review)/lock_expires_at plus a trigger_auto_lock_on_deploy — but observed unlocked in practice on the investigated pages, and 013 doctrine says execution agents process explicit items regardless of locks. The adoption-faithfulness design's 90-day timed re-plan lock is non-functional: transferDirectiveLocks copies only locked_at/locked_by (no type/expiry) and nothing creates the adoption/timed lock — only the first-plan convergence branch works. Open question recorded: does save_page_sections honor locked_at (locking a tool section as a zero-code clobber mitigation — probably not).
- **sources:** HANDOFF_2026-06-15(2).md#schema-corrections; running_notes_14(26).md#part-14h-14i; NOTES(44) open sub-questions
- **relations:** adoption faithfulness convergence; interactive clobber mitigations; locks doc 031/053/054.
- **verify-later:** auto_lock_on_deploy function; transferDirectiveLocks; write_site_plan lock creation.

<!-- SOURCE: U09_adoption.md -->
### Adoption faithfulness via timed locks (90-day window)
- **category:** locks
- **status-signal:** partial
- **status-evidence:** "Verified landed state (2026-06-05)… 053 schema — applied… 054 partially live (first-plan branch only)… write_site_plan Changes 1-3 — not deployed… Consequence: the 90-day re-plan window is non-functional. The only working faithful↔normal boundary today is the first-plan branch."
- **what:** The faithful first pass after adoption is protected by a timed lock (`locked_by='adoption'`, `lock_type='timed'`, `lock_expires_at=NOW()+90d`) on page-scoped preserve-directives, self-releasing so the site later develops normally. Approved policy table adds `adoption` alongside deploy (+30d) and auditor (+90d) timed locks; human locks stay permanent. As landed, only the first-plan branch works; re-plans within 90 days rely on the LLM "preserve existing pages" prompt, not locks.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md, PLAN_lock_coherence.md, HANDOFF_2026-05-25#larger-work
- **relations:** first-plan branch detection; preserve-directive lock origination (pending); lock-model coherence plan; 031_locks docs
- **verify-later:** 053_lock_expiry.sql applied state (lock_type/lock_expires_at on assets, page_components, site_components, site_plan_directives); live `load_existing_pages` query; `transferDirectiveLocks` in write_site_plan_action.go (still copies locked_at/locked_by only?)

<!-- SOURCE: U09_adoption.md -->
### Adoption-side lock origination (superseded design)
- **category:** locks
- **status-signal:** superseded
- **status-evidence:** "REVISED 2026-05-19 after schema check: `site_plan_directives` is plan-scoped… adoption writes pages + specs but not plans or directives. So the lock cannot originate at adoption time" (FOCUS_adoption_faithfulness_via_locks(5)); the old2 base version still describes "Adoption writes a per-page preserve directive… locked locked_by='adoption'".
- **what:** The original design had adoption itself writing locked preserve-directives into site_plan_directives. Superseded because directives are keyed by plan_id and adoption creates no plan; the lock now originates at the planner's first `write_site_plan` (detected by `prevPlanID == uuid.Nil` AND existing pages present). There is no adoption-side Go change.
- **sources:** old2/FOCUS_adoption_faithfulness_via_locks.md, FOCUS_adoption_faithfulness_via_locks(5).md#how-this-drives
- **relations:** replaced by write_site_plan first-plan lock origination
- **verify-later:** confirm no adoption-side directive writes exist

<!-- SOURCE: U09_adoption.md -->
### write_site_plan preserve-directives + lock transfer patch (Changes 1–3)
- **category:** locks
- **status-signal:** aspirational
- **status-evidence:** "write_site_plan Changes 1-3 — not deployed. `transferDirectiveLocks` (verified) still copies locked_at/locked_by only… nothing emits page preserve directives or creates an adoption/timed/+90d lock" (2026-06-05).
- **what:** Three coordinated changes written as a patch doc but never deployed: (1) emit a page-scoped `preserve` directive per plan row; (2) on the first plan after adoption, lock those directives adoption/timed/90d; (3) extend `transferDirectiveLocks` to carry `lock_type` + `lock_expires_at` and skip already-expired timed locks (so expired locks release rather than chain forward). Needed only to protect re-plans within the window; re-prioritised low after the first-plan branch proved sufficient for the faithful first pass.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md#implementation, old2/write_site_plan_adoption_patch(1).md
- **relations:** adoption faithfulness via timed locks; lock coherence plan step 2
- **verify-later:** `write_site_plan_action.go` transferDirectiveLocks SELECT/UPDATE column lists

<!-- SOURCE: U09_adoption.md -->
### Lock-model coherence plan (one pattern, one lifecycle column, one predicate, one policy function)
- **category:** locks
- **status-signal:** aspirational
- **status-evidence:** "Status: PLAN ONLY — NOTHING IN THIS PLAN HAS BEEN APPLIED… held deliberately" (PLAN_lock_coherence, 2026-05-19).
- **what:** Collapse the accreted lock model: Pattern A everywhere (`locked_by` identity, `lock_type` permanent|timed|review, `lock_expires_at`), one improvable predicate, a single `LockPolicyFor(lockedBy)` policy function; retire Pattern B (`site_specs.pinned`, functionally dead in chassis code but exposed via core-manager pin/unpin HTTP endpoints) and the hard/soft `locked_by` string-switch in `check_component_lock.go` (`IsHard = lock_type=='permanent'`). Also resolves the snapshot×lock interaction (does revert_site_to_snapshot clobber human locks?). A fourth `lock_class` column was considered and dropped as redundant.
- **sources:** PLAN_lock_coherence.md, old2/PLAN_lock_coherence(2).md
- **relations:** 031_locks target model; adoption faithfulness runs on the current model without waiting for this
- **verify-later:** `check_component_lock.go` switch; `site_specs.pinned` column existence; `server.go` HandlePinSpec/HandleUnpinSpec; `\sf take_site_snapshot`/`revert_site_to_snapshot`; the 6 improvable-filter callsites vs locked-row finders (three distinct predicate semantics)

<!-- SOURCE: U10_imagery.md -->
### Asset locking (2A) and hard-vs-soft lock semantics
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** "2A — assets.locked_at + locked_by ✅ delivered 2026-05-09"; docstring "v3 (final, applied 2026-05-08)".
- **what:** `assets` gains `locked_at timestamptz` + `locked_by text` + partial index, mirroring `page_components` exactly. Canonical lock model (settled after three docstring iterations): detection via `locked_at IS NULL`; classification hard (admin/admin-removed/checkpoint) vs soft (deploy/manual/auditor names) via `locked_by`; NO time-based expiry exists in production. Human uploads/locked assets are excluded from auditor queries and regeneration.
- **sources:** STATUS_imagery_2026-05-08.md#Phase-2A, PLAN_imagery_loop_closure.md#2A, old/README.md
- **relations:** logo permanence (D5) is the first real consumer; timed lock-expiry project (deferred).
- **verify-later:** `check_component_lock.go`; assets table DDL; the store-asset lock guard `WHERE assets.locked_at IS NULL`.

<!-- SOURCE: U10_imagery.md -->
### Timed lock-expiry project (deferred)
- **category:** locks
- **status-signal:** aspirational
- **status-evidence:** "Approved policy (2026-05-08): implement timed expiry as a focused future project… Sequenced after the imagery loop work completes."
- **what:** One migration adding `lock_type` + `lock_expires_at` to all four Pattern A tables (page_components, site_components, site_plan_directives, assets); auto-lock writers default from a policy table ('admin' permanent, 'deploy' timed/30, auditor approvals timed/90); ~8–10 callsite filter expansions; CheckComponentLock extended; new `expired_review_locks` discovery check. Restores the rhythm doc 004 v4 designed, of which only the audit-pass-counter-reset half shipped.
- **sources:** old/README.md, STATUS_imagery_2026-05-08.md#Lock-expiry-investigation, PLAN_imagery_loop_closure.md#Decisions
- **relations:** references LOCKS_should_locks_expire.md (outside this unit); asset locking 2A.
- **verify-later:** whether lock_type/lock_expires_at columns exist on any Pattern A table.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Locks — HITL durability across the platform
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 031_locks(2) "This doc is the canonical reference for lock semantics"; "Tech debt: lock-model coherence (target model) … Status (2026-05-19): the lock model has accreted"
- **what:** Two per-row lock patterns protect human-edited data: Pattern A (`locked_at`+`locked_by`, dominant) and legacy Pattern B (`pinned` boolean on site_specs, don't use for new tables). Every writer must read lock state before writing and preserve it when superseding. A coherence cleanup to three orthogonal columns under the invariant permanent⟺human is recorded as deferred tech debt.
- **sources:** WM/031_locks(2).md#the-two-patterns-in-use, WM/031_locks(2).md#lock-transfer-across-rebuilds, WM/031_locks(2).md#tech-debt-lock-model-coherence-target-model, WM/030_phase1_plan_and_reconciler(4).md#lock-transfer-across-plan-rebuilds
- **relations:** human direction/lock lifecycle (007); adoption faithfulness via locks; site plan directives
- **verify-later:** migration 053; check_component_lock.go; FOCUS_adoption_faithfulness_via_locks.md; PLAN_lock_coherence.md

<!-- SOURCE: U18_sql_for_agents.md -->
### Section/component locking with timed expiry
- **category:** locks
- **status-signal:** partial
- **status-evidence:** 115 header: "the project doc 004 v4 designed and docs 031_locks... approved (2026-05-08). Implemented now (Option A)... This migration is SCHEMA + BACKFILL only. The Go follow-on... lands as separate code changes."
- **what:** Locking is the improvement loop's termination and protection mechanism: verified/human-edited rows get locked_at set; auditors exclude locked rows (086). 115 adds lock_type + lock_expires_at to all four Pattern A lock-bearing tables (page_components, site_components, site_plan_directives, +1) in one transaction for coherence. Policy: admin/manual/checkpoint = permanent; deploy = timed +30d; visual-design-auditor / imagery-quality-auditor / adoption (new, faithful-first-pass) = timed +90d. Unlock predicate: `locked_at IS NULL OR (lock_type='timed' AND lock_expires_at < NOW())`. Go-side sweep of 11 callsites still pending at write time.
- **sources:** 115_locks.sql; 086_visual_design_auditor.sql
- **relations:** adoption faithfulness (FOCUS_adoption_faithfulness_via_locks.md); expired_review_locks discovery check (planned)
- **verify-later:** the 11 `locked_at IS NULL` callsites; CheckComponentLock extension; whether expiry sweep landed

<!-- SOURCE: U19_sql_tables_components.md -->
### Pattern A lock convention (locked_at / locked_by, hard vs soft)
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** 041 Phase 2A codifies: "a row is locked if locked_at IS NOT NULL. No time comparison... timed expiry is documented design intent (004 v4, 007 v4) but not implemented"; canonical classifier named (check_component_lock.go CheckComponentLock → IsHard).
- **what:** Uniform HITL/agent lock across four tables (page_components, site_components, assets, site_plan_directives — plus site_plan_imagery): locked_at timestamp + locked_by identity. Hard locks ('admin', 'admin-removed', 'checkpoint', 'manual' upload) only humans clear; soft locks ('deploy', auditor names, 'audit-pending') agents may clear when a work item references the row. Discovery skips both; execution skips hard. locked_by vocabulary is convention, not CHECK, to allow new identifiers without migration. A future lock-expiry project would add lock_type/lock_expires_at across all Pattern A tables in one migration.
- **sources:** docs/agent_docs/sql_for_tables/041_assets.sql#Phase2A; docs/agent_docs/sql_for_tables/012_site_components.sql#phase-7a; docs/agent_docs/sql_for_tables/040_site_plans_schema.sql#directives
- **relations:** 031_locks.md canonical doc; site-level lock; imagery/directive lock transfer.
- **verify-later:** CheckComponentLock consumers; lock-expiry project status.

<!-- SOURCE: U19_sql_tables_components.md -->
### Site-level lock (sites.locked_at)
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** "Phase 7: Site-level lock — prevents all automated agent activity" (012 tail); scheduled-task pre_query patched to exclude locked sites (020 site-lock section).
- **what:** locked_at/locked_by on sites acts as a master switch: when set, no automated agent activity (discovery, dispatch, improvement) touches the site. Scheduler pre_queries filter locked sites out of candidate selection.
- **sources:** docs/agent_docs/sql_for_tables/012_site_components.sql#phase-7; docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#site-lock
- **relations:** Pattern A locks; scheduler pre_query gating.
- **verify-later:** all dispatch/discovery entry points honour sites.locked_at.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Auto-lock on deploy (page_components lock trigger)
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** Schema captured 2026-07-01: "trigger_auto_lock_on_deploy auto-locks on deploy... lock_type permanent|timed|review"; lock check run pre-rebuild (all 4 index rows unlocked).
- **what:** page_components carries locked_at/lock_type/locked_by with a trigger that auto-locks components on deploy (fires on UPDATE). Operational consequence observed: deployed components MAY be locked, so rebuilds/re-renders must check lock state (a lock could block re-render of a target or protect neighbours); on the vonc index all rows were NULL-locked so the behaviour never actually bit in this corpus.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-01-~13:25 + #2026-07-01-~13:55; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§7
- **relations:** locks category (031); save_page_sections (does it honour locks? open question in 016b Part 4)
- **verify-later:** trigger_auto_lock_on_deploy definition; save_page_sections lock handling

<!-- SOURCE: U25_leopardess_social.md -->
### auto_lock_on_deploy trigger and the stillborn strict-mode subsystem
- **category:** locks
- **status-signal:** abandoned
- **status-evidence:** RUNNING_NOTES_minilobby (2026-07-09 record): "the strict-mode subsystem it belonged to was stillborn — no Go code reads schema_mode … snapshot columns never created … fired exactly once in the system's history"; dropped via migration 009 with saved reversal.
- **what:** A BEFORE UPDATE trigger stamping schema_mode='strict' + lock fields when a row reached deployed on first_deploy sites. Never functional as designed: save_page_sections INSERTs rows already deployed (trigger never fires), its companion snapshot columns were never created, and nothing reads the lock. It nearly sabotaged the section-editor fix (every edit would have locked its row) and was dropped 2026-07-10 with the function body backed up. schema_mode/strict_mode_trigger columns and the orphaned lock/unlock functions deliberately retained.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-09-the-dropped-trigger; docs/social001_vonc_tiktok_social/minilobby_task/auto_lock_on_deploy.FUNCTION_BACKUP.sql; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#8.2
- **relations:** locks category (031 lock semantics); build_status defect (the near-collision)
- **verify-later:** trigger absence on page_components; 009_drop_auto_lock_on_deploy.sql; leftover lock functions
