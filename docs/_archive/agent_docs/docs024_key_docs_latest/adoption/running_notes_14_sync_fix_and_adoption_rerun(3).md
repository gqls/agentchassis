# Running Notes — Section-Index Sync Fix, Adoption Re-run, and Post-Run Catalogue

**Session date:** 2026-05-27 → 2026-05-28
**Scope:** Close out the section-index hub reversion bug (diagnosis → fix decision → patch → deploy → verify), recover from a Kafka/consume outage that blocked re-adoption, verify the fix end-to-end, then catalogue what remains broken on the adopted site.

This is a thinking/decisions log — including the dead-ends and the corrections — not a clean writeup. The durable reference version of the technical findings lives in the debugging guide (`016_debugging_guide`, v2_26); this captures *how we got there* and *why we chose what we chose*.

---

## Part 1 — Diagnosing the section-index hub reversion

**Symptom.** After the prior session's fixes (analyze_site section-index prompt; populate_nav section-index nav), a re-adopt deployed `gamesdesign.co.uk` with the header/footer linking to flat `/tools-index.html`, `/games-index.html` — not the nested `/tools/index.html`. The hubs had reverted to flat.

**The narrowing (the useful part):**

1. **Split the declarative artefact from the realised table.** `site_plan_pages` (current plan) showed the hubs correctly as `section-index` + nested URLs. `pages` showed them flat (`content`, `/games-index.html`). Same names, disagreeing. Conclusion: the divergence is in the plan→pages write surface, not the planner. (This is the "authoritative output is `site_plan_pages`" discipline used in reverse: rendered output wrong, plan right ⇒ look at the writer between them.)

2. **The source fix gave a controlled before/after.** Because the analyze_site prompt fix was already in, the adoption-time checkpoint had shown `pages` correct (section-index) earlier in the same run. So `pages` *started* correct and went flat later ⇒ regression is a stage that runs *after* adoption.

3. **In-place overwrite, not duplication.** Two hub rows, not four ⇒ the later stage overwrote `url`/`page_type` in place.

4. **Dead-end corrected — rerender is not the writer.** The flat rows' `updated_at` fell inside the `page_rerender` wave, which made rerender the tempting culprit. Reading `rerender_single_page_action.go` killed that: it only *reads* `pages`, writes none of name/url/page_type (its sole page write is `update_page_status`), and derives the deployed filename from the *existing* `pages.url`. So the flat links are a faithful render of a value rerender never wrote; the matching timestamp was the status update touching the row, not authorship. → Logged as checklist item 20: a matching `updated_at` is not proof of authorship; confirm the action writes the column before blaming it.

5. **Dead-end corrected — reconciler is not the writer.** `reconcile_site_plan_action.go` only reads `site_plan_pages`/`pages`, diffs, and emits `needs_page` work items; its only writes are to `site_work_items` and `sites.last_reconciled_at`. It keys work items on the plan's names (so they're faithful). Exonerated — read/decide/emit only.

6. **`site_plans` count killed the "stale plan" branch.** Exactly one plan row, `is_current`, `write_site_plan`, nested/correct. So sync didn't read a different/stale plan — it *re-derived*. That collapsed "copy vs re-derive" to **re-derive** before even reading the writer.

7. **Cause pinned — `SyncPagesToDBAction` (`site_db_actions.go`).** Two findings together:
   - It reads pages from `page_plan` in collected data (`extractPagesFromPlan`) — the raw planner output — **not** from `site_plan_pages`. So Part A's correction never reaches it.
   - It calls `CanonicalisePage` directly with **no `ValidateRoles`** first. Part A (the `-index` name → `section-index` rule) lives in `ValidateRoles`, which runs only in `WriteSitePlanAction`. So a `games-index` typed `content` in `page_plan` stays `content` → `CanonicalisePage` produces flat `/games-index.html`. Exactly the observed row.
   - `upsertPage` does `ON CONFLICT (site_id,name) DO UPDATE SET url=…, page_type=…` → overwrites the correct adoption row with the flat values; and its INSERT never sets `built_from_plan_version` (→ NULL → reconciler treats as `stale` → re-emits forever: a second, independent churn symptom).

