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

---

## Part 10 — A1 closed end-to-end; post-A1 defect triage; Group 2 (missing homepage) opened (2026-06-03)

**A1 VERIFIED CLOSED.** After the `e9609749` run fully drained, the deployed site confirms it: all five games committed (`games/{auto-battler,economy-simulator,jelly-invaders,p2p-networking,pathfinding}/index.html`), tools deploy, jelly-invaders runs, tested tools run, and `tools/index.html` lists tools with working "Open tool" links to the real pages. The interim 4-of-11 partial was the dispatch throttle (Family J), not an A1 fault. The three-file fix (parser fallback + deploy-time `built_from_plan_version` stamp + flip removal) is confirmed in production. Catalogue A1 → VERIFIED CLOSED.

**Post-A1 walk surfaced the next defects; triaged by root cause into four groups** (catalogue §0b):
- Group 1 (links/lists, ours, highest leverage): re-confirms Families B/C/D. Key contrast — the tool-list resolves real pages (`/tools/<slug>/index.html`) while hero CTAs and game-list emit static placeholders (hero→`/contact.html`/`/services.html`; game-list `Play Now` `href=""`, `img src=""`, fabricated/duplicated set with Jelly Invaders twice; guides hub has no list). One likely root cause; tool-list is the working template.
- Group 2 (missing homepage, ours, structural): STARTING.
- Group 3 (content polish, ours): source-suffix titles, empty footer mailto/tagline (Family E), guide tables rendering poorly (new), guides→tools cross-linking (new enhancement).
- Group 4 (brochure nav feel): design-chat territory (Family G); nav links themselves work.
- Sequence: Group 2 → Group 1; Group 3 after; Group 4 elsewhere.

**Group 2 finding — reframed by the query.** Expected "planner never made the homepage"; instead the row is `name=index, page_type=landing, url=/index.html, build_status=deployed, stamped=t` — yet no `index.html` at the repo root. So it's a **DB-deployed-but-file-absent mismatch** (specific instance of A2; logged as A4), not a planner/unbuilt gap. Candidate cause (NOT concluded): the deploy path marks `deployed` independently of commit success (tool-recreation marks `deployed` before `deploy_page`/`git_commit`), and the git-adapter `updateRef` is `force:false`+no-retry, so a concurrent commit to the shared `sites` repo loses with a silent non-fast-forward (FOCUS_dispatch open item 4). A homepage committing amongst many during the cascade is a plausible loser. Alternatives not ruled out: empty-assembly skip (0 components) or a path bug. Decisive reads pending before any fix: homepage `page_components` count; its rerender work-item error; git-adapter logs for a 422 on `index.html`.

**Dispatch thread (Lever C + guardrails) remains parked** as agreed, ready to resume; the missing homepage may be the first real-world instance of its open-item-4 git race (to be confirmed by the Group 2 diagnosis).

## Principles reinforced (Part 10)
- A "missing page" can be a deploy/commit failure, not an absent plan. The query distinguished them: `deployed`+`stamped`+no-file ≠ never-planned. Always check DB state before assuming the planner skipped something.
- Group defects by root cause, not symptom. Several "separate" bad things (hero CTA, game-list, guides hub) are one link-resolution cause; the working tool-list is the template to diagnose against.

---

## Part 11 — Group 2 root cause: homepage auto-completed after a lost response (2026-06-04)

**Staleness check (user-requested): the index row is current, not stale.** `needs_page:index` completed 2026-06-03 14:36 (18 min after the `e9609749` adoption at 14:18); `page_rerender_index` ran 2026-06-04 02:56; `deployed_at` 2026-06-04 07:04 — all on the live site_id (via domain). The 06-03 `needs_page:index` still existing rules out a teardown+re-adopt since (cascade delete would have removed it). The `c4e1f68f` correlation on the rerender items is a later build-pipeline pass on the same site, not a different run. (Confirm-if-desired: `sites.created_at` should predate 06-03 14:36.)

**Reads ruled out planner gap and (earlier) the git race; pinned the real cause.**
- Plan gave the homepage **6 sections**: `["hero","system-stats","tool-list","guide-list","game-list","call-to-action"]` → not a planner gap. (Note: the planner put `tool-list`/`guide-list`/`game-list` on the homepage — relevant to Group 1.)
- `needs_page:index` (`needs_content_page`) is `complete` with error **"Auto-completed: work verified done despite lost response."**

**Root cause:** the homepage's content build was dispatched, the **handler's response was lost** (pod death / dead in-process timeout goroutine — same family as the "Claim timed out — handler pod likely died" entries and the dev-guide "#1 cause of pipeline stalls"), and the recovery path **optimistically auto-completed** the work item *without verifying the artifact*. Components were never written (`n_rendered=0`), but the item says `complete`, so nothing re-queues it. A later rerender ran on the empty page and a deploy path marked it `deployed`+`stamped` despite 0 components → the reconciler now skips it permanently → no `index.html`.

**Systemic implication (not homepage-specific):** "auto-complete on lost response" falsely completes *any* page whose handler response is lost — artifact-less but marked done. Likely behind some other defects (empty list cards, missing sections) wherever a handler died mid-build. Same handler-death problem as the dispatch claim-timeouts, but a *different, more dangerous* recovery branch: it declares success instead of re-queuing (which Lever C does for claims).

**Fix direction (two-part; not prescribed until the mechanism is read):**
- **A (deeper):** auto-complete-on-lost-response must **verify the artifact** before completing — for `needs_content_page`, confirm `page_components` exist; if not, re-queue (→`triaged`, attempt++) rather than "verified done." Belongs with the Lever-C reliability thread. Needs the orchestration-engine code that sets the "Auto-completed…" string (not in the actions dump — same gap as the "Claim timed out" string).
- **B (guard):** a 0-component page must never reach `deployed`+`stamped` (gate `deployed` on real content), so the reconciler rebuilds it. Same principle as A1/Option B, applied to the landing-page deploy path.
- **Immediate unstick (operational, not structural):** `UPDATE pages SET build_status='needs_rebuild', built_from_plan_version=NULL WHERE … name='index'` to force a rebuild now. Recurs if the response is lost again until A is fixed.

