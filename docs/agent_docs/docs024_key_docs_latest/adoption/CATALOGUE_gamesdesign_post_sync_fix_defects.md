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

---

## Family A — Deployment gaps (highest user impact)

**A1. Tool and game *pages* did not deploy at all. [NEW, CONFIRMED, TOP PRIORITY]**
`pages` contains rows for every tool and game (e.g. `tool-ttk-calculator → /tools/ttk-calculator/index.html`, `game-jelly-invaders → /games/jelly-invaders/index.html`), and their `needs_page` / `needs_tool_recreation` work items show `complete`. But the repo's `/tools/` and `/games/` directories contain only `index.html` (the hub) — none of the child tool/game files are present. The interactive tools and games are the actual product of this site, and none of them reached the deployed output. This is the most important defect.
*Tentative area:* the gap is between "page row built/complete in DB" and "file committed to git". Could be the rerender/deploy path skipping nested child pages, the tool-recreation handler not producing a deployable artefact, or a git-path mismatch for nested `/tools/<slug>/index.html`. Needs: the deploy/rerender path for child pages, and one `needs_tool_recreation` item's output, read before concluding.

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