**Structural framing:** Tension #2 pinned to code — two canonicalisation surfaces that disagree. `WriteSitePlanAction` runs `ValidateRoles + CanonicalisePage` → `site_plan_pages` section-index. `SyncPagesToDBAction` runs `CanonicalisePage` only on the raw `page_plan` → `pages` flat. Part A fixed the first surface; the second was never touched.

---

## Part 2 — Choosing the fix

Two candidate directions:

- **Option 1 (single source of truth):** sync reads identity from `site_plan_pages` and carries `url`/`role`/`page_type` through verbatim.
- **Option 2 (make the two surfaces agree):** sync runs `ValidateRoles` before `CanonicalisePage`, reusing `datahelpers.ValidateRoles`, so both surfaces apply the identical pipeline.

**Investigation that decided it.** `sync_pages_to_db` has five callers. `build-site-planner` (the adoption planner) has the step order `… plan_site → validate_plan → write_site_plan → sync_pages → populate_nav → reconcile_site_plan` — so sync runs *immediately after* `write_site_plan`; a current plan exists there, so Option 1 would work *in that workflow*. **But** `pageflow-builder` (confirmed `active`), `multipage-website-builder`, and `site-work-orchestrator` call sync and never write a plan / never reference `site_plan_pages`. So at sync time in those workflows there is **no plan to read** — Option 1 as a blanket change would break the active `pageflow-builder`, and guarding it with a `page_plan` fallback re-introduces the very divergence we're removing.

**Decision: Option 2.** Uniform across all five callers, reuses existing code, and collapses the divergence at its root (one canonicalisation pipeline in both writers). Corrected an earlier framing that called Option 1 "the structural one" — Option 2 is the structural fix here; Option 1 is coupling. Recorded that `pageflow-builder` deprecation is now *independent* (not a prerequisite).

**Guides side-effect discovered before shipping.** Running `ValidateRoles` in sync also affects guides: `ValidateRoles` strips the `guide-` prefix to derive the slug, and `blog-post` uses the bare slug → guides de-prefix from `guide-rng-design` to `rng-design`. The OLD sync (no `ValidateRoles`) had preserved `guide-rng-design`. So the patch pulls `pages` toward the plan's de-prefixed form for guides. This is consistent (plan already de-prefixes; the patch makes pages agree) but it is the less-faithful direction — flagged as a deliberate decision for the user rather than a silent change (debug-doc item 16: don't change a value earlier evidence proved correct without surfacing it).

**Falsified the "type guides as `guide`" companion.** Hypothesised the faithful fix was to type guides as `guide` so `CanonicalisePage` rebuilds the prefix. A read of `pages` falsified it: the source guides live flat at `/blog/guide-rng-design.html`, while a `guide` role would *nest* them at `/guides/<slug>/index.html` — wrong directory and different slug. So typing `guide` would be *less* faithful. The faithful fix (if wanted) is to stop the de-prefix for `blog-post` guide pages, preserving `/blog/guide-rng-design.html` — and that needs `page_canonical.go` to do correctly. Left as an open product decision; did NOT ship the wrong patch. (Good example of read-before-prescribe catching a plausible-but-wrong fix.)

---

## Part 3 — The patch (`site_db_actions.go`)

Three edits, all in `SyncPagesToDBAction` / `upsertPage`, mirroring `write_site_plan_action.go:235–278` verbatim (reuse, not reconstruction):

1. Normalisation loop builds `[]datahelpers.LLMPlannedPage{Name, Role: firstNonEmptyField(page,"page_type","type","role"), Slug, URL, ParentSection}`, runs `datahelpers.ValidateRoles` over the whole set (order preserved, so `validated[i]` aligns with `pages[i]`), then per-page `CanonicalisePage{Role: v.Role, Slug: firstNonEmpty(v.Slug, v.Name), ParentSection: v.ParentSection}`. Logs `CorrectedFromRole`.
2. Resolves the current plan id once (`SELECT id FROM site_plans WHERE site_id=$1 AND is_current=true`) and threads it to `upsertPage`. Best-effort: greenfield callers with no plan → invalid `NullUUID` → column left NULL → unchanged behaviour (this is what keeps it safe for `pageflow-builder`).
3. `upsertPage` gains `planID uuid.NullUUID`, writes a `built_from_plan_version` `$12` column, and on conflict does `built_from_plan_version = COALESCE(EXCLUDED…, pages…)`.

