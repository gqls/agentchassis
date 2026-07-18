# RESULTS — robot-hands.com site fixes (R1–R6), 2026-07-17/18

Continuation of `HANDOFF_2026-07-17_robot_hands_site_fixes.md`. This doc
records what was actually found (several root causes differ from the handoff's
working hypotheses), what was changed, and what remains.

## Corrections to the handoff's evidence

- **The blue header predates the 2026-07-16 regen by a week.** gqls/sites
  bisect: blue gradient entered index.html between af0ead8da1
  (2026-07-08T22:44) and 78532b8c63 (2026-07-09T09:13). The 07-16
  needs_rerender wave *spread* it to all pages; it did not create it.
- **component_versions holds NO snapshots** for header/footer/head — the
  "restore from component_versions" route did not exist. Restoration was done
  by construction (var-based components), not from snapshots.
- **hardcoded_section_colors never rewrote anything.** Both "completed" items
  actually failed inside their workflow (`WORKFLOW_INVALID: step
  'fix_text_colors' with action 'fix_forced_text_colors' requires a topic`)
  and were still marked complete. The white-look damage came from the palette,
  not from this handler. Two open bugs worth their own attention:
  (a) the fix_forced_text_colors workflow config is invalid;
  (b) a failed workflow response can mark a work item complete.
- **The "blank" MatchMatrix page is not blank** — 6.3KB of visible text. It
  *looked* blank: card-heavy page with `--color-card-bg: #ffffff` (white
  cards, light text) plus `color: var(--color-primary)` misused as a text
  colour (#1A1F2E on #0F1218). Palette fix + CSS re-render addresses it.

## Root causes (R1)

1. **B7 swapped only `css_themes.layout_id`.** Three palette copies stayed
   brochure-blue/light: `palettes.colours`, `style_collections.color_palette`
   (read by RenderSiteComponentsAction → fills `{{.primary_color}}` etc.),
   `css_themes.color_palette`. Every component regeneration re-baked blue.
2. **site_components pointed at DEACTIVATED colour-baking components**
   (`header-bold-gradient`, `footer-4-column`) — exactly what the three
   FAILED 2026-05-13 deactivated_component items said.
   `renderAndStoreSiteComponent` ignores `is_active`
   (render_site_components_action.go:489).
3. **webdesign-agent re-rolls core colours every run.** Its analyze_design
   LLM step invents a `color_scheme` per run and render_css_from_spec's merge
   lets spec win on core slots. The prompt only renders the STRUCTURED
   `design_intent.palette` block — robot-hands had only free-text
   `colour_mood`, so the LLM always fell to "no intent → invent". Four CSS
   rewrites on 2026-07-17 alone; the 20:31 run shipped a light background
   (#F4F5F7) onto the dark site.
4. **generic_theme misfires fleet-wide** (the churn's trigger): the check
   demands `site_specs aspect='webdesign'`, which NO code path has ever
   written (0 rows fleet-wide). webdesign-agent actually stores its spec in
   `sites.content_data`. Every themed site re-fires every pass.

## Changes applied (all DB-live unless noted)

| # | Change | Artifact |
|---|--------|----------|
| R1 | New var()-based `header-theme-chrome` / `footer-theme-chrome` components (markup the dark layout styles natively; zero literal colours); robot-hands header/footer slots repointed; all three palette copies rewritten to the approved dark set; 2026-05-13 deactivated_component items closed | `SQL_2026-07-17_r1_dark_theme_restore.sql` (commit 17b4adcc7) |
| R1b | `design_intent` superseded with structured palette+typography pins (reference_values = approved dark set, hard "never light" guidance) | `SQL_2026-07-17_r1b_design_intent_palette_pin.sql` (same commit) |
| R1c | generic_theme check also accepts `sites.content_data ? 'color_scheme'` — **Go, inert until next chassis image build+roll** | `check_generic_theme.go` (commit 3437f2212) |
| R3 | Learning Center canonical = `/learning-center-hub.html` in primary AND utility nav; grid page kept active but demoted (avoids phantom-link churn from baked in-body links); `learning-center/index.html` archived; plan nav flags in lockstep; 2 failed index-rebuild items closed | `SQL_2026-07-17_r3_learning_center_ia.sql` (7be3720b9) |
| R5 | Hub's `show_load_more` → false AND `content-listing` schema fallback → false (dead control must be opt-in). Explicit trues left on dartsonline/idea.uk (other owners) | `SQL_2026-07-17_r5_dead_load_more.sql` (c10ec34e7) |
| R6 | 6 dead article rows archived (3 scaffolds + 3 duplicate slugs); their 12 card/hero assets superseded; 5 "re-render after asset landed" review items closed. Listing eligibility itself was already fixed by imagery F2.1 (`listedOnly`, commit 4e35c8064 — verified in the running pod) | `SQL_2026-07-17_r6_retire_dead_article_rows.sql` (898ff3494) |

Re-renders executed: site components force-re-rendered 2026-07-17 20:23
(header 2547B, footer 3335B, both var-based, zero `#3b82f6`); CSS re-rendered
twice — the post-pin run committed **f635b1e49** with background #0F1218,
accent #E8500A, text #E2E8F0, IBM Plex Sans body. A 37-page re-render batch
was created 20:23 and promoted to priority 20 on 2026-07-18 (it sat behind the
stale priority-35 churn backlog).

## Verification state (as of writing)

- DB: components/palette/CSS verified via SQL + gqls/sites blob reads. ✔
- Live styles.css: dark, pinned values. ✔ (f635b1e49)
- Live pages: still carry yesterday's baked blue header until the 37-page
  batch drains and deploys — batch monitor armed. ⏳
- R2 (yellow-on-white) and R4 "blank" MatchMatrix: expected fixed by
  card_bg/palette + page re-renders; VERIFY VISUALLY after the drain.
  Watch: `forced-text-color-fixer` must not "fix" the amber once dark
  returns (its workflow is currently broken — see above — so low risk).

## Remaining / not done here

- **R4 tools**: sweep done — 3 of 5 tool pages live; `tool-matchmatrix` and
  `tool-robot-payload-budget-calculator` 404 (the two `incomplete_page_group`
  review items). Body CTAs on index/matchmatrix pages point at the 404
  (`/tools/matchmatrix/index.html`) — rewritten there by a recent link pass
  that trusts *planned* pages. Building the two tools = experience_loop
  workstream's class (`../experience_loop/HANDOFF_2026-07-17_experience_loop_start.md`);
  coordinate, don't duplicate.
- **Chassis image**: the generic_theme fix needs a `make build-agent-chassis`
  (builds committed HEAD) + roll; batch it with the next planned deploy —
  do NOT roll while the site's page batch is mid-drain.
- **fix_forced_text_colors workflow invalid** + **failed-workflow-marked-
  complete**: loop-integrity class, flagged for the fixloop/empty-sections
  workstreams.
- **improvement-sweep scheduled task is disabled** (since 2026-07-17 13:18).
  Deliberate? Another session's call — left as found. Dispatch runs via
  build-pipeline-trigger (30s) which processes ONE item per site at a time;
  57-item backlogs take many hours.
- Learning-center category grid: folding its (phantom-CTA) cards into the hub
  or fixing their targets is content work for the loop once tools exist.
