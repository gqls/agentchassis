# FOCUS — Site Planner Ignores Adopted State (Duplicate/Spurious Pages)

Date: 2026-05-19
Status: diagnosed; fix is doc 029 Phase 1 (planner reads realised state). Short-term mitigations available.

## One-line summary

After a clean adoption produces 27 faithful pages, `build-site-planner` independently generates a 9-page generic site skeleton that ignores the adopted pages. `reconcile_site_plan` then faithfully creates work items for the plan's pages, adding spurious/duplicate pages on top of the adopted ones.

## How this surfaced

gamesdesign.co.uk re-adoption (correlation a3efeaaf, 2026-05-19):
- 13:57–14:00 — adoption ran. analyze_site emitted 20 clean pages, `apply_adoption_plan` wrote them. Faithful to gamedesign.uk. (This is the part we fixed via the max_tokens and error_step work.)
- 15:39:03 — `build-site-planner` wrote a fresh site_plan (9 pages) via `write_site_plan`.
- 15:39:11 — `reconcile_site_plan` read that plan and emitted `needs_page` work items; page-build-handler created the pages.

Result: 27 page rows including duplicates (`games` + `games-index`, `tools` + `tools-index`), renamed tool dups, and a `guide-post` placeholder.

## Corrected diagnosis (an earlier hypothesis was wrong)

Initial hypothesis: the plan was *stale* (pre-canonicalisation), and the reconciler was processing old data. **Wrong.** The plan is fresh — `created_at 2026-05-19 15:39:03`, `source_agent build-site-planner`, `created_by write_site_plan`. `WriteSitePlanAction` (which calls `ValidateRoles` + `CanonicalisePage`) produced it. The canonicaliser is wired in and ran.

The actual problem: the plan the LLM produced is a generic skeleton, not a reflection of the adopted site.

## What the plan contains (the evidence)

`site_plan_pages` for the current plan (9 rows):

| name | role | url | issue |
|---|---|---|---|
| index | landing | /index.html | ok |
| tools | content | /tools.html | flat content page, not converged to tools-index |
| games | content | /games.html | flat content page, not converged to games-index |
| guides-index | section-index | /guides/index.html | correct |
| about | content | /about.html | net-new (acceptable) |
| contact | content | /contact.html | ok |
| post | blog-post | /blog/post.html | placeholder; canonicalises to guide-post |
| tool-lanchester-combat-calculator | tool | /tools/.../index.html | renamed dup of adopted tool-lanchester-sim |
| tool-loot-probability-calculator | tool | /tools/.../index.html | renamed dup of adopted tool-drop-rate-simulator |

Compare to the 20 pages adoption emitted (game-jelly-invaders, game-pathfinding, tool-ttk-calculator, guide-skinner-box, guide-rng-design, etc.). The plan overlaps only on `index` and `guides-index`. Everything else is parallel invention.

## Two mechanisms, both confirmed

### Mechanism 1 — planner doesn't read adopted/realised state

`build-site-planner` generated a "what a game-design site should have" skeleton from the site's identity/archetype, not from the pages adoption already wrote. This is the exact root cause doc 029 documents:

> two surfaces (adoption and the site-planner) both write `pages` rows and queue work items, and they don't share a common identity space, don't read each other's output, and neither one knows about the existing realised state of the site.

The planner here had no idea the site already had 8 tools, 5 games, 6 guides. It planned 2 arbitrary tools, 0 games, 0 guides.

### Mechanism 2 — canonicaliser can't converge a childless plan

`CanonicalisePage` + `ValidateRoles` are wired into `WriteSitePlanAction` and ran. `guides-index` converged correctly. But `games`/`tools` stayed as flat `content` pages because:

- `ValidateRoles` rule that promotes a page to `section_index` fires only when **another page in the set declares it as ParentSection**.
- This plan has **no child pages** — every `parent_section` is empty.
- So the only `section_index` promotions that happen are ones the LLM labels correctly itself. The LLM labeled `guides` as section-index but `games`/`tools` as content; ValidateRoles had no structural signal to correct the latter.

The convergence machinery works as designed; it's defeated by a plan with no children. Mechanism 1 (childless generic plan) is the upstream cause of Mechanism 2's visible symptom.

## Why this matters for the faithful-first-pass goal

User's stated priority: extra pages/tools are tolerable if they work and adopted tools work, but the ideal is a faithful first pass — only changes being components to fit our code and making tools/games work (rewriting if needed).

The planner actively works against that: it generates a parallel generic structure regardless of what was adopted. As long as `build-site-planner` runs on the post-adoption build cascade without reading adopted state, every adoption will be followed by this generic-skeleton contamination.