**Decision:** the auto-complete false-positive joins the reliability thread (with Lever C). Next read needed: the orchestration-engine lost-response/auto-complete code (for A) and the landing-page deploy path (for B). Group 2 not yet fixed — cause pinned, fix gated on those reads. Group 1 still queued after Group 2.

## Principles reinforced (Part 11)
- A `complete` work item is not proof of done. "Auto-completed: work verified done despite lost response" marked a build complete with zero artifact — verify the artifact (page_components), not the status.
- Check record freshness by timeline, not assumption: the surviving 06-03 `needs_page:index` proved no teardown since, so the row is this run's.
- Two recovery branches for the same handler-death problem behave oppositely (re-queue vs false-complete); the false-complete one silently loses work and is the more dangerous.

---

## Part 12 — Option A is a SQL pre_query, not Go; the migration already exists; Lever C is redundant (2026-06-04)

**Correction (user-found):** "Auto-completed: work verified done despite lost response" is set by the **`claimed-item-timeout` scheduled task's SQL `pre_query`**, NOT a Go reaper. My request to upload a reaper .go file was based on a wrong assumption. The completion + reset logic lives entirely in that task's pre_query.

**`migration_claimed_item_timeout_evidence_check.sql` (uploaded) is essentially Option A + Lever C, already authored:**
- `completed_by_evidence` CTE = **Option A**: marks a stuck claim `complete` only with positive evidence the *specific* page deployed after claim (`page_id`-specific, `build_status='deployed'`, `deployed_at > claimed_at`). Replaces the old loose "any page on the site updated since claim (updated_at)" check that caused false positives (gaswholesalers fuel-industry-insights, 2026-05-12, per the migration header).
- `reset` CTE = **Mode 3 = our Lever C**: stuck claims past 40 min, no evidence → `attempt_count+1`, `status = CASE WHEN attempt_count+1 >= max_attempts THEN 'failed' ELSE 'triaged'`, clear `claimed_by`/`claimed_at`. This is the stale-claim reset I was about to build as a NEW watchdog. **It already exists** — so the FOCUS_dispatch `reset_stale_claims` watchdog is redundant; do NOT build it.

**Does it fix the homepage?** Yes. At 14:36 06-03 the homepage's `needs_content_page` had no deploy evidence (page not deployed until 07:04 06-04) → under this logic it would NOT auto-complete; it would reset at 40 min and retry. It WAS auto-completed (old loose evidence) → **the migration was not applied as of 06-03.** First action: verify whether applied now; if not, apply.

**Option B still needed (not optional):** the migration's evidence trusts `build_status='deployed'`, which the homepage proves can be `true` with 0 components / no file. Without B, a mislabeled-deployed page can still let this task false-complete a related item. B (UpdatePageStatusAction guard — never mark a 0-component page deployed) makes the flag trustworthy and independently prevents the 07:04 mislabel. The migration reads the flag; B keeps it honest. Complementary.

**Optional hardening:** for `needs_content_page` the produced artifact is `page_components`, not a deploy (content-write `needs_page:index` and deploy `page_rerender_index` are separate items here). Crediting it by `deployed_at` checks the *next* item's output; checking `page_components` existence would avoid re-doing content that succeeded but isn't deployed yet. Not required for the homepage.

**Plan now:** (1) verify the migration is applied; apply if not (Option A core + Lever C reset). (2) Option B guard in current `UpdatePageStatusAction` (need the current file — the upload is pre-A1). (3) Do NOT build a new claim watchdog — the `reset` branch is it. (4) Optional: strengthen `needs_content_page` evidence to `page_components`. (5) Owe the v2_30 guide a §9 row once shipped.

## Principles reinforced (Part 12)
- Reuse-before-build, again: the stale-claim reset and the evidence-gated completion already existed in a scheduled-task pre_query. Searching the DB (not just Go) found them. I almost built a duplicate watchdog.
- An evidence check is only as good as the signal it trusts. `build_status='deployed'` is a derived flag that can lie (the homepage). Evidence should prefer ground-truth artifacts (`page_components`) and the flag should be guarded at its write (Option B).

### Part 12 addendum — revised migration delivered (2026-06-04)
Delivered `migration_claimed_item_timeout_evidence_v2.sql` (supersedes the uploaded `migration_claimed_item_timeout_evidence_check.sql`; idempotent — safe to apply whether or not the prior was applied). Only the `needs_content_page` evidence branch changed: from `pages.deployed_at > claimed_at` to `EXISTS(page_components pc … pc.updated_at > wi.claimed_at)` — i.e. the content artifact this item actually produces, with a recency guard so stale rows from a prior generation can't false-credit it, and decoupled from the untrustworthy `build_status='deployed'` flag. `page_rerender` (deployed_at — correct for it), `needs_design`, the `reset` branch (Lever C), error strings, and the returned column names are all unchanged. Schema confirmed first: `page_components(page_id, component_id, rendered_html, updated_at)` exist. Option B (UpdatePageStatusAction guard) still owed and still gated on the current `v3_site_actions.go`. Apply order: verify current pre_query → apply v2 → (separately) Option B.

### Part 12 addendum 2 — Option A applied; Option B delivered (2026-06-04)
- **Option A APPLIED + verified live:** the v2 migration is in place — `claimed-item-timeout` shows `enabled=t`, `interval_seconds=120`, and the new "provably done on the specific targeted artifact" pre_query (updated_at 14:09). So evidence-gated completion (+ the Lever-C reset) is now running.
- **Option B delivered:** `v3_site_actions.go` patched (`v3_site_actions_optionB.patch`, against the current post-A1 file). `UpdatePageStatusAction` now guards the deployed branch with a new package-local helper `pageHasComponents(ctx, db, pageID)` (EXISTS on `page_components` with non-null `component_id` + non-empty `rendered_html`; mirrors the `lookupPageID`/`execDB` `*sql.DB`/`*pgxpool.Pool` type switch). On 0 components it refuses `deployed`, sets `build_status='needs_rebuild'` + `built_from_plan_version=NULL` (matching the proven unstick), and returns `updated:false`. Fail-open on a check error (a transient failure must not halt legitimate deploys; A is the other layer). Only `newStatus=="deployed"` is affected; the A1 stamp on the has-components path is unchanged; no variable names changed. Validated locally: imports present, braces/parens balanced. Build/gofmt/vet pending on the user side (no Go toolchain in sandbox). Confirmed no pre-existing `pageHasComponents` to clash with.
- **Net for the homepage:** unstuck (needs_rebuild) → rebuilds; if its content handler response is lost again, A re-queues it (retry, not false-complete); B stops it being marked deployed empty. Robustly handled once B is deployed. Owe the v2_30 guide a §9 row for the silent-completion class.

