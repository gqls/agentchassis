# Defect & Changes Catalogue — gamesdesign.co.uk (post sync-fix adoption run)

**Date:** 2026-05-28
**Run:** clean re-adopt of `gamedesign.uk → gamesdesign.co.uk` on chassis image carrying the `SyncPagesToDBAction` `ValidateRoles` fix + `built_from_plan_version` companion. Current plan `89b70231-a030-4671-89e1-2496d2262b8b`.
**Purpose:** Enumerate every observed defect as a *separate* item before fixing, so distinct causes are not conflated into one rolling investigation. Causes marked "tentative" have NOT been pinned — they are starting hypotheses, to be confirmed by reading the relevant action/source before any fix.

---

## 0. What landed and is verified closed (do not reopen)

The section-index hub reversion — the thread this run was built to test — is fixed and verified end-to-end:

- All three hubs deployed as `section-index` at nested URLs: `games-index → /games/index.html`, `tools-index → /tools/index.html`, `guides-index → /guides/index.html`.
- `built_from_plan_version` populated on all three, matching the current plan (`89b70231…`) — the companion fix works; the reconciler should stop re-emitting these as stale.
- `pages` and `site_plan_pages` agree on the section-index/nested form for the hubs (divergence closed).
- Rendered header and footer across all three index pages link to the nested hub URLs (`/tools/index.html`, not flat `/tools-index.html`).
- Work items settled: `needs_page` 21 complete, `needs_rerender` 2 complete, `page_rerender` 26 complete.

The nav and the index-page links are correct. Everything below is separate from this and was either pre-existing (parked) or surfaced by getting far enough to see the rest of the site.

**A1 (tool/game pages never deployed) is now also verified closed** — see Family A. All games/tools deploy and run; the tools hub lists them with working links.

---

## 0b. Post-A1 verification triage (2026-06-03)

With A1 closed, a walk of the deployed site surfaced the next defects. Grouped by root cause (not symptom), because several share one:

**Group 1 — link & list resolution [ours; highest leverage; likely one root cause].** Re-confirms Families B, C, D on the post-A1 build. Diagnostic contrast: the **tool-list resolves real pages** (`href="/tools/<slug>/index.html"`) while the hero CTAs and the game-list emit static placeholders. Symptoms: hero "Launch a Tool"→`/contact.html`, "Browse All Tools"→`/services.html` (nonexistent) on every hub (B); the tool-list's own "Browse All Tools" `href=""` (B); game-list `Play Now` `href=""`, `img src=""`, and a fabricated/duplicated set — Jelly Invaders appears twice (C); guides hub has hero only, no guide-list (D). Fix hypothesis: make game-list/guide-list use the tool-list's real-page resolution, and rewrite hero CTAs to real targets. → next, after Group 2.

**Group 2 — missing root homepage [ours; structural]. STARTING.** See A4: homepage is `deployed`+`stamped` in DB but absent from the repo — a DB/file mismatch, candidate cause the silent git-adapter commit race. Diagnose before fixing.

**Group 3 — content polish [ours; lower urgency].** Tool-card titles carry the source suffix "- GameDesign.uk"; footer `mailto:`/tagline empty (Family E, partly improved — hero H1s are now distinct and good). New: guide tables render poorly (needs a deployed guide *content* page to diagnose); guides should cross-link to the relevant tools (enhancement).

**Group 4 — brochure-style nav feel [design-chat territory].** Nav links work; the aesthetic is the issue (Family G). Left to the design/composition chat per the standing split; flagged back only if structural.

Sequence agreed: **Group 2 → Group 1**, Group 3 after, Group 4 elsewhere.

---

## Family A — Deployment gaps (highest user impact)

**A1. Tool and game *pages* did not deploy at all. [VERIFIED CLOSED, 2026-06-03]**

*Original symptom:* `pages` had rows for every tool/game with `complete` work items, but the repo's `/tools/` and `/games/` held only the hub `index.html` — no child tool/game files. The interactive tools/games (the site's actual product) never reached deployed output.

*Cause pinned (two coordinated root causes):*
1. **Parser:** `save_page_sections`' HTML fallback `saveSectionsExtractFromHTML` extracts only `<section>…</section>` blocks, but `tool-recreation-handler`'s `recreate_tool` prompt emits the tool as `<div class="tool-page">…</div>` (no `<section>`), and `save_sections` sets no `sections_metadata_field`. Zero matches → zero `page_components` → the rerender's `getPageSections` returns empty → page skipped, no git commit. (Content pages use `<section>`, so they deployed — hence the contrast.)
2. **Flip churn:** `upsertPage`'s `ON CONFLICT` flipped `deployed → needs_rebuild`, so any tool deployed before `sync_pages` ran was flipped and reprocessed by `page-build-handler`. This was a pre-design stand-in for drift detection that was never completed (the build-time `built_from_plan_version` stamp was deferred per `HANDOFF_2026-05-07` #5).

