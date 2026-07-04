# HANDOFF — component-regen clobber → recovery → F8 fallback contamination
_Last updated: 2026-07-04. This document is the cold-start entry point. The RUNBOOK (step-by-step, ticked) and
NOTES (§1–9au investigation log, every correction owned) in this folder are the authoritative detail._

## Platform operating model (essentials)
- Agents = rows in Postgres `agent_definitions` (DB `clients_db`), run as pods in `-n ai-persona-system`,
  Kafka messaging, work items in `site_work_items` claimed by build-dispatch-loop (dispatchable status = `triaged`).
- All SQL/kubectl/git is run by the human:
  `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`
- Workflow jsonb changes are immediate; Go changes = build → GitHub Actions → Backblaze → image_tag bump.
- Backups: `snapshot_agent('<type>','<reason>')` for agents; CTAS `*_bak_*` tables for data (two exist, see Cleanup).
- **Another chat co-manages** components / vonc.com / idea.uk — freshness-check + coordinate before shared writes
  (optimistic-lock UPDATEs on `updated_at` are the established pattern; `UPDATE 0` = stop and re-read).
- RECURRING GOTCHA: pasted attachments often extract EMPTY. Fix: capture psql output with `\o /tmp/file.txt … \o`
  and upload the file (readable at /mnt/user-data/uploads via bash), or paste inline in the message body.

## Incident 1 — component-regen clobber: RESOLVED + RECOVERED (verified)
- Root cause: LLM regeneration of shared `system-stats` (fdd92ad4-521a-4602-89cf-7ee1a66c10f1) on 2026-06-24 15:06
  renamed the field contract; dependents' stored content stopped matching; five live pages showed empty bands.
- Fix set DEPLOYED + PROVEN LIVE: **F1** field-contract guard in StoreGeneratedComponentAction (observed rejecting);
  **F1-prompt** (`load_existing_component` loader + regeneration field-name rule + function pin); **F3** scoped
  rerender (create_rerender_items scopes to component dependents + stamps `spec.reason`; store carries
  `section_data_resolved`; gate matched end-to-end).
- Recovery COMPLETE: content re-keyed (20 mappings), per-site `site_specs` `cta` aspects inserted
  (`/contact.html` ×3, verified real), all five pages re-rendered with distinct md5s + content needles true
  (R6b pass 2026-07-03), leopardess/index confirmed live by screenshot. vonc.com/index arrived mid-recovery as a
  healthy sixth dependent (new builds consume the guarded component correctly).
- R6 live-eyeball ledger: leopardess/index ✔ · robot-hands/index stats band ✔ · ai-agent/index mostly-ok per user
  (case-study-card images render as alt-text — image-pipeline territory, parked) ·
  **ai-agent/case-study-kafka-consumer-group-remediation never eyeballed** · gripper-detail → became R6c/R6f below.

## Incident 2 — R6e / F8: brief-explanation fallback contamination (harm-stop DONE, tail PENDING)
- Mechanism (proven): shared `brief-explanation` was regenerated 2026-07-01 12:46 (PRE-guard, vonc-work window)
  with vonc's gauntlet copy baked into **static-field FALLBACK VALUES** (cta "Play Today's Gauntlet" /
  "See Past Results"; stat_1..3 "New Gauntlet Daily"/"24hrs", "Players Scored"/"10K+", "Free to Play"/"100%").
  **F8: the F1 guard checks field NAMES, not fallback CONTENT.** Fallbacks were merged into dependents'
  content_data by the stored⊕resolved merge and consumed by new builds (idea.uk, built 07-03 16:27+).
  The component was also updated 07-03 13:22 with NO version row (provenance open).
- EXECUTED 2026-07-03 (other chat informed): manual snapshot v2 → neutralize UPDATE 1 with optimistic lock
  (stats → source=llm/required:false/no fallback; CTA labels → static "Get Started"/"Learn More") →
  backup CTAS `page_components_bak_briefexp_20260703` (3 rows) → strip 8 merged keys (UPDATE 3: idea.uk ×2 +
  robot-hands index) → 2 scoped rerender triggers (`f8_rerender_briefexp:<component_id>`, INSERT 0 2).
- RESULT hours later: robot-hands index + how-it-works CLEAN (identical 9181B "empty shell" band — survives the
  visible-content filter via the neutral CTA text); **gripper-selection-guide re-rendered 18:04 WITH gauntlet**
  (writer-rebuild suspected via auto-escalation on its empty content_data); **idea.uk ×2 STILL gauntletised**
  post-strip → suspicion: gauntlet migrated into their LLM copy at build time (F8 knock-on — strip can't reach it;
  writer content re-pass needed, coordinate); matchmatrix skipped by the fan-out (status-filter candidate).