---

## Part 13 — Group 1 root cause confirmed (list resolution) (2026-06-04)

Reads (page_type vocabulary + the three list component schemas) confirm the cause; the architecture (queryresolve + Step-3 anti-fabrication) is already built, so this is per-component schema/template work, not workflow.

**Current-site page_type vocabulary:** `tool` (6), `game` (5 — game-auto-battler/economy-simulator/jelly-invaders/p2p-networking/pathfinding), `section-index` (3 hubs). Guides are `page_type=blog_post` under `/blog/` (the "blog convention"). **Correction to FOCUS_component_schema_patterns.md:** that doc said there was no `game` page_type; the current site HAS it (its recommendation 1 was implemented since), so the doc's headline game-list blocker is gone — game-list will resolve once it's Tier D.

**Per-list findings:**
- **game-list (`game-list_pre_037`)** — legacy numbered-flat anti-pattern: `game1_title…game6_*` all `source:llm` (fabricated → duplicates like Jelly Invaders ×2), `gameN_cta_url`→`site_specs.games.gameN_url` (absent → `href=""`), `gameN_image_url`→`site_assets.image` (→ `src=""`). NOT Tier D. This is the fabrication.
- **guide-list (`guide-list_pre_037`)** — already Tier D (`items`, `source: query.pages_where_type:guide`), but guides are `page_type=blog_post`, so the query returns 0 rows → empty list. Pure vocabulary mismatch.
- **tool-list** — Tier D + tools=`tool` → resolves ✓. Its empty `"Browse All Tools" href=""` is the cta_url issue (`source: site_specs.identity.tools_index_url`, unpopulated, no fallback → dropped).

**Group 1 = two root causes:** 1A list resolution (game-list not Tier D; guide-list wrong vocabulary) + 1B cta_url (unpopulated site_specs paths + no fallbacks → empty hrefs; hero CTAs to generic `/contact.html`/`/services.html` are a separate hero-component matter).

**Fix plan (priority):** (1) game-list → Tier-D items-array querying `pages_where_type:game` (games exist → resolves) — biggest win. (2) guide-list vocabulary. (3) cta_url fallbacks. (4) hero CTAs separate.

**Decisions pending (product/structure, not mechanical):**
- game-list richness: pages has only url/title/meta_description, so genre/rating/platform/filter can't be real → recommend simplifying to a tool-list-parity card list (drop the fabricated decoration) vs keeping the rich layout (needs a data source that doesn't exist).
- guide vocabulary: band-aid (query `blog_post`) vs structural (add `guide` to the classifier + re-type guides, parity with `game` and the tools/guides/games mission) — user's structural-over-quick preference favours the latter but it's bigger (classifier + re-type + `/blog/`↔`/guides/` URL question).

**Need to write game-list:** tool-list `html_template` (migration 041) as the canonical Tier-D reference (exact `{{range .items}}` + funcMap conventions) — the read returned schemas, not templates.

