# HANDOFF — Design/Imagery Triggers Deployed; Adoption Page_Type Re-Type Diagnosed

Date: 2026-05-26
Status: Checkpoint. Multiple structural fixes deployed this session and from other
chats. Verification of the post-deploy state — specifically whether the adopted
`game-*` pages now round-trip rather than duplicating as `tool-game-*` — is the
first task next session.

## Scope and context

The session opened from the robot-hands.com rebuild handoff (2026-05-20) and was
quickly reconciled against the 2026-05-21 worklist (FAQ prevention + news), which
had landed in the interim. The user then steered away from per-site work to fix
the general flow first. Three structural gaps surfaced and stacked on each other.
robot-hands and then gamesdesign.co.uk served as worked examples; conclusions are
about the flow, not the sites.

## What deployed this session

`emit_design_items` and `emit_imagery_items`, with the shared `imageryplan`
package, wired into `build-site-planner` as plan-time workflow steps. The
deployed step order in v1.0.1047 is
`read_specs → ensure_site → load_existing_pages → load_components → load_styles → plan_site → validate_plan → write_site_plan → sync_pages → populate_nav → reconcile_site_plan → emit_design → emit_imagery → complete`,
with each step's `output_field` (`design_items`, `imagery_items`) listed in
`complete.output_fields` and `config.site_id = "input_data.site_id"` matching each
action's `Required` input. This closes the long-standing "trigger gap" (composition
+ imagery never emitted at build time since the Phase-1 refactor moved the
terminal step away from `WriteBuildItemsAction`).

The `write_site_plan` workflow-step description was updated via `snapshot_agent`
+ `jsonb_set` on `agent_definitions.default_config` to include `site_plan_imagery`
and the imagery HITL-lock transfer (cosmetic, non-executable; `UPDATE 1`,
`snapshot_agent` returned `source_version=1`).

Additional code changes deployed in other chats and applied to production on
2026-05-26 evening. Their specific content is not in this session's context; the
candidate targets are the things this session diagnosed as open (page_type
vocabulary, write/sync canonicaliser divergence, nav dedup, gap C palette emission).
Treat their presence as unknown; the first verification queries below distinguish
which landed.

## What was verified end-to-end on the adoption path

gamesdesign.co.uk was adopted from gamedesign.uk and traversed the full cascade.
`needs_domain_research → needs_strategy → needs_briefing → needs_site_plan → build-site-planner → emit_design → emit_imagery` all completed. `needs_composition`
ran via `site-design-planner` and `install_site_composition` populated
`sites.style_collection_id` (NULL pre-deploy). `needs_design` completed via
`webdesign-agent`. Nine `needs_imagery` items appeared at the documented priority
bands (65 for index-hero, 70 for site-logo, 75/80 for other site and page-hero,
98 for clamped section-scope). Adoption does not bypass the planner — it routes
through it via the strategy→briefing→site_plan chain, as `007_adoption_pipeline_v4.md`
intended.

So Gap A is closed on both fresh-build and adoption paths.

## The diagnosis that landed (durable)

The 2026-05-26 gamesdesign adoption produced exactly five spurious duplicate
pages. The authoritative `pages` query at completion shows 27 rows with five
duplicated stems — `auto-battler`, `economy-simulator`, `jelly-invaders`,
`p2p-networking`, `pathfinding` — each present both as `game-*` (adopted,
`page_type=game`) and as `tool-game-*` (`page_type=tool`). Nothing else duplicated:
the six adopted calculators (`tool-*`, kept `tool`), the five guides (`guide-*`,
kept `blog-post`), and the indexes are all single rows.

The root cause is confirmed from the planner's `response_text`. The LLM emitted
the games faithfully by name but assigned them `page_type: "tool"`:

```json
{ "name": "game-auto-battler", "page_type": "tool", "sections": [], ... }
```

The deployed `plan_site` prompt's Canonical Page Types list is
`index | content | landing | entity-directory | entity-page | tool | blog-index | blog-post`
— there is no `game`. The planner cannot emit `page_type: game`, so every adopted
game is forced to `tool`. The §9 canonicalisation in the debugging guide then
drives the rename via the `tool` branch. The asymmetry between guides (de-prefixed
in the plan but kept `blog-post`, did **not** duplicate) and games (re-typed
`game`→`tool`, did duplicate) is the proof: a name change alone does not duplicate
a page; a `page_type` change does.