Caveats recorded with the patch (sandbox has no Go toolchain): run gofmt/vet/build; the plan-id lookup uses `params.DB.QueryRowContext` (matches `reconcile_site_plan_action.go`, so `ActionParams.DB` is `*sql.DB` in practice, but `upsertPage`'s type-switch also tolerates `*pgxpool.Pool`); left the existing `deployed → needs_rebuild` ON CONFLICT flip untouched (separate behaviour, not part of this regression).

---

## Part 4 — Kafka / consume outage (the thing that ate the afternoon)

Re-adoption wouldn't start. Triggers printed IDs but produced no orchestration rows. The narrowing:

- A `sites` row was created but `orchestration_states` stayed empty across multiple triggers — "the run isn't running," upstream of anything we changed.
- **Initial mis-framing corrected:** assumed `ensure_site_record` ran inside the workflow ⇒ workflow started. But an orchestration writes a state row at creation; none existed. So the site row was being created by the trigger path, and the orchestration was never created ⇒ message not consumed.
- The topic `system.agent.generic.requests` *existed* (broker list confirmed) and had messages (end offset climbing to 50+). So producer was fine — not a missing-topic problem.
- The decisive read: `kafka-consumer-groups --describe --group generic-requests-group` returned **empty** — no committed offsets, no live members. The chassis logged a clean consumer setup but wasn't actually joined to the group. Root: the Kafka topic wipe destroyed the group's `__consumer_offsets`/membership and the consumer hadn't re-established it.
- **Mistake during recovery:** ran a `--reset-offsets --to-earliest` (the path I'd cautioned against), which set the group to replay all 50+ backlog messages — including two stale adopt triggers. Risk: replaying would spawn duplicate/stale adoptions. Mitigation chosen: don't replay history; restart the chassis to rejoin, park the group at latest, then fire one fresh trigger.
- **Red herring noticed and set aside:** the running orchestration's embedded `agent_config` for `site-adoption-orchestrator` showed a one-step no-op `complete` workflow ("scheduled task pre_query already did the work"), which looked like a stub that would no-op the adoption. Held off concluding — and it was indeed harmless: the real three-step `spawn_adopter → call_adopter → complete` ran.
- **Also corrected:** a `CHILD_ORCHESTRATION_FAILED` / `sql: no rows in result set` error in the logs was a race between the DB cleanup and an in-flight *scheduler* tick (`build-pipeline-trigger` → dispatch loop), not the adoption — noise from deleting rows mid-flight, not a fault.

**Resolution:** chassis restart re-established group membership; a fresh trigger produced orchestration `c29d5dc3…` (orchestrator) which spawned child `fd3b22fa…` (`owner_agent_type=site-adoption-agent`). Both completed. Adoption ran on the fixed image.

---

## Part 5 — Verification (fix landed)

After full cascade drain (work items: 21 `needs_page`, 2 `needs_rerender`, 26 `page_rerender` all complete; `active=0`):

- Hubs: `games-index`/`tools-index`/`guides-index` all `section-index`, nested URLs, `build_status=deployed`, `built_from_plan_version = 89b70231…` (the current plan).
- `site_plan_pages` and `pages` agree on the hubs.
- Rendered header/footer link to the nested hub URLs across all index pages.

Both the core fix (sync no longer flattens hubs) and the companion (`built_from_plan_version` set, reconciler stops churning) are confirmed. **Thread closed.**

---

## Part 6 — Post-run reality: the rest of the site

Getting far enough to deploy revealed the adoption is far from complete. Full enumeration in `CATALOGUE_gamesdesign_post_sync_fix_defects.md`. Headlines:

- **Tools and games PAGES did not deploy at all** (confirmed). `pages` has the rows, `needs_page`/`needs_tool_recreation` complete, but `/tools/` and `/games/` in the repo contain only the hub `index.html`. The interactive product is absent. Top priority.
- Pervasive **silent-fallback links** to non-existent `/contact.html`, `/services.html`, `/terms.html` across hero/cta/header/footer, plus empty `href=""` on "Browse All Tools" and "Play Now".
- **List components:** empty card descriptions; game-list fabricates games (and each instance fabricates a *different* set) instead of querying realised game pages, while tool-list *does* query real pages; blank card images; non-functional filters and load-more.
- **Section-data gaps:** guides hub has no guide-list; empty feature `<h3>` headings; two `needs_section_data` at `needs_human_review`.
- **Content quality:** hero H1 reused across all three hubs; titles carry source "- GameDesign.uk"; empty meta descriptions; empty footer contact/tagline.
- **Guides duplication** (`guide-rng-design` + `rng-design`) — the parked faithfulness decision.
- **Design fidelity** vs source — likely the separate design chat's domain.

Agreed approach: catalogue first (done), then work each as its own thread with the same discipline — read the responsible action/source, pin the cause, confirm against data, propose fix. No fixing from tentative causes.

---

## Decisions log (concise)

- **Fix = Option 2** (sync runs `ValidateRoles` before `CanonicalisePage`), not Option 1 (read `site_plan_pages`), because `pageflow-builder` (active) and other callers have no plan at sync time. Reuse `datahelpers.ValidateRoles`.
- **Companion:** set `built_from_plan_version` in `upsertPage` when a plan id is available; greenfield → NULL, unchanged.
- **Did NOT** retype guides as `guide` — the read showed source guides are flat `/blog/`, so that would be less faithful. Guides-faithfulness left as an open decision needing `page_canonical.go`.
- **Did NOT** touch the `deployed → needs_rebuild` ON CONFLICT flip (separate behaviour).
- **`pageflow-builder` deprecation** decoupled from this fix.
- **Kafka recovery:** restart-to-rejoin + park at latest + one fresh trigger, rather than replay-from-earliest (which would spawn stale adoptions).

## Principles reinforced this session

- A matching `updated_at` is not proof of authorship — read the action; confirm it writes the column (rerender was exonerated this way). Checklist item 20.
- Fixing the upstream value first turns "which stage is wrong" into "which stage changed a known-good value" (the prompt fix made `pages` provably correct at adoption time, isolating the downstream writer).
- Split the declarative artefact (`site_plan_pages`) from the realised table (`pages`) to localise which surface diverged.
- Read-before-prescribe caught a plausible-but-wrong fix twice: the rerender mis-attribution, and the "type guides as `guide`" companion that the data falsified.
- "The run completed" ≠ "the cascade drained" ≠ "the deploy is correct" — verify at each stage, not on the first green signal.
- A trigger script printing IDs only means a message was sent, not consumed — confirm consumption (offsets, group membership), not just that the producer ran.

---

## Part 7 — A1 investigation: tool/game pages never deploy a file (IN PROGRESS)

Picked A1 (no tool/game pages in the repo) as the first catalogue thread. Trace so far:

**Site-id correction (cost a near-miss).** A batch of A1 queries was pinned to the hardcoded `site_id = 166bb28d-…` carried over from earlier notes. Every query returned zero rows, and a `pages LEFT JOIN site_work_items ON w.page_id = p.id` came back empty — which was nearly misread as "no work items were emitted for these pages" and "the work items never set `page_id`." Both wrong. The live site for the clean 12:43 run is `5edc4130-…` (its `sites.created_at` matched the trigger); `166bb28d` was a *prior* teardown's row, gone. Re-running against `5edc4130` returned 86 work items and a populated `page_id` join. Lessons captured as checklist item 21: a teardown deletes the `sites` row and re-adopt mints a new UUID, so never carry a `site_id` literal across a teardown — resolve via `(SELECT id FROM sites WHERE domain=…)`; and a LEFT JOIN empty on the left means wrong/absent anchor id, not a missing right-side link (a real missing link returns left rows with NULLs, not an empty set). The fix-verification queries earlier in the session used the `WHERE domain=` subquery, so they were correct throughout — only the hardcoded-id A1 queries broke.

**What the work items show.** Every tool/game work item is `complete` — not stuck, failed, or DLQ'd. So A1 is not a queue problem. The `agent_error_log` for `5edc4130` is empty; the historical errors (ephemeral `job.*.responses` topic-not-found, unrendered `{{end}}` blockers, LLM usage-limit) are all from prior site_ids — ruled out.

**The status split is a red herring.** `needs_rebuild` = 6 tools + `game-auto-battler`; `deployed` = the other 4 games. This tracks the `site_db_actions.go` upsert `ON CONFLICT` branch `WHEN pages.build_status='deployed' THEN 'needs_rebuild'`: `sync_pages` runs ~16:08 and flips every then-`deployed` page; only tool-recreations completed by 16:08 (the 6 tools + auto-battler) got flipped, the 4 later games didn't. Real status-churn bug — but independent of the missing files: `game-jelly-invaders` is `deployed` and *also* has zero components and no file.

**The actual cause (confirmed): empty `page_components` → rerender skip.** `assemblePage` → `getPageSections(page_id)` reads `page_components` (`rendered_html`, ordered by `position`), **not** `pages.sections`. When it returns nothing, `assemblePage` returns `""`, `skipped=true`, the workflow routes to `complete_skipped`, and **neither `git_commit` nor `update_page_status` runs**. The decisive query — `count(pc) FILTER (WHERE rendered_html<>'')` per page — returns **0** for every `tool`/`game` page, 1 for a `blog-post`, 2 for a `section-index` hub. So no tool/game page has rendered components, the rerender skips all of them, and no file is written. `build_status` is irrelevant to file presence.

**Where it should have worked.** The `tool-recreation-handler` workflow: `recreate_tool → check_completeness → validate_tool → save_sections → update_status(deployed) → spawn_rerender → deploy_page`. `save_sections` = action `save_page_sections`, "Persist generated tool HTML to page_components", reading `validation_result.clean_html`. `deployed_at` is set on every tool/game page, so `update_status` ran, so `save_sections` ran before it — yet `page_components` is empty. So `save_page_sections` is not landing the recreated tool HTML as a readable `page_components` row.

**Next step (not yet pinned — read before prescribing).** Read `save_page_sections`. Two leading hypotheses, to be settled by the source: (a) it splits HTML into sections on an element boundary (e.g. `<section>`), and the tool output is a single `<div class="tool-page">…` → zero sections persisted; or (b) it writes `pages.sections` jsonb while the rerender reads `page_components.rendered_html` — a save/read surface mismatch (Tension #2 family). The recreated tool artefact exists upstream (the `recreate_tool` Opus step produced it; training-data was saved) but never becomes a deployable component. No fix until `save_page_sections` says which.

**Secondary, noted not fixed:** the `deployed → needs_rebuild` upsert flip churns status for whichever pages are `deployed` when `sync_pages` runs. Separate from A1's file cause; revisit after A1.

## Principles reinforced (Part 7)

- A hardcoded id from a prior run is stale after any teardown — re-resolve from `sites` by domain; an empty LEFT JOIN means wrong anchor id, not a broken relationship. (Checklist item 21.)
- `getPageSections` reads `page_components`, not `pages.sections` — "has sections" and "has rendered components" are different facts; the deploy path depends on the latter.
- The deployed/needs_rebuild status and the presence of a committed file are independent — don't read `build_status=deployed` as "the file shipped."

---

## Part 8 — A1 fix: parser fallback + Option B (deployed→needs_rebuild investigation)

Two coordinated fixes shipped for A1 (tool/game pages never deploy a file).

**Parser fix (`save_page_sections_action.go`).** Pinned cause: `saveSectionsExtractFromHTML` extracts only `<section>…</section>` blocks, but `tool-recreation-handler`'s `recreate_tool` prompt emits the tool as `<div class="tool-page">…</div>` (no `<section>`), and `save_sections` sets no `sections_metadata_field` so it relies on that fallback. Zero matches → zero `page_components` → the page-rerender's `getPageSections` returns empty → skip (no git commit). Confirmed by contrast: blog-post `n_rendered=1`, hub `n_rendered=2`, every tool/game `n_rendered=0`. Fix: when zero `<section>` blocks match but HTML is non-empty, store the whole fragment as one section (guarded against full documents via `<html`/`<!doctype` so assembled pages aren't double-chromed). Reuses the existing insert path; stops silent content loss.

**Option B investigation — the `deployed → needs_rebuild` flip.** The user asked to determine why the flip exists, why in `upsertPage`, and what was broken before it, then decide if B was still right. Findings, grounded in the repo:
- Intended design (`029`/`030`): stamp `pages.built_from_plan_version` at *build time* (029:299–300, 319) and detect staleness in the *reconciler* (029:279, 392; 030 item 9 "Drift detection … Reconciler enhancement").
- What shipped (`HANDOFF_2026-05-07` #5, verbatim): the stamp was **never written by page-build-handler** — explicitly deferred, with "reconciler treats NULL as stale and emits needs_page for every deployed page" noted as the known consequence and "User explicitly OK'd this." Backfill listed again in the 05-08 handoff.
- So the flip is a pre-design *stand-in* for "re-sync invalidates deployed pages," living in `upsertPage` because that's the single chokepoint where the plan's page set is written (so "already exists + deployed" is detectable in one pass). It over-fires (every deployed page, every sync) and mis-fires on pre-plan deploys (tools) — the A1 churn. Same symptom on another site in `HANDOFF_robot_hands_rebuild`.
- My own companion fix earlier this session (stamp the version in *sync* via `COALESCE(EXCLUDED,…)`) was a misplaced partial of the deferred stamp — wrong location, wrong COALESCE direction — which papered over the hubs (deployed after sync) but not the tools, and would corrupt drift across re-plans.

Conclusion: **B is still right, and the investigation strengthened it** — B completes the originally-designed-but-deferred architecture rather than inventing or patching around. Rejected the narrow Option A (exclude tool/game from the flip) because it would entrench the workaround and leave the real debt.

**Option B shipped (two files, coupled):**
- `v3_site_actions.go` `UpdatePageStatusAction` deployed branch: stamp `built_from_plan_version = COALESCE((SELECT sp.id FROM site_plans sp WHERE sp.site_id=pages.site_id AND sp.is_current=true), built_from_plan_version)`. The deferred build-time stamp, at the single deploy-status chokepoint (covers tool-recreation/page-rerender/page-build). Keeps existing value when no plan exists yet (pre-plan tools).
- `site_db_actions.go` `upsertPage`: COALESCE → fill-if-null (never overwrite a real build version; preserves re-plan drift; adopts pre-plan deploys); removed the `deployed→needs_rebuild` CASE branch (kept `NULL→planned`, `ELSE` passthrough). Drift now flows through the reconciler's `decideEmit`.

Traces: single adoption — pre-plan tool (version NULL) → sync fills current, no flip → reconciler skips → tool stays deployed. Re-plan — page keeps v1, no flip → reconciler sees v1≠v2 → rebuild → deploy re-stamps v2 → settles. Coupled: flip removal without deploy-time stamp would re-churn on re-plan.

**Is A1 fully fixed? Honest assessment.** The two fixes address the two *confirmed* causes (div-based HTML the regex can't parse; the flip churn). One link is *unverified*: that the tool HTML reaches `save_page_sections` intact via `validation_result.clean_html` (chain: `recreate_tool → check_completeness → validate_tool → validation_result.clean_html`). The observed `n_rendered=0` fits both "HTML arrived without `<section>`" (fixed) and "HTML never arrived" (not fixed by the parser change). The next adoption settles it: tools showing `n_rendered≥1` + a file at `/tools/<slug>/index.html` ⇒ A1 closed; still `0` ⇒ read `validate_page_content`'s output contract next. Also: the fixes apply to *new* builds — the current gamesdesign site's existing tools need a re-adopt to benefit (not retroactively deployed). The secondary `deployed→needs_rebuild` churn that was previously "next item after A1" is now resolved by Option B (folded in, not separate).

## Principles reinforced (Part 8)
- Before fixing a misbehaving mechanism, check design docs + handoffs for deferred debt — the bug may be a half-implemented design; complete it rather than patch around it (checklist item 22). The flip was a stand-in for the deferred drift stamp.
- A status value set correctly by one action can be invalidated by another action's `ON CONFLICT` — trace the upsert, not just the setter.
- State necessary-vs-sufficient honestly: a confirmed-correct fix (parser fallback) can still be insufficient if an unverified upstream link (clean_html flow) hasn't been proven; name what the next observation must show.

---

## Part 9 — A1 verification on the fixed image + dispatch-throughput finding (2026-06-03, correlation e9609749)

Re-adoption on the image carrying all three fixes (parser fallback, deploy-time stamp, flip removal). New site_id (resolved via `WHERE domain='gamesdesign.co.uk'`).

**A1 parser fix CONFIRMED WORKING.** Post-(partial) cascade, the four tools whose recreation completed — `tool-ttk-calculator`, `tool-ehp-calculator`, `tool-drop-rate-simulator`, `tool-progression-architect` — all show `n_rendered=1`, `build_status=deployed`, and `tools/progression-architect/index.html` is committed to the repo. This is the **first time a tool page has produced a `page_component` and a deployed file**. It also settles the last open question from Part 8: the tool HTML *does* reach `save_page_sections` (the "<div> not <section>" case, now handled) — the `clean_html`-flow hypothesis is moot. Hubs unchanged (section-index, nested URLs); no regression. `stamped=f` on the deployed tools is expected: they deployed before the plan existed (tool-recreation runs early), so the deploy-time stamp had no current plan to write; `sync` fills the version on its pass, which hasn't completed because the cascade is still draining. Recheck `stamped` after full drain.

**The other 7 tools/games are not an A1 problem — they're throttled by dispatch throughput.** They sit `triaged`/`planned`, unbuilt. Cause, per `FOCUS_dispatch_diagnostic_4` Q3 (confirmed against this run's data): the dispatcher is one-site-per-tick (LIMIT 1, spawned per scheduler tick via `build-pipeline-trigger`, ~5 items then exits) and NOT-EXISTS-blocked (a site is excluded *entirely* while any of its items is `status='claimed'` — absolute, no fall-through, line 276). So tools dispatch serially within the site, each held ~5 min for its Opus build; and the two "Claim timed out — handler pod likely died" items froze the whole site for the claim-timeout window (the 47–67 min gaps between claims at 14:41 / 15:04 / 16:11 / 16:58). The scheduler is `scheduled_tasks`/kafka-scheduler driven — no `build-pipeline-trigger` k8s cronjob (cronjobs are only `agent-job-cleanup` and `database-backup`) — so between ticks an eligible site can idle (the 48-min gap we saw). At time of writing a fresh `build-dispatch-loop` is spawning and a gamesdesign item is `claimed` (17:53), so it's moving again, one tool at a time.

**Decision:** let the site complete naturally (it's draining, just slowly). Spin the dispatch-loop throughput work into a **separate thread** (catalogued as Family J): whether to relax one-site-per-tick / NOT-EXISTS (bounded per-site concurrency or per-item exclusion), shorten the claim-timeout/reaper window so dead handlers don't freeze a site, and the `build-pipeline-trigger` cadence. Not blocking A1 correctness.

**A1 status:** fix shipped and parser fix verified on the 4 built tools; full closure pending natural drain of all 11 (re-run the decisive read once `active=0` and the `needs_tool_recreation` set is all `complete`). Catalogue A1 entry updated to "FIX SHIPPED — PARTIALLY VERIFIED"; Family J added for the dispatch thread.

## Principles reinforced (Part 9)
- Don't read a slow/partial run as a failed fix. `active=0` with outstanding `triaged` work = paused between dispatch waves, not done. The 4 deployed tools proved the fix; the 7 unbuilt were a throughput artefact, not a correctness one.
- Separate the fix under test from the infrastructure carrying it. A1's code was validated even though the run didn't finish, because the validation only needed *some* tools to traverse the fixed path — not all of them.