*Fix shipped (three files, all deployed):*
- `save_page_sections_action.go` — parser fallback: when zero `<section>` blocks match but HTML is non-empty, store the whole fragment as one section (guarded against full documents). Reuses the existing insert path; stops silent content loss.
- `v3_site_actions.go` `UpdatePageStatusAction` — deploy-time `built_from_plan_version` stamp from the site's current plan (the deferred design item).
- `site_db_actions.go` `upsertPage` — COALESCE → fill-if-null (never overwrite a real build version); removed the `deployed → needs_rebuild` flip. Drift now via the reconciler's `decideEmit`.

*Verification (2026-06-03 adoption, correlation `e9609749`, confirmed after full drain):* **Closed end to end.** All five games are deployed with committed files (`games/auto-battler|economy-simulator|jelly-invaders|p2p-networking|pathfinding/index.html`); tools deploy; the tested game (jelly-invaders) runs; the tested tools run; and `tools/index.html` lists the tools with working "Open tool" links pointing at the real tool pages. The interim partial read (4 tools, rest `triaged`) was the dispatch-throughput throttle (Family J), not an A1 fault — once drained, every tool/game landed. The three-file fix (parser fallback + deploy-time stamp + flip removal) is confirmed in production.

*Remaining:* none for tools/games. The fixes apply to *new* builds; pre-existing rows are not retroactively deployed.

**A4. Root `index.html` (homepage) deployed in DB but absent from the repo. [NEW, CONFIRMED 2026-06-03 — Group 2 focus]**
The homepage row exists and looks shipped — `name=index`, `page_type=landing`, `url=/index.html`, `build_status=deployed`, `built_from_plan_version` set — yet there is **no `index.html` at the repo root** (`sites/gamesdesign.co.uk/` has `about.html`, `contact.html`, `assets/`, `blog/`, `games/`, `guides/`, `tools/`, but no `index.html`). Every nav "Home" link targets `/index.html`, so it 404s. This is a **DB-deployed-but-file-absent mismatch** (a specific, confirmed instance of A2), not a planner/unbuilt gap.
*Cause PINNED (2026-06-04, reads in):* not a planner gap and not a git race. The plan gave the homepage **6 sections** (`hero, system-stats, tool-list, guide-list, game-list, call-to-action`), but its build work item `needs_page:index` (`needs_content_page`) is `complete` with error **"Auto-completed: work verified done despite lost response."** The content handler's response was **lost** (pod death / dead in-process timeout goroutine — same family as the "Claim timed out — handler pod likely died" entries), and the recovery path **auto-completed the item without verifying the artifact** → `page_components` never written (`n_rendered=0`) → falsely `complete`, so nothing re-queues it → a later rerender ran on the empty page and a deploy path marked it `deployed`+`stamped` despite 0 components → reconciler skips it → permanently fileless. The index row is **current, not stale** (the 06-03 `needs_page:index` still existing rules out a teardown since; `deployed_at`/rerender are 06-04).
*Systemic:* "auto-complete on lost response" falsely completes any page whose handler response is lost — likely behind some other defects too. Same handler-death problem as the dispatch claim-timeouts, but a more dangerous recovery branch (declares success vs re-queuing).
*Fix (located 2026-06-04 — the auto-complete is a SQL `pre_query`, not Go):* **(A)** the completion logic lives in the **`claimed-item-timeout` scheduled task's `pre_query`**, and `migration_claimed_item_timeout_evidence_check.sql` (already authored) is Option A — it gates completion on *page-specific* deploy evidence (`deployed_at > claimed_at`), replacing the old loose site-level `updated_at` check that false-completed the homepage. Its `reset` CTE is also the stale-claim reset (= Lever C). The homepage was false-completed under the OLD evidence on 06-03, so **the migration was not applied then — verify and apply it.** Optional hardening: credit `needs_content_page` by `page_components` existence (the artifact it produces) rather than `deployed_at` (a later item's output). **(B)** still required — the migration *trusts* `build_status='deployed'`, which this very page proves can be `true` with 0 components; guard `UpdatePageStatusAction`'s deployed branch (never mark a 0-component page `deployed`) so the flag is trustworthy and the 07:04 mislabel can't recur. Needs the **current** `v3_site_actions.go` (the uploaded copy is pre-A1, missing the stamp). *Immediate unstick (DONE 2026-06-04):* `UPDATE pages SET build_status='needs_rebuild', built_from_plan_version=NULL WHERE … name='index'` → UPDATE 1 (recurs if a response is lost again until A is applied).