One thing the diagnosis left unread on purpose: the two downstream canonicalisers
disagree about the same `tool`-typed page — `WriteSitePlanAction` stored
`tool-auto-battler` in `site_plan_pages` while `sync_pages_to_db` realised
`tool-game-auto-battler` in `pages`. Both follow from `page_type: tool`, but the
strip-vs-no-strip split is a code question for `datahelpers/page_canonical.go`
plus the two action paths. Per the debugging discipline (§0 item 19), do not
prescribe a fix against the inferred mechanism — read the code first.

## Wrong turns owned, recorded in the guide

This session twice misdiagnosed the renames before the existing
`FOCUS_planner_ignores_adopted_state.md` was surfaced. First as a build-timing
race (false: every adopted page was created at 14:06, well before the planner
ran at 16:25; `existing_pages` was complete). Then as LLM non-compliance with the
prompt's "do NOT rename" rule (false: the `prompt_rendered LIKE` only proved the
slugs were in the planner's *input*; `response_text` later showed the LLM was
faithful on names). The actual cause was already documented. Both wrong turns
and the meta-causes (input ≠ output evidence; check the guide before generating
fresh hypotheses; design tests to falsify, not confirm) are logged as §0 item 19
in the debugging guide.

## The three stacked gaps as of this checkpoint

(A) Composition and imagery never triggered. CLOSED by this deploy on both
build and adoption paths. Confirmed via gamesdesign cascade.

(B) Planner drifts from adopted state — specifically, the `page_type` vocabulary
mismatch forces `game`→`tool` re-type, which drives the canonicalisation
rename + duplication. OPEN structurally; may have been addressed by the
other-chat fixes deployed on 2026-05-26 evening. Verify post-deploy with the
queries below.

(C) Planner emits colour decision as prose (`design_intent.colour_mood`) not as
structured `design_intent.palette.reference_values`. `site-design-planner`'s
palette cascade therefore misses the design_intent slot and falls to the
layout-seed/default base palette; the planned colours reach render only via the
webdesign-agent overlay, not the base composition. OPEN.

## Other open items

`WriteSitePlanAction` / `sync_pages_to_db` canonicaliser divergence — same
`tool`-typed page yields different names in `site_plan_pages` and `pages`. Read
`CanonicalisePage`, `ValidateRoles`, and the two call sites before designing the
fix.

Index-hub mistype — `games-index` and `tools-index` are `page_type=content` while
only `guides-index` is `blog-index`. They survived the 2026-05-26 run by
name-matching realised pages, but the type is wrong and is a latent §9 sub-case.

`sites.build_status` is vestigial — defaulted to `'pending'` at insert, never
advanced by any code path (every `UPDATE sites` writes other columns:
`content_data`, `style_collection_id`, `last_built_at`, `last_deployed_at`,
`default_components`, `status`). Real build/deploy state lives in
`last_built_at` / `last_deployed_at` / `last_reconciled_at` and in per-page /
per-component `build_status` (those do advance). Decide whether to maintain or
drop the column.

Nav dedup guard — `FOCUS_planner_ignores_adopted_state.md` sub-issue #4. With
duplicate page rows on gamesdesign, the rendered nav can show `Games` twice
unless the nav builder dedups on canonical name. Independent of the upstream fix;
worth landing as the user-visible-harm backstop.