- **IMMEDIATE NEXT ACTION — the A–D reads are still outstanding** (three attachment attempts arrived empty/wrong
  buffer). A single `\o /tmp/abcd.txt` block is staged in the last chat turn containing:
  A = chain since 17:00 incl. `w.created_by` (the planner's reconcile_site_plan ALSO emits needs_page/needs_rerender
  — attribution matters); B = `substring(lower(rendered_html) FROM '.{0,70}gauntlet.{0,70}')` on flagged renders;
  C = LATERAL jsonb_each_text(content_data) WHERE value LIKE '%gauntlet%' (**THE DECIDER**); D = robot-hands
  pages.status. Decision tree: C shows gauntlet in idea.uk heading/step_* → writer content re-pass for those fields
  (coordinate with the other chat); B decides gripper-selection-guide (echo vs ordinary selection-guide English);
  then clone the auto-escalated needs_page item's shape for how-it-works (NEVER guess a needs_page spec).

## R6f — theming vocabulary drift (structural fix pending, owner located)
- Newer sites (robot-hands, vonc) have rootless 8009B heads by design; theme vars live in `/assets/css/styles.css`
  (rendered + committed by the **webdesign-agent**; `storage_actions.go` writes it; `fix_harcoded_colours` is the
  post-pass precedent). leopardess = old inline-`:root` head pattern (why it's immune).
- The drift: styles.css `:root` does NOT define the newer template vocabulary — `--section-text`,
  `--section-text-muted`, `--section-surface`, `--section-border`, `--spacing-section`, `--border-radius`,
  `--color-heading`, `--color-white`, and defines `--container-max` while templates consume `--container-max-width`.
  Sections on the new vocabulary render fallback-dark on a dark canvas → invisible (gripper-detail's "blank" page);
  bands with lighter fallbacks stay visible (per-component lottery).
- Structural fix: align the generated `:root` with the consumed vocabulary (feed it into the webdesign generation,
  or a palette-mapped post-pass). Not designed yet.

## Flags / follow-ups (all with evidence in NOTES)
- **F8 mitigations (pending):** prompt rule — fallbacks in shared components must be site-neutral; optional
  store-time lint on fallback strings. Plus the knock-on lesson: contaminated fallbacks poison generated LLM copy.
- **F7 (re-scoped):** update_component_html's snapshot INSERT is already fixed in current code; residual = no
  placeholder⇄schema sync validation on template swaps; hero's 16:43 zero-version update unexplained (likely
  pre-fix image or a benign touch; two hero rows exist but old one is inactive → lookup deterministic).
- **F6:** store guard's NOT EXISTS status list omits `unresolved` (the dedup index treats it as terminal) +
  `itemsCreated++` not gated on RowsAffected in create_rerender_items.
- **F5:** guard extension — a regen ADDING a required, fallback-less field strands renderability (this incident's
  second facet: `cta_url` required from `site_specs.cta.primary_url` with no spec anywhere).
- **F4 (softened):** fork advisory when a regen creates instead of matching; the real evidence remains the F2-3b
  fork; the duplicate `hero` rows were Jan-2026 manual seeding (old row inactive).
- Hygiene: 40 stale `unresolved` page_rerender items (2026-05-01); loose dispatch item-status semantics
  (five sightings: complete-at-dispatch, error-in-complete, status-change-without-timestamp-bump, etc.).
- Cleanup when comfortable: `DROP TABLE page_components_bak_sysstats_20260702;`
  `DROP TABLE page_components_bak_briefexp_20260703;`

## Key IDs / artifacts
- system-stats fdd92ad4-521a-4602-89cf-7ee1a66c10f1 · brief-explanation id resolvable via
  `function='brief-explanation' AND forked_from IS NULL` · sites: gamesdesign e33263f4 (test),
  leopardess 4851f6fc, ai-agent-orch 2a8ebf9c.
- `agent_error_log` timestamp col = `occurred_at`; dedup index `idx_swi_dedup` = UNIQUE(site_id,item_key)
  WHERE status NOT IN (complete,verified,rejected,wont_fix,failed,unresolved).
- This folder: RUNBOOK_/NOTES_ (authoritative), F1/F3 patches + staged .go files, F1prompt*.sql migrations,
  load_existing_component_action.go, store_generated_component_guard_test.go, BUNDLE_/PLAN_ files.
- Full conversation transcripts: /mnt/transcripts (see journal.txt there for the catalog).