## Principles reinforced (Part 13)
- Verify the doc against the live DB: FOCUS_component_schema_patterns.md said `game` didn't exist; it does now. Acting on the stale claim would have mis-scoped the fix.
- The contrast (tool-list works, others don't) decomposed into two independent, well-understood causes (Tier-D coverage + page_type vocabulary), both already-known deferred follow-ups — not a new defect.

### Part 13 addendum — game-list Tier-D migration delivered; decisions locked (2026-06-04)
**Decisions (user):** (1) game-list → simplify to tool-list-parity card list (drop fabricated genre/rating/platform/filter — `pages` has only url/title/meta_description); (2) guides → first-class `page_type=guide` (structural: classifier + re-type existing rows + `/blog/`↔`/guides/` URL question).

**Delivered `migration_game_list_tier_d.sql`** (validated): rewrites `game-list_pre_037` from the 50+-field numbered-flat anti-pattern to the Tier-D items-array, mirroring tool-list. Field vocabulary IDENTICAL to tool-list (items, cta_url, cta_label, eyebrow_label, section_intro, card_link_label, section_heading, cta_supporting_text) so the Step-3 content-writer + `merge_with` path treats it the same; only values differ — `items.source: query.pages_where_type:game` (resolves now that games are `page_type=game`) and `cta_url` gains `fallback: /games/index.html` (tool-list lacks one → its "Browse All Tools" renders `href=""`; the fallback fixes that class). html_template mirrors tool-list's `{{range .items}}` conventions with `gl-*` classes + a play-triangle icon. UPDATE in place (resolution is by function; single game-list row); quality metadata left (render reads only template/schema). Validated locally: dollar-quotes balanced, schema parses to 8 fields, placeholders mirror tool-list. It's a content_components DATA change → live on COMMIT, no code deploy. Trailing (run after commit): `UPDATE pages SET build_status='needs_rebuild', built_from_plan_version=NULL WHERE … sections @> '["game-list"]'` (games hub + homepage) to re-resolve.

**Still open in Group 1:** guide structural fix (needs the classifier — likely `003_site_classifier.sql`; plus re-type existing guide rows and decide `/blog/`↔`/guides/` URLs). cta_url fallbacks for tool-list/guide-list deferred until the `fallback`-on-`site_specs`-sourced-url mechanism is confirmed in `plan_sections_action.go` (else populate site_specs); the doc's follow-up implies fallback is honored. Hero CTAs separate.

### Note — DB-change snapshot standard (2026-06-04, user instruction)
Every migration / manual DB change must include a real, restorable snapshot taken BEFORE the change (not just a diagnostic "before" SELECT). Convention: `CREATE TABLE IF NOT EXISTS <table>_bak_<change> AS SELECT * FROM <table> WHERE <affected rows>` inside the txn before the UPDATE (in-txn AS-SELECT captures pre-change values), with the rollback statement documented in the migration. `migration_game_list_tier_d.sql` updated to this standard (backs up the content_components row; and a pages snapshot before the rebuild flip). NAMEDATALEN=63 — keep backup names short. (The already-applied claimed-item-timeout v2 captured its pre_query via SELECT — text, restorable by copy; future ones use backup tables.)

### Part 14 — Guide page_type: classifier located, schema checked, two migrations (2026-06-04)
**Where page_type is classified:** the `site-adoption-agent` (id 4e2d8e8e-47a7-476a-95ca-8d71f32e894a), in its `analyze_site` execute_llm_prompt step. The LLM emits each page's `page_type` AND `url`. LIVE enum: `content|tool|game|section-index|blog-index|blog-post|landing`; guidance line explicitly folds guides into blog-post ("individual article, guide, or essay"). `site-classifier` = site-type only (not pages). `chief-strategist` = component types, not page_type. `ifpClassifyPage` (chassis) = hint only. SECOND planner: `build-site-planner` (id f263eaa1) has its own STALER vocabulary `index|content|landing|entity-directory|entity-page|tool|blog-index|blog-post` (no game/section-index/guide) but is told to PRESERVE existing pages' page_type verbatim, so adopted game/guide rows survive it; only NEW guide planning there would need the addition.

**Schema (checked before SQL):** `pages.page_type` = varchar(50), only constraint `chk_page_type_kebab_case` (kebab FORMAT, NO value allowlist). `guide` is valid format → re-type allowed. `queryresolve.go` NOT on disk (in listing, absent) — guide-resolution inferred from the working `game` precedent (game resolves via pages_where_type:game with no resolver change).

**Live rows (gamesdesign):** 5 real guides `guide-*` (blog-post, /blog/guide-*.html, sections=["generic-text-block"]) + 5 EMPTY bare duplicates `economy-basics|fairness-in-rng|p2p-architecture|rng-design|skinner-box` (blog-post, sections=[]) + `guides-index` (section-index, /guides/index.html, sections=[]). Duplication = separate defect (adoption/planner emitted both guide-X and bare X); NOT auto-converted.

**Delivered (both with snapshot+rollback):**
- `migration_retype_guides_to_guide.sql` — re-type ONLY the 5 `guide-*` (name LIKE 'guide-%' AND page_type='blog-post') -> guide. Snapshot `pages_bak_retype_guides`. Trailing (commented) guide-list rebuild flip with its own snapshot `pages_bak_guidelist_rebuild`. Coverage note: guide-list is on the HOMEPAGE; the guides hub `guides-index` has sections=[] so re-typing won't populate the hub — hub needs a guide-list section added (separate, same shape as tools hub got tool-list).
- `migration_adoption_add_guide_page_type.sql` — going-forward: 2 quote-free/newline-free replace() edits on default_config::text->jsonb: (1) enum +guide; (2) narrow blog-post guidance + route instructional content to guide. Snapshot `agent_definitions_bak_adopt_guide`. Pre-check counts anchors (each must = 1), verify confirms. Optional coherence tweaks (section-index "+ guide", blog-index example off /guides/) noted, not applied. Caveat: apply_adoption_plan / validate_site_plan not on disk — confirm a future adoption persists page_type='guide'.

**Separate/parked from this:** empty bare-duplicate guide pages (investigate build_status/page_components/created_at before deleting); guides hub empty sections (needs guide-list); /blog/->/guides/ URL migration (needs real page_canonical.go, not on disk; touches CanonicalisePage+nav+link_registry); build-site-planner vocabulary for NEW guides.

### Part 14b — Guide migrations APPLIED; queryresolve+canonicaliser confirmed; URL migration prepared (2026-06-04)
APPLIED: migration_retype_guides_to_guide.sql (UPDATE 5 — guide-* now page_type=guide) and migration_adoption_add_guide_page_type.sql (enum+guidance, anchors 1/1, verify t/t). Rebuild flip for guide-list NOT yet run (do alongside game-list flip to refresh homepage).
CONFIRMED FROM SOURCE (uploads): queryresolve.resolvePagesWhereType keys on page_type=$2 with status IN ('active','deployed'), value-agnostic, reads pages.url — so guide-list resolves the 5 guides on homepage rebuild. page_canonical.go ALREADY has a `guide` case: role=guide -> name guide-<bare>, url /guides/<bare>/index.html, page_type guide (nested, peer of tools/games). Live guides are at /blog/guide-*.html (blog-post origin) — NAME already canonical, only URL diverges.
build-site-planner (live, v1052) canonical page_type list still lacks game/section-index/guide but preserves existing pages verbatim + validate_site_plan tolerates non-listed types (games already coexist) — so guides survive it; adding guide there is optional hardening for NEW guide planning.
PREPARED (not yet applied): migration_guides_url_to_canonical.sql — moves the 5 guides /blog/guide-<slug>.html -> /guides/<slug>/index.html (url = '/guides/'||substring(name from 7)||'/index.html'), snapshot pages_bak_guides_url, build_status flip. Gated on 3 diagnostics (link_registry rows targeting guides; navigation_structures with /blog/guide-; page_components rendered_html with /blog/guide-) to size: rebuild of linking pages + git cleanup of orphaned /blog/ files. guide-list links are dynamic (auto-update on rebuild).

### Part 14c — Bare-guide duplicates = documented "planner ignores adopted state" (2026-06-04)
URL migration APPLIED (UPDATE 5; guides now /guides/<slug>/index.html, needs_rebuild). All 3 ref diagnostics for /blog/guide-% returned 0 rows — orphaned /blog/guide-*.html files are harmless; deletion optional.
Bare empties (economy-basics, fairness-in-rng, p2p-architecture, rng-design, skinner-box) — prior analysis FOUND:
- FOCUS_planner_ignores_adopted_state.md (2026-05-19): build-site-planner generates a generic skeleton ignoring adopted pages; reconciler then creates spurious/duplicate pages on top. Line 41 names adoption's output as the guide-PREFIXED pages (guide-skinner-box, guide-rng-design) = faithful set; "everything else is parallel invention." Root: doc 029 — two surfaces (adoption + planner) both write pages/queue work, don't read each other's realised state. Fix: reconciler as sole queue producer (can't dup by construction) = Phase 1.
- HANDOFF_2026-05-11: confirms the 5 bare names are gamedesign.uk's ACTUAL guide topics; documents page-content-writer FABRICATING guide cards; notes "NO page_type='guide'; guides are blog_post" (state we fixed).
Caveats: (1) docs cover the MECHANISM + name the guide-X adoption pages, but not this exact guide-X vs bare-X pairing — fresh instance of the known pattern. (2) Phase 1 PARTIALLY landed: live build-site-planner now has "Existing Pages — PRESERVE EXACTLY" (examples lifted from the FOCUS doc), yet this 06-03 adoption still produced the bare empties — so the instruction stops rename/replace but NOT a differently-slugged sibling for an already-adopted topic (Mechanism 1 / sub-issue #2). blog-content-planner (plans 3-4 posts incl. one "guide" post) is also a candidate creator of the bare set.
NEXT before deleting: (a) provenance query (created_at order + n_rendered to confirm guide-X=adoption/content, bare=later/empty); (b) ref check on the BARE /blog/<topic>.html urls (earlier check was /blog/guide-% only); (c) site_work_items needs_page:* for these names to identify the creating surface (tighten source, not just delete symptom). Then delete migration WITH snapshot.
Still queued: guides-hub guide-list section (guides-index sections=[]).

### Part 14d — Bare duplicates: provenance CONFIRMED; durable cleanup prepared (2026-06-04)
Provenance query result: guide-<topic> (5) created 2026-06-03 14:36:50 (adoption batch), sections=["generic-text-block"], now page_type=guide + /guides/<slug>/index.html + needs_rebuild; n_rendered 3 of 5 (guide-fairness-in-rng + guide-p2p-architecture = 0, will build on rebuild — watch silent-completion). Bare (5) created 2026-06-03 20:25:30 (separate batch ~6h later), blog-post, sections=[], build_status=planned, n_rendered=0 — never built, inert. link_registry refs to bare /blog/<topic>.html urls = 0. => bare set = the documented post-adoption second-surface invention; safe to remove.
FK picture (schemas_all): pages-delete CASCADES page_components/link_registry.source/flow_pages/research_results, SET NULL site_nav_items; BLOCKS on non-cascading link_registry.target_page_id / page_component_history / redirects. site_work_items.page_id has NO FK (work items linger; non-terminal ones hold idx_swi_dedup). reconcile diffs plan vs realised => must also remove from current site_plan_pages (+ site_plan_sections; sections FK is to site_plans not plan_pages) or reconciler re-creates.
PREPARED: migration_cleanup_bare_guide_duplicates.sql — self-guarding (DO block aborts if non-cascading refs exist), snapshots 4 tables (pages_bak_del_bare, splanpages_bak_del_bare, splansecs_bak_del_bare, swi_bak_del_bare), terminalises work items (wont_fix), deletes current-plan plan-sections + plan-pages, deletes pages. Durable vs reconciler; does NOT stop build-site-planner/blog-content-planner re-inventing on a future plan_site run.
RUN FIRST: A1 (site_work_items needs_page:<bare> source/created_by — identify creating surface for the UPSTREAM fix; sections=[] hints build-site-planner not blog-content-planner) and A2 (are bare names in current site_plan_pages — recreation risk / whether plan-deletes do real work).
Still queued: guides-hub guide-list section (guides-index sections=[]); upstream planner-gap fix (stop differently-slugged sibling of adopted topic).

### Part 14e — Bare-duplicate SOURCE confirmed = planner prompt gap (2026-06-04)
A1: needs_page items source=reconcile_site_plan (innocent — queues the plan delta). A2: 5 bare names ARE in CURRENT plan 77d88a60 (written by build-site-planner 2026-06-03 20:25:26) as role=blog-post → cleanup MUST remove from site_plan_pages (migration does) AND recurs on any fresh build-site-planner run.
DECISIVE (llm_call_log plan_site @ 20:25:22): saw_guide_pages=t, prompt_says_no_existing=f, planned_bare_in_response=t; existing_pages_block lists "guide-economy-basics | page_type: blog-post | url: /blog/guide-economy-basics.html" etc. So the planner WAS given the adopted guides and emitted economy-basics anyway → PROMPT-RULE gap (the "no sibling versions" rule only shows games/games-index + tool-rename; didn't generalise to guide- prefix sibling). NOT a wiring/status gap.
FIX (recommended, structural, Go): deterministic guard in validate_site_plan/write_site_plan — drop a planned page whose topic STEM (role-prefix stripped, as CanonicalisePage computes via TrimPrefix "guide-"/"tool-"/"game-") collides with an existing page. Guarantee, not LLM-dependent. NEEDS validate_site_plan/write_site_plan source (NOT on disk) — request it; reuse CanonicalisePage prefix-stripping, no parallel stem fn.
STOPGAP (optional, prompt): migration_planner_topic_sibling_rule.sql — one quote-free replace() appending a topic/prefix/role-duplicate rule (guide-economy-basics example) after "but never duplicate, replace, or rename an existing page." Snapshot agent_definitions_bak_planner_sibling + pre-check anchor=1 + verify. Nudge only.
APPLY-NOW: migration_cleanup_bare_guide_duplicates.sql is good to apply (A2 confirms plan-removal needed); clears the live symptom. Land the source fix before the next build pass or cleanup is undone.

### Part 14f — Cleanup durable; list-page flip + guides-hub diagnostic (2026-06-04)
VERIFIED: current-plan bare-name query returns 0 rows -> cleanup durable, reconciler won't recreate. Current plan = 17 clean pages (index, about, contact, 5 game-* + games-index, 6 tool-* + tools-index, guides-index). DIVERGENCE NOTED: plan has guides-index but NO individual guide pages; the 5 guide-* live only in pages (realised), not the plan. games/tools have both hub+items in plan. Likely tied to the empty guides hub.
MECHANISM (from files): reconcile_site_plan_action.go has NO list-section logic (queue producer only); no deterministic code path attaches tool-list/guide-list to a section-index page -> hub list sections come from the PLAN (LLM emits per-page sections). So empty guides-index = plan gap (tools-index got tool-list, guides-index got nothing). resolvePagesWhereType confirmed: guide-list resolves page_type='guide' AND status IN(active,deployed) -> typed guides will populate on rebuild.
THREAD A (deliverable now): migration_rebuild_list_pages.sql — snapshot pages_bak_relist + flip build_status='needs_rebuild' for pages whose sections array contains an element LIKE 'guide-list%' OR 'game-list%' (robust to versioned names; jsonb_typeof guard). guide-list IS Tier-D -> populates. game-list populates ONLY IF Tier-D (pre-check (0) reports query_sourced per component); apply migration_game_list_tier_d.sql first if game-list still flat. If step (2) shows 0 rows, section strings differ -> adjust match.
THREAD B (diagnostic this turn, fix next): diagnostic_hub_sections.sql (read-only) — (1) realised pages.sections for index/tools-index/games-index/guides-index, (2) current-plan site_plan_sections for same, (3) guide-list/tool-list component identity+tier. Goal: mirror the working tools-index list section onto guides-index (plan + realised + rebuild) once I see whether tool-list lives in site_plan_sections, the realised page, or both, and the exact component_name.

### Part 14g — list flip applied + BOTH empty hubs fixed (2026-06-04)
THREAD A APPLIED: pre-check shows game-list_pre_037, guide-list_pre_037, tool-list all query_sourced=t. Flip hit 2 pages: index (homepage, was needs_rebuild) + games-index (was deployed -> needs_rebuild). On rebuild they re-resolve tool/guide/game lists.
THREAD B DIAGNOSIS (corrected assumption): tools-index is ALSO empty (sections=[]), not just guides-index. Only games-index is populated (["hero","game-list"]). Current plan: games-index has hero+game-list; tools-index & guides-index have NO site_plan_sections at all -> planner populated one hub, left two bare. Naming convention confirmed: site_plan_sections.component_name + pages.sections both use FUNCTION names (game-list/tool-list/guide-list), not *_pre_037 (resolved downstream). guide-list_pre_037 sources_guides=t.
BUILD SECTION SOURCE confirmed: page build reads pages.sections (chassis "SELECT sections FROM pages", "consistent with how plan_sections reads it"). So fix sets BOTH pages.sections (build reads it) AND site_plan_sections (plan stays consistent, no re-empty on re-sync).
THREAD B FIX: migration_populate_empty_hubs.sql — snapshots splansecs_bak_hubfix + pages_bak_hubfix; DO-block guard (abort if hubs already have plan sections -> re-run safe); INSERT plan sections hero(0)+tool-list(1) for tools-index, hero(0)+guide-list(1) for guides-index; UPDATE pages.sections to ["hero","tool-list"] / ["hero","guide-list"] + build_status=needs_rebuild; verify. Mirrors working games-index. tool-list/guide-list Tier-D -> resolve tools/guides on rebuild.
STRUCTURAL FOLLOW-UP (noted, not built): planner inconsistently populates section-index hubs (one hub got a list, two didn't; plan also has guides-index but no guide pages). Durable fix candidate = a deterministic step (same ValidateSitePlanAction/reconcile area as the itemStem dedup) that guarantees every section-index hub gets its matching list section. Reasonable-step-size: deferred; data fix resolves gamesdesign now.
SHARED CAVEAT (A+B): guide-list cards link to /guides/<slug>/index.html; the 5 guide pages are needs_rebuild, so links 404 until they build+deploy (separate in-flight step).

### Part 14h — ROOT CAUSE revised: adoption-faithfulness convergence is INERT (2026-06-04)
Earlier sibling diagnosis ("Pass C stem too narrow") was WRONG/incomplete. TRUE root cause: reconcilePlanWithRealised gates on rm["adoption_locked"]; the live load_existing_pages query does NOT emit adoption_locked (backup line169 query = name,page_type,url,title,nav_label,in_header,in_footer only) -> lockedPages always empty -> reconcile ALWAYS no-ops. Explains all three: bare sibling survived, 5 guides never unioned into plan (Pass A), AND my itemStem Pass C2 is dead code (sits past the early return).
DOC CONFIRMS: FOCUS_adoption_faithfulness_via_locks.md status — convergence "Inert until 054 + write_site_plan land." Subsystem pieces: 053 schema (lock_type+lock_expires_at on 4 tables) "written, ready to apply"; 054 load_existing_pages exposes adoption_locked "written (SQL)"; write_site_plan locks adopted pages adoption/timed/90d on first plan (pending); validate convergence "done (v3_site_actions.go)" = the inert reconcile.
LIVE STATE: schemas_all shows lock tables have ONLY locked_at/locked_by — NO lock_type/lock_expires_at -> 053 NOT applied. Backup query lacks adoption_locked -> 054 NOT applied. 031_locks.md lines142-146 = the Pattern-A+ extension 053 implements (lock_type 'permanent'|'timed'|'review', lock_expires_at; read = locked_at IS NULL OR lock_expires_at<NOW). QueryDatabaseAction scans bool cols as Go bool (so adoption_locked bool expr satisfies .(bool); only []byte stringified).
adoption_locked semantics (doc 86-116): first-plan branch = "no current plan + pages exist" (NO schema needed); re-plan branch = page has live timed lock (needs 053 + write_site_plan locking). From-scratch (no plan + no pages) = false.
PROPER FIX = land the subsystem in order: 053 schema -> 054 query (emit adoption_locked) -> write_site_plan locking (Go; transferDirectiveLocks plumbing already in write_site_plan_action.go) -> itemStem Pass C2 (already delivered) becomes live. MINIMAL schema-free path available: 054 first-plan branch only (adoption_locked = NOT EXISTS current plan) activates convergence for first passes without 053.
REUSE: doc says 053/054 "written" -> requested the files from user before recreating (per reuse principle). No 053_/054_ files found on mounts. verify_adoption_lock_subsystem.sql delivered (read-only: lock columns + live load_existing_pages query + pages lock cols + log-grep note).
GAMESDESIGN CAVEAT: its first pass already ran inert; current plan exists (lacks guides). Activating convergence helps FUTURE first passes, NOT gamesdesign's subsequent passes (adoption_locked=false once a plan exists). Remediation = re-run first pass (delete current plan) OR one-off lock adopted pages + re-plan. Sequence later.

### Part 14i — 054 IS live; real proper fix = union must carry sections (clobber) (2026-06-04)
CORRECTION (again): earlier 053/054 "not applied" was from STALE dumps. Live state: 053 APPLIED (lock_type+lock_expires_at on assets, page_components, site_components, site_plan_directives; NOT pages -> adoption lock is DIRECTIVE-based). 054 APPLIED — live load_existing_pages emits adoption_locked via FIRST-PLAN branch only: CASE WHEN NOT EXISTS(current plan for site) THEN true ELSE false. No re-plan/directive-lock branch. So convergence IS active on the first post-adoption pass. My "inert" call was wrong.
20:25 explained: 054 post-dated the gamesdesign build (May-25 backup lacked it). Corroborated: guides kept realised sections + absent from plan = Pass A never ran then.
NEW CONFIRMED BUG (the real proper fix): activating reconcile exposes a clobber. upsertPage (site_db_actions.go ~1107) ON CONFLICT does sections=EXCLUDED.sections, meta_description=EXCLUDED.meta_description, nav_order=EXCLUDED.nav_order (nav_label is COALESCE-preserved). normaliseRealisedToPlanPage set sections=[] and the query didn't select sections/meta_description/nav_order -> Pass A union of an LLM-omitted adopted page writes empty values -> sync clobbers the adopted page's real sections/meta/nav_order to empty. (At 20:25 reconcile didn't run, so omitted guides were left untouched in realised, not unioned-and-clobbered — which is why their sections survived.)
write_site_plan locking: transferDirectiveLocks (lines 637-700) copies locked_at/locked_by ONLY (no lock_type/lock_expires_at) and nothing creates an adoption/timed/90d lock -> 90-day RE-PLAN window is non-functional. But the first-plan branch needs none of it, so re-adoption converges regardless. (Re-plan locking = separate lower-priority follow-on: write_site_plan create timed lock + extend transferDirectiveLocks.)
QueryDatabaseAction scans jsonb as []byte->string; so rm["sections"] arrives as a JSON STRING -> normaliseRealisedToPlanPage parses it (json.Unmarshal), tolerates native []interface{}.
DELIVERED (proper fix, both must land together):
 (a) migration_load_existing_pages_carry_fields.sql — adds p.sections, p.meta_description, p.nav_order to the load_existing_pages SELECT (snapshot agent_definitions_bak_lep_carry; pre-check anchor=1; verify). Shared planner; harmless where reconcile no-ops.
 (b) v3_site_actions.go normaliseRealisedToPlanPage now carries sections(parsed)+meta_description+nav_order. Same file as itemStem Pass C2. Patch regenerated (v3_site_actions_itemstem_dedup.patch, 118 lines). Needs user gofmt/vet/build/deploy (no toolchain here; braces 1249/1249 parens 1589/1589).
SEQUENCE: deploy v3 (Pass C2 + carry) + apply migration (a) together -> re-adopt gamesdesign from scratch -> verify: (1) planner log "reconciled with adoption-locked pages" unioned_in>0 (and dropped_collision>0 if siblings proposed) = reconcile ran; (2) adopted pages (guides) KEEP their sections/meta after sync (no clobber); (3) no bare economy-basics sibling; (4) guides in plan.
STILL SEPARATE: empty-hub planner gap (LLM omits list sections for tools-index/guides-index) — convergence won't fix; hub-convergence follow-up. data-fixed for current site.

### Part 14j — re-adoption triggered + gate verified; source hubs ARE populated (2026-06-05)
RE-ADOPTION fired: gamedesign.uk -> gamesdesign.co.uk via orchestrate request on system.agent.generic.requests. correlation 01534a57-ca61-4d1f-9337-4a663b85de16, orchestration a0f6b3e5-4f2d-4949-b41f-67a9cc0fffbf, 2026-06-05 08:18:51Z.
GATE (G1) VERIFIED: SELECT current is_current plan for gamesdesign -> 0 rows. So NO current plan -> on the build pass load_existing_pages computes adoption_locked=TRUE (first-plan branch) -> convergence (reconcile + Pass C2 + union-carry) WILL engage. (Either a teardown happened or adoption produced a fresh state; either way the empty-plan gate is met.)
RE-ADOPTED STATE (faithful, verified):
 - guides typed page_type=guide directly with sections=["generic-text-block"] (build_status planned). So migration_adoption_add_guide_page_type.sql worked — adoption classifier emits guide, NO post-hoc re-typing needed this run.
 - hubs carry list sections FROM SOURCE: guides-index=["guide-list"], tools-index=["tool-list"], games-index=["game-list"]; index=["hero","features","tool-list","guide-list","game-list","call-to-action"].
 - current site_plan_pages: 0 rows (no plan yet — build pass hasn't run / hasn't written one).
KEY REFRAME: source hubs ARE populated -> the OLD site's empty hubs (tools-index/guides-index sections=[]) were the UNION CLOBBER (normaliseRealisedToPlanPage sections=[] -> sync upsert sections=EXCLUDED.sections), NOT a planner gap. So the carry fix (a)+(b) protects hubs; the separate "hub-convergence" follow-up is NOT needed for adopted sites. (The data-fix migration_populate_empty_hubs covered the old site.)
PAIRED-FIX WARNING (critical for this run): because the gate is met, the build pass WILL union LLM-omitted adopted pages. With (a)+(b) deployed -> carries sections -> hubs/guides survive. WITHOUT (a)+(b) -> union clobbers the now-faithful hub+guide sections to []. So this is the run where the clobber bites HARDEST (there is real faithful content to lose). Deploy (a) migration_load_existing_pages_carry_fields.sql + (b) v3 (Pass C2 + carry) BEFORE the build step runs. Build looked decoupled from adoption (June-3 build trailed adoption by hours on its own correlation) -> window exists to confirm deploy.
DELIVERED: verify_readoption_fix.sql (staged: G1 gate + commented G2 plan-retire [only if (a)+(b) live] + during-build reconcile-log grep + P1 clobber check on guide-*/page_type=guide + P2 sibling/plan check + P3 hub check). G2 commented (safe by default); file otherwise read-only.
POST-BUILD VERIFY ORDER: (1) planner log "reconciled with adoption-locked pages" unioned_in>0 (dropped_collision>0 if sibling proposed) = reconcile ran; (2) P1 guides keep ["generic-text-block"] + P3 hubs keep *-list = no clobber (carry fix in effect); (3) P2 no bare economy-basics + guides in plan. All-zero reconcile counts despite empty-plan gate => existing_pages not reaching validate_plan (next thing to chase).

### Part 14k — docs updated (2026-06-05)
Updated TWO project docs (not the running log) to capture this subsystem's corrected state:
 - FOCUS_adoption_faithfulness_via_locks.md: top-of-doc pointer + appended "Verified landed state (2026-06-05)" section. Corrects the [done] markers (053 live; 054 first-plan branch ONLY; write_site_plan Changes 1-3 NOT deployed -> 90-day re-plan window non-functional; only the first-pass protection works). Documents the union-clobber bug + (a)+(b) fix, the itemStemOf Pass C2 dedup, the empty-hub clarification (source hubs populated), and guide as a first-class page_type. Re-prioritised follow-ons (re-plan locking lower priority).
 - 016_debugging_guide_v2_31.md (bumped from v2_30): §9 two new failure patterns — "Adoption convergence clobbers adopted page sections (union carry gap)" and "Adoption convergence is a no-op (reconcile never runs)", each symptom/why/diagnose/fix; §6.5 extended with guide page URL nesting (/guides/<slug>/index.html) + guide-list query.pages_where_type:guide Tier-D resolution. Cross-reference the FOCUS doc rather than duplicating.
DID NOT touch 029/030 (planner/reconciler narrative) or 031 (lock design) — offered to fold the same "written-but-not-deployed" corrections into 031 if wanted (the [done]/written statuses there for write_site_plan locking are the same overstatement).

### Part 14l — DEFINITIVE root cause: []map[string]interface{} vs []interface{} type mismatch (2026-06-05)
Re-adoption result (plan 337fb25d, 12:51, is_current): n_pages=21, n_guide_prefixed=0, n_bare=5. SINGLE plan, NO converged predecessor. Guides all status=active (so load_existing_pages would return them). => NOT multi-pass overwrite, NOT status filter. The FIRST pass itself no-op'd despite adoption_locked-should-be-true (G1 earlier = 0 current plans at 12:51 load time) AND existing_pages populated for plan_site (LLM derived economy-basics by stripping guide- off the adopted guide-economy-basics).
ROOT CAUSE (definitive, code-verified): QueryDatabaseAction output_format=array returns `[]map[string]interface{}` (chassis: `var results []map[string]interface{}`; `return results`). ValidateSitePlanAction does `if ep, ok := ev.([]interface{}); ok { existingPages = ep }` — and in Go `[]map[string]interface{}` does NOT satisfy `.([]interface{})`. So the assertion ALWAYS fails, existingPages stays empty, reconcilePlanWithRealised early-returns (lockedPages empty) => NO-OP for EVERY site. SILENT (no log on empty). The whole adoption-faithfulness convergence (doc 029 Phase 1) has never functioned since deploy. Pass C2 + the carry fix are downstream of this -> also dead until now.
This SUPERSEDES the "inert because adoption_locked not emitted" (14h) and lock-window (14i) framings: 054 IS live, adoption_locked IS emitted and would be true on a first pass; the killer was the type assertion swallowing existing_pages. The discipline (verify, don't prescribe the lock fix) is what surfaced it.
FIX (v3_site_actions.go, added to the existing Pass C2 + carry changes): ValidateSitePlanAction extraction now `switch ev.(type)` handling BOTH []interface{} AND []map[string]interface{} (converts the latter to []interface{} so reconcile's per-element rp.(map[string]interface{}) works); PLUS a Logger.Info "existing pages loaded for convergence" count so an empty set is never silent again. Patch v3_site_actions_itemstem_dedup.patch now 147 lines. Needs gofmt/vet/build/deploy (no toolchain here; braces 1254/1254 parens 1597/1597).
v3 now contains FOUR changes: (1) existing_pages type fix [KEYSTONE], (2) itemStemOf, (3) normaliseRealisedToPlanPage carries sections/meta/nav_order, (4) Pass C2 item dedup. (2)-(4) only take effect because (1) lets reconcile run.
RE-CONVERGE SEQUENCE: deploy v3 -> retire current plan 337fb25d (migration_retire_current_plan_for_reconverge.sql, snapshot) so next build's load_existing_pages sees NO current plan -> adoption_locked=true -> re-trigger build (re-adopt, or build-only trigger) -> reconcile NOW runs: unions guide-* (with sections via carry), drops the 5 bare siblings (Pass C2). VERIFY: planner log "existing pages loaded for convergence" existing_pages>0 (was effectively 0 before) AND "reconciled with adoption-locked pages" unioned_in>0 dropped_collision>0; new plan has guide-* and NO bare names; guides keep ["generic-text-block"].
FOLLOW-UP: the bare PAGES (economy-basics etc., active/planned) persist after reconcile drops them from the plan -> orphaned (won't deploy, not in plan); run the bare-page cleanup again AFTER confirming reconcile converges.
DOC CORRECTION OWED: the FOCUS "Verified landed state" + debugging-guide "no-op" entries (written 2026-06-05) say convergence runs on the first pass — now known WRONG (type bug made it never run). Correct: convergence was INERT due to the []map vs []interface mismatch; first-pass protection only works after this v3 fix.