`needs_section_data` human-review escalations — two or three items on gamesdesign
sitting in `needs_human_review` from `plan_sections` (a content-build escalation
when a section couldn't be resolved). Triage when convenient.

SECURITY — rotate `STABILITY_API_KEY` and `BANANA_API_KEY` (plaintext exposure
flagged in the imagery handoff). Ops-only action; not addressed in this session.

Per-site cleanup on gamesdesign — five live `tool-game-*` duplicate page rows
deployed. Cleanup is marking them inactive / soft-deleting plus re-rendering
nav. Separate from the structural fix; available whenever wanted.

The robot-hands.com rebuild thread from the 2026-05-20 handoff remains queued.
Its `design` (adoption green/cyan palette) and `structure` (2-page seed) aspects
are still present; `style_collection_id` is NULL. Whenever the page_type and
palette fixes are verified on a fresh adoption, robot-hands is the natural
re-adoption candidate to exercise the corrected flow.

## Where to resume

Run a fresh adoption (or replan an adopted site) on the freshly-deployed code
and use these three reads, in order. They distinguish which of the candidate
fixes from other chats landed, and where.

```sql
-- 1. Stem-grouped pages: did the page_type re-type still produce duplicates?
SELECT name, page_type, status,
       regexp_replace(name, '^(tool-game-|tool-|game-|guide-)', '') AS stem
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = '<test-domain>')
ORDER BY stem, name;
-- Zero stems appearing twice = duplication fixed. Five paired stems (game-* + tool-game-*)
-- = upstream re-type still happening.

-- 2. response_text page_type check: did the planner keep page_type=game,
--    or is it still emitting tool?
SELECT substring(response_text from '\{[^{}]*"name"[^{}]*"game-auto-battler"[^{}]*\}')
       AS game_entry
FROM llm_call_log
WHERE agent_type = 'build-site-planner' AND step_name = 'plan_site'
  AND created_at > '<deploy-time>'
ORDER BY created_at DESC LIMIT 1;
-- page_type:game = fix at the planner (vocabulary updated or LLM preserves).
-- page_type:tool with no duplicate in query 1 = fix downstream (deterministic
-- page-identity preservation overrides the planner's type for existing pages).

-- 3. Composition install on the new adoption
SELECT domain, style_collection_id IS NOT NULL AS has_collection,
       last_built_at, last_deployed_at
FROM sites
WHERE status = 'active'
ORDER BY created_at DESC LIMIT 10;
-- New builds should show has_collection = t.
```

Branching from those reads. If query 1 returns zero paired stems, gap (B) is
landed — read query 2 to confirm where the fix lives, then move to gap (C)
(palette structured emission) or to canonicaliser unification. If query 1 still
shows the five game pairs, the page_type re-type is still happening; the next
step is reading `CanonicalisePage`/`ValidateRoles` (in `datahelpers/page_canonical.go`)
plus `WriteSitePlanAction` and `sync_pages_to_db` to design the structural fix
against the actual code. Either way, query 3 is the cheap monitor on gap (A)'s
continued correctness across new builds.

If both queries 1 and 2 show the fix landed cleanly, the natural next thread is
gap (C): the planner does not emit `design_intent.palette.reference_values` in
the `plan_site` output, so the composition cascade's design_intent slot misses
and the base palette resolves to the layout-seed/default — the planned colours
reach render only via the webdesign overlay. Either add a structured
`palette.reference_values` block (core keys: `primary, secondary, accent,
background, surface, text, text_muted, border`) to the `plan_site` output schema,
or change `site-design-planner` to consume `colour_mood` directly.

## Documents landed this session

`/mnt/user-data/outputs/FOCUS_design_composition_flow_and_adoption_fidelity.md`
— new focus doc. Records the design-composition flow investigation, the three
stacked gaps with evidence references, and the (already-decided, partly-
implemented) adoption-fidelity model the gaps sit inside. Updated mid-session
when emit_design / emit_imagery deployed.

`/mnt/user-data/outputs/016_debugging_guide_v2_22.md` — updated debugging guide.
Two additions: a §9 addendum "Adoption slug-mangling re-confirmed (2026-05-26)"
extending the existing "Adoption faithfulness …" entry with the page_type
vocabulary root cause, the final-state authoritative count, and the
write/sync canonicaliser divergence flagged as code-read-needed; and §0 item 19
"A `LIKE` on `prompt_rendered` proves what the model was *told*, never what it
*did* — and a familiar-looking failure may already be diagnosed," recording
this session's misdiagnoses and the four meta-causes.

## Cross-references

`HANDOFF_2026-05-20_robot_hands_rebuild.md` — the rebuild thread this session
opened from. Steps 1–2 of its plan (clean stale specs, news enrichment) were
deferred when the work pivoted to general flow.

`HANDOFF_2026-05-21_faq_prevention_and_news.md` — prior FAQ prevention deploy
(v1.0.1029) and news `files_field` fix; both de-risked the trigger work in this
session.

`FOCUS_planner_ignores_adopted_state.md` — the 2026-05-19 prior diagnosis of the
same site. The §9 addendum extends it.

`FOCUS_adoption_faithfulness_via_locks.md`, `028_platform_mission_and_pipeline_direction.md`,
`FUTURE_adoption_source_destination_separation.md` — the adoption-fidelity model
(fidelity dial, variant axis, timed-lock enforcement, phasing).

`029_site_plan_and_reconciler_*.md`, `030_phase1_plan_and_reconciler_5_.md` —
the structural endgame (declarative plan; deterministic reconciler; planner
stops emitting work items). The doc 029 direction is what the page-identity
preservation fix ultimately tracks toward.

`027_design_and_site_planner_v2.md` — the composition resolver and its palette
cascade; relevant for gap (C).
