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
`carryStoredSection`, and `deriveRenderMode` (~~:561/:639~~ **line numbers drifted;
re-read 2026-08-25: defined `store_generated_component_action.go:2024`, called from
:732 and :822 — the INSERT and UPDATE paths. `carryStoredSection` is
`rerender_page_sections_action.go:1242`** — runs on INSERT and
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
| ~~**`lock_blocked_change:darts-calendar-density:evidence-timeseries-pdc-calendar`**~~ **TWO items, cause identified, ACCEPTED — see §8** | ~~`needs_human_review`, UNTRIAGED — found 08-24 while verifying watch-points. Something tried to change our article's locked timeseries section and the lock refused (the darts `misdirected_cta` rerender FAILED in the same window — likely the same event). Read the item, decide honour/reject.~~ **CORRECTED 2026-08-25 — both halves of that sentence were wrong.** There are **TWO** items, not one (the second is `:robot-demand-step-change:evidence-timeseries-ifr`, 20 min later), and the cause is **not** the `misdirected_cta` event: it is the **283 / RFC_034 instance-scope conversion** of the shared `evidence-timeseries` template. Owner ruled **ACCEPT** 2026-08-25. Full trace, evidence and the new byte baseline: **§8 below**. |
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

---

## 8. The two lock_blocked_change items — traced, ruled, ACCEPTED (added 2026-08-25)

Written by the session that picked this handoff up. It corrects §3 and adds a trap that
bites P1 specifically. Everything here is measured 2026-08-25 unless marked otherwise.

### 8.1 What actually happened

Not the `misdirected_cta` event §3 guessed at. The chain, from the driver rather than
from the item:

| when (UTC) | what |
|---|---|
| 08-23 12:33:33.979 | `component_versions` v1 written for `evidence-timeseries` (`fb870e82-…`), `change_source='scope_component_instance'` — the **283 / RFC_034** lane |
| 08-23 12:33:33.990 | `content_components.html_template` updated, 11 ms later |
| 08-23 12:33:34.475 | `component-template-fixer` files two `section_edit` items, `spec.reason='template_changed'` |
| 08-23 12:41:46 | darts delivery refused by our lock → `lock_blocked_change` |
| 08-23 13:01:52 | robot-hands delivery refused by our lock → `lock_blocked_change` |

The lock gate behaved exactly as `section_editor_actions.go:335-362` documents — skip-result,
not error (bugs_open/058). Nothing was broken; the signal was working.

The batch converted **five** components, not one: Generic Text Block (187 instances,
12 locked), FAQ Section (88, 0), mechanism-flow (6, 1), `evidence-timeseries` (3, **3**),
plus a dartsonline tool component on 08-24. `evidence-timeseries` is the only one where
**every** instance is locked — so the conversion reached **zero** of its three consumers.

### 8.2 The change, and why accepting was safe

One line, 5,739 B → 5,738 B:

```html
- <section id="{{.ComponentID}}" class="ev-ts" data-component="evidence-timeseries">
+ <section id="{{.InstanceID}}"  class="ev-ts" data-component="evidence-timeseries">
```

`{{.ComponentID}}` was never the component uuid on our rows — the seed put the **slot name**
into `content_data.ComponentID`, which is why all three instances served slot-name ids.

**The reservation worth having** was that `{{.InstanceID}}` might not be *bound* on the
render path, in which case `missingkey=zero` ships `id=""` — a defect class the platform
only grew a detector for on 2026-08-24 (`reEmptyElementID`,
`component_instance_scope.go:208-236`). Probed at the artefact over the same batch:
**253 instances re-rendered since the conversion, ZERO empty ids** (GTB 161, FAQ 87,
mechanism-flow 5). The zero has real demand behind it, so it is informative.

**Dry-run gate before any write** (this is the reusable bit — see RUNBOOK): the local
harness reproduces both stored rows **byte-for-byte** from the v1 template + live
`content_data`, which is what makes the second render trustworthy. Rendering the LIVE
template with `InstanceID` bound changes **only the id**:

| page | id before | id after | delta |
|---|---|---|---|
| robot-demand-step-change | `evidence-timeseries-ifr` | `c-evidence-timeseries` | −2 B |
| darts-calendar-density | `evidence-timeseries-pdc-calendar` | `c-evidence-timeseries` | −11 B |

`InstanceToken(function, occurrence)` with occurrence 0 returns `"c-" + function`
(`component_instance_scope.go:102-115`) — **`c-evidence-timeseries`**, NOT
`evidence-timeseries-0`. Both pages get the same token; they are different pages, so
nothing collides.

**Owner ruling 2026-08-25: ACCEPT.** The lane recommended honouring (the id is inert —
class-based CSS, one reference per page, nothing selects on it); the owner chose
consistency with the fleet-wide convention. Executed as: unlock → re-dispatch the
`section_edit` → verify at the served artefact → re-lock → close both items `complete`.

### 8.3 ⚠ TRAP FOR P1 — the byte baseline MOVED on 2026-08-25

P1's acceptance is *"served page byte-equivalent"*. **That baseline changed today.**

| page | bytes before (08-24) | predicted after |
|---|---|---|
| robot-demand-step-change.html | 94,351 | 94,349 `[PREDICTED]` |
| darts-calendar-density.html | 92,883 | 92,872 `[PREDICTED]` |

`[PREDICTED]` = from the dry run, **not yet confirmed at the served page**. Whoever runs
P1 must re-measure both pages first and use the live number — do not take these two
figures as measured, and do not diff P1's recompose against the 08-24 baseline.

The related hazard, stated so nobody hunts it: `carryStoredSection`
(`rerender_page_sections_action.go:1242-1267`) carries `rendered_html` **verbatim**. So a
P1 walk that CARRIES a locked ev-ts row is byte-stable, and one that RE-RENDERS it is not.
If a P1 acceptance run shows an id diff on an ev-ts section, the walk took the re-render
branch — that is the finding, not a mystery.

### 8.4 Reported to peers, nothing owed back

- **283 / RFC_034 lane**: told we accepted, plus §8.4a below and the 4 pre-existing
  `id=""` rows on Generic Text Block (all pre-date the conversion — not theirs).
- **381 lane**: `period-calendar` boundary confirmed; they were told our E2 is **planned,
  not built**, and is specified in the `mechanism-flow` idiom.

**8.4a — a fan-out gap that emits no signal.** The third instance, oufe.com `thames-water`
/ `evidence-timeseries-leakage` (locked by `oufe-workstream`), got **no** `section_edit` —
only a whole-page `page_rerender` at 12:32:25Z, *before* the 12:33:33Z conversion. So no
delivery was attempted, no lock gate fired, no `lock_blocked_change` exists, and its lane
was never told. It still serves the pre-conversion id. **A locked instance that receives no
delivery attempt is indistinguishable from one that needed none** — so a coverage count
built on `lock_blocked_change` rows reads this batch as 2 blocked / 1 fine when the truth
was 3 unconverted. Not ours to fix; reported to the batch owner. There is no live `oufe`
session, so nobody there has been told directly.

### 8.5 ⚠ Verifying the acceptance — the item completing is NOT the page changing

Added 2026-08-25 from the **283 lane's** own measurement, and it is the trap most likely to
bite whoever runs the three scripts:

> Measured on that programme 2026-08-24: **three repairs all `complete`, with correct stored
> bytes, and one page served the old version for hours afterwards.**

So the verification order is **served page first, and nothing closes on the row**:

1. `section_edit` item reaches `complete` — necessary, **not** sufficient. Do not stop here.
2. `page_components.rendered_html` carries `id="c-evidence-timeseries"` — also necessary,
   **also not sufficient**. Stored is not served.
3. **`curl` both pages** and assert on the served bytes, with the stylesheet control run
   first (08-21 §9.1): positive control `c-evidence-timeseries` **present**, negative
   control `id=""` **absent**, and the byte delta fully accounted for by the id.
4. Only then re-lock (`B_relock.sql`), and only then close the two items
   (`C_close_lock_items.sql`).

If (1) and (2) pass but (3) still shows the old id, that is a **publish lag, not a failed
edit** — the same shape the 283 lane measured. Wait and re-check before dispatching
anything else; a second delivery on top of a pending publish is how you get two writes
racing at an unlocked flagship row.

**Post-conversion census for the peers:** all three `evidence-timeseries` placements were
unconverted as of 2026-08-25 (0 of 3 at the placement level — ours refused, oufe's was never
asked). The 283 lane's fleet census the same day: ~~**48 of 437 placements unconverted, 26 of
them locked.**~~ **CORRECTED same day, by them, and I had already repeated the bare number
here — which is exactly how it misleads:** 48 unconverted is **consistency debt, not
damage**. A literal element id only bites where the same component appears **twice on one
page**. Narrowed: **48 unconverted → 8 on a multi-instance (page, function) pair → 1 page
with a genuinely duplicated id → 0 reaching a visitor.** The one page is
`webdesign.uk/index.html`, and `webdesign.uk` 302-redirects to the separate `webdesign.co.uk`
site, so it serves nobody (they ran the parked-domain redirect control and a 4-id negative
control on the followed target). **The standing risk is the redirect, not the count:** if
`webdesign.uk` ever serves its own content, that page becomes damage with nothing else
changing. Once ours land, tell them so their count moves 3 → 1 rather than being
re-derived. The oufe row is covered by their CONTRIB at
`docs/agent_docs/docs024_key_docs_latest/oufe/CONTRIB_2026-08-25_from_283_lane_thames_water_evidence_timeseries_never_took_the_scope_conversion.md`
— **not ours to action**, and that page separately has three `lock_blocked_change` rows
sitting at `needs_human_review` since 2026-07-29.

### 8.6 P1 seam re-read — done 2026-08-25, three constraints the walk must satisfy

The handoff's §1 says re-read the seams FIRST. Done; here is what they actually say, so
the next session starts from the constraints rather than re-deriving them.

1. **The walk must call `RenderTemplate`, and must not become a second executor.**
   `render_seam_one_spelling_test.go` enforces three rules by **AST** (not grep — its own
   prose contains the symbol names, which is why): exported `RenderTemplate*` symbols must
   be exactly `[RenderTemplate, RenderTemplateWithMap]`, so **do not name the walk
   `RenderTemplateComposite`** or anything matching; only `RenderTemplate` may call
   `executeGoTemplate`; and any new function that builds AND executes a template must be
   added to `declaredTemplateExecutors` with a justification naming its dialect. A walk that
   calls `RenderTemplate` per node adds nothing to that map and needs no test change — which
   is the design 035 D4 already assumes. Note the test carries its own **control** (it fails
   if the traversal parsed too few functions), so a green result means something.
2. **`RenderTemplate`'s signature is `(templateStr, ctx, logger) → (html, missing,
   inURLAttr, err)`** (`component_library.go:1060`). The two reports are not optional
   sugar: discarding them is how `bugs_open/238` shipped five `<img src="">` to a live
   homepage. A walk that discards them must write `_, _,` visibly at each node.
3. **`deriveRenderMode` must check `slots` BEFORE the llm-field loop.** The function
   (`store_generated_component_action.go:2024`) returns `agent` the moment any field
   declares `source: "llm"`, else `template`. D3's own worked example declares `slots`
   **and** a `standfirst` llm field — so with the checks in the other order every composite
   with an llm field derives as `agent` and never routes to the composition build. Both
   call sites (:732 INSERT, :822 UPDATE) take the same function, so one edit covers the
   regeneration-reverts hazard (§6.6).

Plus **§6 hazard 9**, added the same day and new: one `RenderContext` reused across the
walk forges RFC_046 provenance, because the digest is MUTATED onto the context, not
returned. Today's code is safe only because every reader renders exactly one template per
context — the walk is the first thing that would not. Falsifier and rule in 035 §6.9;
the 357/RFC_046 lane has been told.
