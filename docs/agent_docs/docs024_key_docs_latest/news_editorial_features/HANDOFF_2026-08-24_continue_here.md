# HANDOFF — news editorial + editorial design uplift, 2026-08-24. START HERE.

Supersedes `HANDOFF_2026-08-21_continue_here.md` for a fresh session — but that
file is NOT dead: **its §3 recipe (proven three times) and §9 traps (every one
paid for) still govern and are pointed at, not restated.** Everything below is
measured 2026-08-24 unless marked otherwise.

This session ("news editorial", Fable 5) ran 08-22→08-24 and covered THREE
threads that now hand off together:

1. **news_editorial_features** — the feature pages (NEWS-020, 4 pages / 2 sites live).
2. **editorial_design_uplift** — how they look (companion lane, same session).
3. **features_open/035** — component hierarchy, written this session BY FABLE
   (the capacity block ended because the owner's session runs Fable; fifth
   attempt, no substitution).

Nothing from these lanes is pending a roll: all shipped work is docs, DB config
(live on apply) and a repo-local proof harness. **P1 will be this thread's first
platform code.**

---

## 1. The headline: 035 is WRITTEN, P0 is PROVEN, P1 is UNBLOCKED

- **`features_open/035_FEATURE_component_hierarchy.md`** is the design: children
  are `page_components` rows via the dormant `parent_instance_id` (1,903 rows /
  0 set as of 08-22); one LLM call per child against a shared `content_brief`;
  composition below the section grain; Go-side walk injecting pre-rendered
  slots; versions + contract-guarded design variants. Read it in full before
  touching anything — §6 hazards and §9 what-NOT-to-do are load-bearing.
- **P0 (local walk proof) PASSED 08-22** — harness at
  `editorial_design_uplift/harness/composewalk/` (own `go.mod`, outside the
  platform build). Eight checks green incl. flat-page byte-identity; the checks
  proven able to fail by two walk mutations, each caught by exactly the aimed
  check. Design refinement fed back into 035 D4.3: **the cycle guard is a
  COMPLETENESS assertion** (every row rendered exactly once), because with one
  parent pointer per row every cycle is unreachable and a path-set alone drops
  rows silently.
- **D6 was CORRECTED same day (read the strike-through in 035):** the version
  pin must be a NEW opt-in `pinned_component_version_id`, never
  `component_version_id` — RFC_046 (ruled 08-22) makes that column a provenance
  STAMP written by renders, so a pin read of it would freeze every instance at
  its last render. `stamp == pin` becomes the honoured-check. The 357 lane was
  told (their NOTES, 08-22 dated append).
- **P1 was deferred 08-22** (both integration files carried the 357/RFC_046
  lane's uncommitted WIP) and **is UNBLOCKED as of 08-24** — both files clean:
  `rerender_page_sections_action.go`, `component_library.go`.

**P1, the recommended next step (owner concurred in-session):** re-read the
seams FIRST — `RenderTemplate`'s final contract (it gained a
`RenderedTemplateSHA` output; the `2817f6661` AST tests enforce it as the ONE
executor spelling — the walk must go through it, never a new path),
`carryStoredSection`, and `deriveRenderMode` (:561/:639 — runs on INSERT and
UPDATE, so hand-seeded `composite` reverts). Then: walk in both render paths +
`deriveRenderMode` third value + `check_render_mode` routing arm + register
entry, ONE council-gated commit; recompose ONE live insights page; acceptance =
served page byte-equivalent, `pages.sections` diff EMPTY, then **rewrite one
prose child and prove every sibling row byte-identical**. Consumers to TELL
(07-29 §3): component-library lane, 283 instance-scope lane, 238/268 carry
lane, inline_guide_imagery lane. **Until P1 ships: no `composite` rows, no
`parent_instance_id` values, anywhere** (035 §9 rule 1).

## 2. What is live (delta since the 08-21 handoff)

- The four editorial pages/hub unchanged and re-verified healthy 08-23
  (stylesheet control first: 25,559 B; both features serve hero + full charts).
- **Contrast fix mig `496`** (design lane Phase B first item): robot-hands
  10→4 findings, dartsonline 8→1, VIZ-014 de-branding control run and passed.
- **Both robot-hands articles have content images** (items complete 08-22; the
  asset keys were not examined — find them before generating duplicates) and
  the electric-vs-pneumatic tool + guide pages have `content_hero_*` heroes
  (assets active 08-23).
- **dartsonline `card_darts_calendar_density` asset active** — the card-image
  derivation this lane approved ran exactly as specified: new asset, **0 writes
  to the article's rows** (watch-point verified 08-24).

## 3. Open work, precise state