**A2. DB-rows-vs-deployed-files mismatch generally. [NEW]**
26 `pages` rows but far fewer deployed HTML files. A1 is the bulk of it (all tool/game children), but the count should be reconciled fully: enumerate `pages` where `build_status='deployed'` vs actual files in the repo, to see whether anything beyond tools/games is also missing.
*Tentative area:* same as A1; confirm by diffing DB deployed-set against repo file list.

**A3. Stale files from prior runs may persist in the repo. [HOUSEKEEPING]**
Repo shows commits "4–5 days ago" alongside this run's. Earlier runs deployed flat `tools-index.html` / `games-index.html` at the repo root; those may still be present alongside the new nested hubs unless overwritten/cleaned. Not a code fix — a repo cleanup — but it muddies visual verification.
*Action:* `ls sites/gamesdesign.co.uk/*.html` and remove stale flat hub files if present.

---

## Family B — Link resolution / silent fallbacks

These all share a shape: a component emits a link to a page that does not exist, or an empty href, rather than resolving to a real target or degrading gracefully. They likely do **not** all share one fix — note the sub-distinction.

**B1. Hero and bottom-CTA links point at non-existent `/contact.html` and `/services.html`. [PARKED #2]**
On every hub and the homepage, hero buttons and the bottom CTA link to `/contact.html` and `/services.html`. Neither page exists in `pages`. Labels also disagree from targets ("Browse the Tools" → `/contact.html`; "Read the Guides" → `/services.html`).
*Tentative area:* hero/cta component contract falling back to a default page reference when the intended target is unresolved (silent-fallback). Section-component renderer.

**B2. Header "Get Started" CTA → `/contact.html`. [NEW]**
Same fallback shape but originates in the **header site-component**, not a page section. Architecturally a different surface from B1, so likely a separate fix even though the symptom matches.

**B3. Footer legal links → non-existent `/terms.html` (and `/contact.html`). [NEW]**
Footer "Terms of Service" → `/terms.html`, which is not in `pages`. `privacy` exists; `terms` and `contact` do not. Footer component emitting fixed legal links without checking the realised page set.

**B4. "Browse All Tools" bottom CTA has empty href (`href=""`). [NEW]**
Different mechanism from B1: not a fallback *to a wrong page* but a missing value resolving *to empty string*. tool-list component's bottom CTA target is unset.

**B5. Game-card "Play Now" CTAs have empty href (`href=""`). [NEW]**
Same as B4 but per-card on the game-list. Probably the same root as B4 (per-item link target field unresolved).

**B6. Broad audit needed: every deployed href vs realised page set. [NEW]**
Across hero/cta/header/footer/legal the site references `/contact.html`, `/services.html`, `/terms.html` — none realised. Need a grep of all deployed files for hrefs, cross-referenced against existing pages AND against the `gamedesign.uk` source crawl, to separate "genuine source page dropped from the plan" from "component-hallucinated reference."

---

## Family C — List-component content (tool-list / game-list)

**C1. Tool-list and game-list card descriptions are empty (`<p class="tl-card-desc"></p>`). [NEW]**
Card structure renders (icon, title, link) but the per-item description string is blank on every card, every list.
*Tentative area:* a field-name mismatch on the list-component contract — the per-item description the planner/section-data emits is not what the renderer reads (or is never produced). Distinct from empty *headings* (Family D) and empty *images* (C4).

**C2. Game-list is populated with fabricated games, not the adopted ones. [PARKED, now sharper]**
The homepage game-list shows invented titles ("The 63% Drop Rate Problem", "PRD vs True Random", "The Pity Timer Lab"…); `/games/index.html` shows a *different* invented set ("Drop Rate Lab", "EHP Sandbox", "Pity Timer Testbed"…). Neither matches the realised `game-jelly-invaders`, `game-auto-battler`, `game-economy-simulator`, `game-pathfinding`, `game-p2p-networking`. So the game-list LLM-fabricates items rather than querying realised game pages, AND each instance fabricates a *different* set so they don't even agree with each other.
*Tentative area:* game-list section generation is content-LLM-driven instead of querying `pages WHERE page_type='game'`.

**C3. Tool-list queries real pages but game-list fabricates — inconsistent behaviour between two sibling components. [NEW]**
Tool-list card titles are the real adopted tools (`TTK Calculator`, `Drop Rate Simulator`, …) — so tool-list *does* query realised pages. Game-list does not (C2). Two components that should behave identically diverge. Worth understanding why one queries and the other invents — they may have different generation paths.

**C4. Card images are blank (`<img src="" alt="…">`). [NEW]**
Every game-list card has an empty `src`. `needs_imagery` items completed (8) yet the card image slots are unfilled.
*Tentative area:* missing link between the imagery pipeline output and the list-component's per-card image field — or the schema asks for images the resolver never supplies. A *different* field from C1 (text vs image asset).

**C5. Game-list filter buttons are hard-coded to genres the data doesn't have. [NEW]**
Fixed pills "All Games / Action / RPG / Strategy / Sports" — none of the adopted games are "sports", and the schema exposes no genre. Decorative, non-functional UI that overstates the page.

**C6. "Load More Games" button is non-functional. [NEW]**
Click handler only disables the button; there is no pagination or further data. Same shell-UI-overstating-capability family as C5; same component, likely same fix.

---

## Family D — Section-data gaps

**D1. Guides hub (`/guides/index.html`) has no list of guides. [PARKED #1]**
Deployed with hero + nothing else — no `guide-list` section. The hub is correctly typed `section-index` but renders no listing of its child guides.
*Tentative area:* either the plan's `guides-index.sections` omits `guide-list`, or the `guide-list` section-data generation hit `needs_human_review` (see D3). Candidate structural fix discussed earlier: a `section-index` hub deterministically gets its matching list component populated from realised children.

**D2. Empty feature headings (`<h3></h3>`) on the homepage features section. [PARKED #1]**
Feature cards render descriptive paragraphs but with empty `<h3>` headings. Section-data/list gap on the features component.

**D3. Two `needs_section_data` items stuck at `needs_human_review`. [PARKED]**
From this run, two section-data work items did not complete. Likely the cause of D1 and possibly C1, but the specific items should be read before assuming the mapping — earlier these were the `guide-list` and an about-page content-block.

---

## Family E — Content quality

**E1. Hero H1 reused across all three hubs. [NEW]**
Homepage, tools hub, and games hub all use H1 "Stop Guessing. Start Calculating." with only the subheadline differing. Makes the three hubs read as duplicates rather than distinct pages.

**E2. Page titles carry the source site's verbatim `<title>` including "- GameDesign.uk". [NEW]**
Card titles show "Drop Rate Simulator - GameDesign.uk" etc. — the adopted source `<title>` used as the display name rather than a clean label. Cosmetic but pervasive (and note: it preserves the *source* brand "GameDesign.uk", not the destination "GamesDesign.co.uk").

**E3. Meta description empty on every page (`<meta name="description" content="">`). [NEW]**
`pages.meta_description` and `site_plan_pages.meta_description` exist; the renderer emits empty `content=""`. SEO/preview gap.
*Tentative area:* field populated by planner but dropped on the way to render, or renderer reading the wrong field.

**E4. Footer contact fields empty (`<a href="mailto:"></a>`, empty address). [NEW]**
Footer "Contact" column renders structure with empty mailto and no details — no graceful no-data path. Source likely didn't expose contact info; component renders empties instead of hiding the block.

**E5. Footer brand tagline empty (`<p></p>`). [NEW]**
Same no-data-path family as E4 — footer brand renders an empty paragraph.

**E6. Hubs share generic boilerplate hero rather than section-tailored content. [NEW, possibly design-chat]**
The hub heroes are near-identical boilerplate (same H1, same `/contact.html`/`/services.html` CTAs). Content-quality pattern from the page-content-writer rather than a hard bug.

---

## Family F — Guides faithfulness / duplication

**F1. Duplicate guide pages: `guide-rng-design` and `rng-design` coexist. [PARKED]**
`pages` holds both the adoption form (`guide-rng-design`, `/blog/guide-rng-design.html`) and the de-prefixed plan form (`rng-design`, `/blog/rng-design.html`). The repo deployed the `/blog/guide-*.html` files (adoption form). This is the open guides-faithfulness question.
*Decision still open:* accept de-prefixed `rng-design`, or make guides faithful — and note the earlier finding that typing guides as `guide` is the WRONG fix here, because source guides live flat under `/blog/`, not nested under `/guides/`. A faithful fix would preserve `guide-rng-design` at `/blog/guide-rng-design.html` and needs `page_canonical.go` to do correctly. Also requires reconciling away the duplicate row.

---

## Family G — Design fidelity (likely design-chat territory)

**G1. Deployed visual personality looks generic vs the source. [NEW, design-chat]**
Teal→black gradient header, plain black body, dark-blue/outlined buttons — does not strongly resemble the source `gamedesign.uk` palette seen earlier. The design-fingerprint extraction / design-intent generation produced something off-source. Flagged for completeness; likely handled in the separate design chat.

---

## Family H — Efficiency / hygiene

**H1. List components ship full inline `<style>` blocks per page. [NEW]**
Every page using tool-list/game-list re-includes the entire component stylesheet inline (~50+ lines each). Pages are heavier and any CSS change must propagate across every page. Should live once in a referenced CSS file.

---

## Family I — Open unknowns (need a read before they can be catalogued properly)

**I1. Do the recreated tools contain working interactive widgets?**
Moot for now given A1 (no tool pages deployed at all), but once A1 is fixed: confirm `needs_tool_recreation` output is an actual interactive tool, not an empty content shell. The tools are the product.

**I2. Does the `tool-game-*` double-prefix recur?**
Prior run produced double-prefixed names (`tool-game-auto-battler`). This run's `pages` does not obviously show it, but it wasn't explicitly checked. Re-confirm it's gone post-fix.

**I3. Imagery-to-card linkage.**
`needs_imagery` completed (8 items) but card images are blank (C4) and the hero image deploys. Understand which imagery targets resolved (hero, icons) vs which didn't (cards) — may be one bug or several.

---

## Family J — Dispatch throughput / pipeline cadence [SEPARATE THREAD — to be investigated]

**J1. Multi-tool/game sites drain extremely slowly (hours), and can appear stalled. [NEW, CONFIRMED 2026-06-03, not an A1 bug]**
Surfaced while verifying A1: on the `e9609749` run, 11 `needs_tool_recreation` items took hours to partially process — claims spaced 20–67 min apart (14:41, 15:04, 16:11, 16:58 …), with long idle gaps where nothing was running. Root mechanism is documented in `FOCUS_dispatch_diagnostic_4` Q3: the dispatcher is **one-site-per-tick** (selection `LIMIT 1`, spawned per scheduler tick via `build-pipeline-trigger`, processes ~5 items then exits) and **NOT-EXISTS-blocked** (a `NOT EXISTS` clause excludes a site *entirely* while any of its items is `status='claimed'` — an absolute blocker with no fall-through, line 276). So tools dispatch serially within a site, each held for its full ~5-min Opus build; and a dead handler that leaves an item `claimed` with a stale `claimed_at` freezes the whole site until a reaper resets it (the 47–67 min gaps came from two "Claim timed out — handler pod likely died" items). The scheduler cadence is `scheduled_tasks`/kafka-scheduler driven (no `build-pipeline-trigger` k8s cronjob), so between ticks a site with eligible `triaged` work can sit idle.
*Impact:* affects every site with many tool/game pages; makes adoption end-to-end take hours and look stalled even when correct. Not blocking A1 correctness (the per-tool build/deploy works), but it gates how quickly any multi-tool site reaches full deployment.
*To investigate in the separate thread:* whether one-site-per-tick + NOT-EXISTS should be relaxed (e.g. bounded per-site concurrency, or per-item rather than per-site exclusion), the claim-timeout/reaper window (shorten it so dead handlers don't freeze a site for long), and the scheduler tick cadence for `build-pipeline-trigger`. Standard manual unstick for now: re-fire `build-pipeline-trigger` / trigger a `build-dispatch-loop` for the site, or reset stale `claimed` items. Reference: `FOCUS_dispatch_diagnostic_4`, `105_dispatch-pipeline-failures-report_v4`, `010_scheduler_and_tasks`.

## Suggested triage order (tentative — for discussion, not committed)

1. **A1** — tools/games pages not deploying. Highest impact: the site's actual product is absent. Trace the child-page deploy/rerender path + one tool-recreation output.
2. **Family B (silent fallbacks)** — most pervasive user-visible breakage; links lead nowhere. Likely 2–3 distinct fixes (B1 section-cta, B2 header, B4/B5 empty-href), so triage within the family after reading the resolver.
3. **D1 (guides hub list) + C2 (game-list fabrication)** — both are "section-index hub / list component not reflecting realised children", structurally adjacent and the natural follow-on to the section-index work just completed.
4. **C1/C4 (card desc + images)** — list-component contract fields.
5. **E3 (meta), E1/E6 (hero reuse)** — content/SEO quality.
6. **F1 (guides duplication)** — needs `page_canonical.go`; product decision still open.
7. **H1, A3, E2** — hygiene/cosmetic.
8. **G1** — defer to design chat.

Each item gets the same discipline: read the responsible action/source, pin the cause, confirm against the data, then propose a fix — no fixing from the tentative causes above.