## The documented fix (doc 029 / 030)

Doc 029 ("Site Plan as Declarative Artefact, Reconciler, and LLM Tiering") specs the fix:

- The site plan becomes a declarative description of desired state.
- The planner **stops emitting work items**; it only describes desired pages.
- A deterministic Go reconciler walks desired-vs-realised and emits work for the diff — "can't produce duplicates by construction."
- Phase 0 (canonicalisation) — landed. The CanonicalisePage/ValidateRoles functions.
- Phase 1 (planner reads realised state and converges; stops being a queue producer) — the architectural shift. **This is what's still needed.**
- Phase 2 — discoverers read the plan to judge fitness.

The decisive missing piece is Phase 1: the planner must be fed the adopted/realised page inventory and instructed to converge on it (extend, not re-propose), rather than planning from identity alone.

## Sub-issues to fix (in priority order)

1. **Planner ignores adopted pages (Mechanism 1) — the core issue.** `build-site-planner`'s `plan_site` step needs the adopted page inventory as input, with instructions to reuse exact existing slugs and only add genuinely new pages. This is Phase 1's essence. Until then, every adoption gets a generic overlay.

2. **Renamed tool duplicates.** Even with adopted inventory as input, the LLM may "improve" names (`lanchester-sim` → `lanchester-combat-calculator`). The plan_site prompt needs an explicit "use the exact slug from the adopted page; never rename existing pages" rule.

3. **`post` placeholder leakage.** The plan contains a `post`/`/blog/post.html` page that becomes `guide-post`. Almost certainly a JSON example in the plan_site prompt being copied verbatim. Remove or clearly mark example placeholders in the prompt.

4. **Nav dedup guard (doc 029 B-029-1).** Independent of the above: the nav builder should dedup on canonical name before emitting `<li>`, so even if duplicate page rows exist, the rendered header/footer can't show `Games` twice. This is the actual user-visible harm; worth a guard regardless of when Phase 1 lands.

## Short-term vs proper fix, mapped to user priority

**Tolerable-now (extra pages OK if they work):**
- Add the nav dedup guard (#4) so duplication can't reach rendered HTML.
- Let the extra pages build. Confirm the adopted tools/games actually work (they're the faithful ones).
- The duplicate `games`/`tools` flat pages and the renamed-tool dups are wasteful but not breaking, provided nav is deduped.

**Proper fix (the faithful-first-pass ideal):**
- Doc 029 Phase 1 — planner reads realised state, converges, stops emitting work items.
- This eliminates Mechanisms 1 and 2 at the source. After it lands, a planner run on an adopted site would produce a plan that matches the adopted pages plus only genuinely additive ones.

## Verification queries for the next step

```sql
-- What input did build-site-planner's plan_site step actually receive?
-- (widen the window — the run was 2026-05-19 15:39)
SELECT created_at, agent_type, step_name,
       LEFT(prompt_rendered, 1500) AS prompt_head
FROM llm_call_log
WHERE agent_type = 'build-site-planner'
  AND step_name = 'plan_site'
ORDER BY created_at DESC
LIMIT 2;
```

Look in `prompt_head` for whether the adopted page list appears. If it does NOT, Mechanism 1 is confirmed at the prompt level — the planner literally isn't told what was adopted.

```sql
-- Does the build cascade trigger build-site-planner with adopted pages as input?
-- Inspect the agent that calls it (site-work-orchestrator / build cascade).
SELECT type, default_config->'workflow'->'steps' AS steps
FROM agent_definitions
WHERE type IN ('build-site-planner', 'site-work-orchestrator', 'build-dispatch-loop')
  AND is_active = true;
```

## References

- `029_site_plan_and_reconciler(1).md` — declarative plan + reconciler design; B-029-1 duplicate nav; planner-divergence root cause
- `030_phase1_plan_and_reconciler*.md` — Phase 1 detail (planner stops emitting work items)
- `FUTURE_adoption_source_destination_separation.md` — adoption variants; `clone` (default) = faithful copy intent
- `HANDOFF_2026-04-23_dispatch_reliability_and_008_validated(1).md` Priority 5 — "Fix planner imagery + page-dedup respect for adoption" (this finding is that item, now diagnosed)
- Code: `WriteSitePlanAction` (chassis ~line 86092, calls ValidateRoles + CanonicalisePage), `ReconcileSitePlanAction` (~74158, reads plan, emits work items — innocent), `ApplyAdoptionPlanAction` (~52933, adoption path)
- Functions: `datahelpers.ValidateRoles`, `datahelpers.CanonicalisePage` (Phase 0)
- This conversation, 2026-05-19