| item | state |
|---|---|
| **P1** (035 read path) | UNBLOCKED, next. §1 above is the brief. |
| **`lock_blocked_change:darts-calendar-density:evidence-timeseries-pdc-calendar`** | **`needs_human_review`, UNTRIAGED — found 08-24 while verifying watch-points.** Something tried to change our article's locked timeseries section and the lock refused (the darts `misdirected_cta` rerender FAILED in the same window — likely the same event). Read the item, decide honour/reject. First-order: it is OUR page. |
| **A2** — `compute_component_quality` on editorial components | still never run (open since design PLAN Phase A). |
| **Phase B furniture** | standfirst / drop cap / pull-quote / rules / `stat-band` reuse. Timely: the 381 lane's writer-vocabulary work means richer semantic prose (h3/ul/ol/strong/table + bare blockquote) is coming — furniture styles it via scoped CSS, NEVER via classes in writer guidance (settled with them, on record). |
| **Rollout site 3** | mortgagecalculator / fundamentallyai strongest (live feed + evidence base). Recipe = 08-21 handoff §3. |
| **Cobot feature (#4)** | still PARKED on a primary 2024 source (negative result recorded — do not re-walk IFR's indexed pages). |
| **`published_page_id` reader** | data exists, reader still designed-not-built (`analysis_url` in `render_news_section_action.go` + 2 components' JS). Council-gated Go. |
| **Phase E1** | start registering dated, cited events in each new feature's substrate. E2's gating is DONE in advance: tag the component `requires-evidence-base` at registration — the 591/593 planner-menu gate is **live and both-ways proven** (381 lane, 08-24). |
| **bugs 349 / 198 / 296** | unchanged from the 08-21 handoff §4. |

## 4. Coordination state (all closed, context preserved)

- **357 / RFC_046**: told about pin-vs-stamp (D6 correction). Their identity
  stamping shipped. Nothing owed.
- **381 (writer html vocabulary + planner menus)**: CLOSED. Adopted from us:
  plain vocabulary, guidance explicitly forbids `<img>/<figure>/<iframe>`
  (the in-blob imagery loss class); gate at the menu row, not the vocabulary.
  Our `evidence-chart`/`evidence-timeseries` carry `requires-evidence-base`
  [MEASURED 08-24].
- **agentchassis-51 improvement-loop misfire at robot-hands (08-22)**:
  RESOLVED. Owner ran the bounded cancel (`UPDATE 23`); all tool-modifying
  items cancelled pre-dispatch; 0 locked rows of 62 writes; zero escalations;
  all overwrites archived. THEY own fixing the retargeting script
  (`076_improvement_loop_trigger.sh` ignores its args and always fires at
  robot-hands — their landmine + WRONG_CALLS entry). Full record: NOTES 08-22.

## 5. Session mechanics a fresh session must know

- **The permission classifier blocks DB WRITES from the session** (reads are
  fine). Protocol proven twice: write the SQL to the session scratchpad,
  idempotent with a `RETURNING`, and the owner runs
  `! kubectl … psql … -f - < <file>`. Never route a blocked write through a
  peer session (laundering — twice declined, correctly).
- **`content_components.updated_at` cannot date or attribute a change** — no
  trigger, no history for non-template columns. New LANDMINES entry (08-24,
  verifier corr `967dc071`) born from a real two-lane attribution chase whose
  answer was "the owner ran the script twice". Ask what would have to be true
  for the OTHER measurement to be right, first.
- **Same-file passengers are real and benign here**: our LANDMINES append rode
  the 333 lane's `68734b771`. Check `git log -S` before re-adding anything
  that "vanished".
- The five working docs live in TWO dirs (this one + `editorial_design_uplift/`)
  — design NOTES 08-22→08-24 tail is the composition thread's technical log;
  README_where_we_are has the owner-facing arc; first design-lane SUMMARY is
  `SUMMARY_2026-08-22_composition_plan_written.md`.

## 6. Traps

**All ten in the 08-21 handoff §9 still stand — read them there.** New since:

1. The stylesheet control (08-21 §9.1) was re-used twice this session and both
   times was what made the audit meaningful. It is not optional.
2. An impact report of a live queue is a FRAME, not an ending — the misfire's
   "findings, not actions" was true at 18:35 and false by 18:38 (promoter 900s,
   dispatcher 60s). Re-census before acting on any queue snapshot.
3. `UPDATE 0` from an idempotent script has two causes — already-applied and
   never-matched. The `RETURNING` clause plus a raw-column re-read separates
   them in one query.

## 7. What NOT to do

- Everything in the 08-21 handoff §10 (unchanged), minus the Fable-substitution
  line (moot — 035 exists).
- No `composite` rows / `parent_instance_id` values before P1's code ships.
- Do not add `{{template}}` or funcmap entries for composition — P0 proved none
  are needed; a P1 that "needs" one has diverged from the proven design.
- Do not read `component_version_id` as a pin (it is RFC_046's stamp), and do
  not add CASCADE to the parent FK (its fail-loud NO ACTION is load-bearing,
  035 §6.1).
- Do not put furniture classes into writer `llm_guidance` (settled with 381;
  pull-quotes etc. are design objects — 035 children or scoped CSS).
