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
