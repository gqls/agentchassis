# FOCUS — Design/Composition Flow Gaps and the Adoption-Fidelity Model

Date: 2026-05-26 (updated same day after the trigger fix deployed)
Status: standing reference. Records an investigation into why a built site can
have no installed theme, the structural gaps that produce it, and the
(already-decided, partly-implemented) adoption-fidelity model those gaps sit
inside.

UPDATE 2026-05-26: gap (A) is being deployed to production now —
`emit_design_items` + `emit_imagery_items` are wired into build-site-planner
(`agent-chassis v1.0.1047`) as plan-time steps. The trigger gap is closed for
both design and imagery. Gaps (B) and (C) remain open; the deploy makes (C) more
visible because composition now runs but resolves a generic base palette. See
§3 for the per-gap status.

Cross-refs:
- `028_platform_mission_and_pipeline_direction.md` — mission + fidelity dial + phasing
- `FUTURE_adoption_source_destination_separation.md` — adoption variants (clone/analysis)
- `FOCUS_adoption_faithfulness_via_locks.md` — timed-lock enforcement of faithfulness
- `FOCUS_planner_ignores_adopted_state.md` — the planner-drift failure mode
- `027_design_and_site_planner_v2.md` — site-design-planner composition resolver
- `FOCUS_site_spec_vs_site_plan.md` — spec aspects vs the site_plan layer
- `emit_design_items_action.go` — the restored design trigger (deploying 2026-05-26)
- `emit_imagery_items_action.go` + `imageryplan.go` — the build-time imagery trigger and its shared selection/classify/spec package (deploying 2026-05-26)

This doc started from a robot-hands.com question but the substance is general.
robot-hands is used below only as a worked example / evidence; the conclusions
are about the flow.

---

## 1. How a built site ends up with no theme

The live render path resolves colour/typography from the **theme layer**, not
from a spec aspect:

```
sites.style_collection_id → style_collections (color_palette, css_theme_id)
                          → css_themes (color_palette, css_content)
```

Confirmed in the chassis actions: render context sets `PrimaryColor` /
`AccentColor` / `SecondaryColor` from the resolved collection
(`coll.ColorPalette[...]`) and from `css_themes`/`style_collections`
(`sc.color_palette->>'primary'`, etc.). Nothing in that path reads a
`site_specs` design aspect.

robot-hands.com has `style_collection_id = NULL` — no collection, no css_theme,
no palette in the render path at all. That is not a one-site quirk; see gap (A)
below — it is the expected outcome for sites built through build-site-planner
since the Phase-1 refactor.

---

## 2. The spec-retire investigation (what prompted this)

The robot-hands site_specs carry two stale adoption-seed aspects from the
original `site-adoption-agent` run (2026-03-31):

- `design` (source=adoption) — a **palette** (`{"palette":{...},"typography":{...},
  "visual_tone":"cyberpunk terminal..."}`), green/cyan terminal scheme.
- `structure` (source=adoption) — `{"pages":["index","product-detail"]}`, the dead
  2-page seed.

Consumer trace (the reason it is safe to retire them, corrected from the
original handoff's loose "superseded by design_intent / site_plan"):

- `site_specs.specs.structure` — **zero readers** anywhere (actions, agent
  definitions, registry). `site_plan` (build-site-planner) is the page-set
  authority. Dead residue.
- `site_specs.specs.design` — **one** reader: `tool-recreation-handler`'s
  `recreate_tool` prompt, as an `{{if}}`-guarded optional "Original site design"
  block. That agent is the *adoption* tool-recreation path; a fresh build routes
  tools through `tool-suggester → tool-generator / tool-deployer`, so it does not
  run. Even if it did, a missing aspect degrades gracefully.

Correction made during the investigation: I first claimed `design_intent` does
not contain the palette. With the full payload, it does — `colour_mood` carries
the scheme with real hex. So the palette is **not** lost by retiring `design`;
`design_intent` holds a (different, refined) version. Retiring both aspects is
therefore safe on their own merits, and is **not** the cause of the NULL theme
layer.

Note on `design_reference`: it is owned by `site-adoption-agent` and records the
*original adopted site's* extracted design. It is not a slot to write our chosen
palette into — doing so would break the ownership model and misrepresent the
source. (This corrected an earlier proposal to port the palette there.)

---

## 3. The three stacked flow gaps

These are distinct, independently real, and stack on top of each other. All
three must be addressed for design composition to come out right; fixing any one
alone is not sufficient.

### (A) Composition is never triggered — missing build step [CLOSED 2026-05-26]

The legacy site-planner emitted `needs_composition` + `needs_design` inside
`WriteBuildItemsAction`. build-site-planner's Phase-1 refactor moved the terminal
step to `write_site_plan` + `reconcile_site_plan`, **neither of which emits the
design trigger**. So every site built through build-site-planner since that
refactor has had no composition installed unless the improvement loop's
`missing_css` discovery check happened to backfill it. This is the missing step
identified in another thread.

`emit_design_items_action.go` restores it as an explicit plan-time step
(deployed 2026-05-26, chassis v1.0.1047):
- `needs_composition` → `site-design-planner` (priority 7), which resolves
  palette/layout/typography and installs the css_theme + style_collection,
  setting `sites.style_collection_id`.
- `needs_design` → `webdesign-agent` (priority 8, `depends_on` composition),
  which renders and commits `styles.css`.
- Guarded on `style_collection_id IS NULL` (no backfill, no duplicate emission on
  replan). Plan-time, not reconcile-time, so the scheduled reconcile tick does
  not backfill every NULL site on a cadence.

**The imagery trigger had the identical bug and is fixed in the same deploy.**
`write_site_plan` records the planner's image requests in `site_plan_imagery`
(`flattenImageryBlock`), but nothing on the build path acted on them —
`needs_imagery` was emitted only by the loop's `unfulfilled_imagery_plan`
discovery check. `emit_imagery_items_action.go` is the build-time emitter, and
`imageryplan.go` is a shared package holding the row selection, priority/severity
classification, `brand_update` rule, `item_key`, and spec body used by *both* the
build emitter (status `triaged`) and the loop check (status `detected`) — the
same anti-drift pattern as the design side. Build-time imagery is priority-banded
(index hero 65, logo 70, site 75, page-hero 80, page-other 90, section clamped to
98) so it precedes the terminal `needs_rerender` (99) and lands in the first deploy.

Deployed wiring in `build-site-planner.default_config.workflow` (v1.0.1047):
`reconcile_site_plan → emit_design → emit_imagery → complete`, with
`emit_design` / `emit_imagery` each taking `config.site_id: input_data.site_id`
and `output_field` `design_items` / `imagery_items` (both listed in
`complete.output_fields`). The `emit_design` REUSE NOTE still stands: its insert
block mirrors `WriteBuildItemsAction`; extract a shared
`emitInitialCompositionAndDesign(...)` helper so the copies cannot drift.

Two non-blocking observations from the deployed config:
- `write_site_plan`'s step description omits `site_plan_imagery` though the action
  writes it — stale doc-string, not a bug.
- `emit_imagery` has no site-level no-backfill guard like `emit_design`'s
  `style_collection_id IS NULL`; it tops up any plan imagery row lacking an active
  asset on every replan (dedup prevents double-queue). Probably intended, but an
  asymmetry between the two emitters worth remembering.

This was the prerequisite — without the trigger, composition never ran and gaps
(B)/(C) were not observable. With it deployed, fresh sites now install a theme and
emit imagery at build time; the *quality* of the resulting palette is gaps (B)/(C).

### (B) The planner drifts from the adopted state [OPEN — not addressed by the 05-26 deploy]

robot-hands was adopted from the live robot-hands.com; the adopted design was a
green/cyan terminal scheme. Under a faithful high-fidelity adoption (the current
de-facto default, see §4) `design_intent` should have **preserved** that. Instead
build-site-planner wrote `design_intent` as a blue/charcoal "professional-dark
engineering dashboard" (`colour_mood`: electric blue `#0080FF–#00B4FF`, amber
`#FFB800`; `typography_mood`: clean sans-serif, "no monospace except for values")
— a different scheme, and one that does not match the `identity` spec's
"sci-fi/cyberpunk, system-level terminology."

This is the documented **planner-ignores-adopted-state / identity-drift** failure
(`FOCUS_planner_ignores_adopted_state.md`; the gamedesign.uk "drifted to
consultancy" case). It is not a missing contract — it is the planner generating
generic design direction that does not condition strongly enough on the
identity/classification it was handed. The designed answer is the fidelity dial +
timed locks (§4), of which only the coarse Phase-1 prompt behaviour exists.

The upstream lever: make build-site-planner produce a `design_intent` faithful to
identity/classification in the first place. That makes everything downstream
correct without enforcement.

The deployed `plan_site` prompt (v1.0.1047) added an "Existing Pages — ALREADY
BUILT, PRESERVE EXACTLY" block, which fixes *page* drift for adopted sites — but
design drift has two structural causes that block leaves untouched:
- **The adopted design is never shown to the planner.** `plan_site` renders only
  `site_specs.specs` identity / classification / strategy / briefing /
  mission_brief / roadmap_brief, plus `existing_pages`. It does **not** render the
  adopted `design` or `design_reference` aspect. The planner cannot preserve a
  palette it never sees; `design_intent` is generated free-form ("colour feeling
  that fits the industry").
- **The vocabulary cannot hold the adopted aesthetic.** `design_intent.style_direction`
  is a fixed 3-value enum — `professional-dark | modern-light | bold-creative` —
  with `validate_plan.default_style = "professional-dark"`. "Cyberpunk terminal"
  has no expressible value, so robot-hands' cyberpunk `identity` collapses to
  "professional-dark engineering dashboard" even when the planner reads the
  identity. The enum itself forces the drift.

### (C) The planner's colour decision reaches the resolver only as prose [OPEN — more visible after the 05-26 deploy]

`site-design-planner`'s palette cascade (doc 027) reads, in priority order:

```
design_reference.suggested_mapping  (adopted source hex — adoption-faithful)
→ mission.preferred_palette
→ design_intent.palette.reference_values  (structured hex)
→ selected layout's seed palette
→ default palette
```

But build-site-planner writes colour as **prose** in `design_intent.colour_mood`,
which `write_site_plan` flattens into `site_plan_directives` rows (category=design,
subjects: colour_mood, style_direction, typography_mood, imagery_direction,
layout_preference, avoid). It does **not** write a structured
`design_intent.palette.reference_values`. So for any site whose colour lives only
in `colour_mood`, the cascade's design_intent slot misses and falls through to the
layout-seed/default palette; the intended colours re-enter only via the
`webdesign-agent` overlay (the one LLM step that reads the prose), not the base
composition.

This is the same producer/consumer schema-drift shape as the features
`{title,description}` vs `{icon,name,description}` and the
`constraints.aspect` vs `style_hints.aspect_ratio` mismatches: one stage emits a
shape the next stage does not read.

The resolver's expected shape (from `createPalette`): core keys are
`primary, secondary, accent, background, surface, text, text_muted, border`
(extra keys pass through). The fix is a **data-shape** change, not a value
freeze: build-site-planner should emit `design_intent.palette.reference_values`
in those keys *alongside* the prose `colour_mood`, so the planner's already-made
per-site decision survives the trip to the mechanical resolver.

Important: `palette.reference_values` is **not** a constraint that forces or
shares colours across sites. `site-design-planner` does no choosing (doc 027: no
LLM calls, thin wrapper over `createPalette` which stores whatever it is handed).
The structured slot only decides whether the per-site decision is consumable. The
choosing — and whether colours are preserved or re-chosen — is governed by §4, not
by this shape.

---

## 4. The adoption-fidelity model (decided; partly implemented)

This is settled design across doc 028, the variants doc, the faithfulness-via-locks
doc, and chats of 2026-02-24, 04-13, and 05-03. Recorded here because the §3 gaps
only make sense inside it.

**Unifying idea (02-24 chat).** Every input — bare domain, direction JSON,
classifier research, questionnaire, HITL, improvement suggestions, a scraped live
site — is the same thing at different **fidelity**: an answer to "what should this
site be?" A scraped live site is near-total fidelity; a bare domain is almost none.
Adoption is the high-fidelity end of one pipeline, not a separate mode.

**Two axes, not one.**

*The fidelity dial* (doc 028 + 05-03 chat) — how much aspiration reaches the first
build and how fast the improvement loop narrows the gap:

| Level | First build | Improvement loop |
|---|---|---|
| **locked / absolute** | faithful spec only, no extensions | runs no aspirational work |
| **high** | preserve adopted identity/palette/voice; minimal aspiration | promotes slowly |
| **medium** | adopted character + modest non-conflicting extensions | promotes faster |
| **low** | full aspirational spec incl. less-certain extensions | aggressive |
| **no adoption (blank)** | dial re-purposes as research confidence tolerance | default medium |

`locked/absolute` is the "just a copy, don't evolve it" case raised on 05-03 as a
fifth point beyond high/med/low.

*The variant axis* (FUTURE_adoption_source_destination_separation + the
consolidated fidelity-and-variants doc) — what the operation *is*: `clone`
(default, faithful copy / Variant C full-clone-with-substitution) vs `analysis`
(take the best parts, build something that is not a clone).

**Enforcement: timed locks** (faithfulness-via-locks doc). The faithful first pass
is locked (~90 days default) so the planner physically cannot drop adopted
pages/aspects; expiry then lets agents restructure. The existing site/component
locks can serve this now, optionally tightened to "update content but don't evolve
the spec."

**Design-specific decisions (04-13 chat).** No fidelity *loop* — a loop implies
expected repeated failure and would generate problems no templated agent can fix.
Instead: extract a fingerprint (concrete CSS values) → write a `design_intent`
mapping the source's values to our CSS-variable conventions → webdesign-agent
matches → component layout *approximates*. CSS (palette/fonts/spacing) does ~70%
of the character work, layout ~30%. If reproduction is intended, `design_intent`
should explicitly say "reproduce the original's palette"; evolution then comes as
deliberate `design_intent` updates, not silent drift.

**Implementation status (the catch).** Only **Phase 1** exists: an adoption-aware
classifier prompt giving implicit `high` fidelity at the prompt level. The real
dial needs:
- **Phase 2** — per-item status on specs (the prerequisite that makes
  fidelity a deployed-vs-planned partition rather than coarse prompt-nudging).
- **Phase 3** — explicit fidelity input parameter (`build_policy` / `adoption_meta`).
- **Phase 4** — classifier produces aspiration alongside the faithful baseline,
  status-marked.

So today fidelity is coarse prompt behaviour, not the deployed-vs-planned model.

**This answers the choice-vs-freeze question.** Whether an adopted palette is
preserved or re-chosen is governed by fidelity + lock, not by the
`palette.reference_values` shape from §3(C). At locked/high the adopted palette is
preserved (and `design_intent` should literally carry "reproduce the original");
at low it is free. The §3(C) shape work is orthogonal plumbing.

---

## 5. Failure-mode principle (doc 028, restated)

An agent that changes behaviour on information it did not put in the spec is a
bug (a different agent won't see it). An agent that overwrites a spec aspect
another agent owns is a category error (breaks the implicit ownership model). An
agent that produces the right output but doesn't write it to the spec is not
helpful (downstream relies on reading it back). Silent override is the failure
mode being eliminated. Both §3(B) drift and §3(C) shape-drift are instances.

---

## 6. Open decisions / next steps

Ordered by dependency. None requires touching robot-hands manually.

1. **Deploy `emit_design_items` + `emit_imagery_items`** (gap A) — **DONE
   2026-05-26 (v1.0.1047).** Composition, design, and imagery now triggered at
   plan-time. Follow-up still open: extract the shared
   `emitInitialCompositionAndDesign` helper so `emit_design`'s insert block can't
   drift from `WriteBuildItemsAction`.
2. **Verify the deploy** — on the next fresh build, confirm `style_collection_id`
   becomes non-NULL and a `css_theme`/`style_collection` is installed, and that
   `needs_imagery` items appear at build time (not just from the loop). This is
   how we see (C)'s symptom directly: composition installed, but check whether the
   resolved base palette matches the planned `colour_mood` or is the layout-seed
   default.
3. **Planner fidelity** (gap B, now the live design item) — two concrete fixes:
   (a) feed the adopted `design`/`design_reference` aspect into the `plan_site`
   prompt so adopted palettes can be preserved; (b) widen or replace the
   `design_intent.style_direction` 3-value enum so it can express aesthetics like
   "cyberpunk terminal" instead of collapsing to "professional-dark." Decide
   alongside this how adoption faithfulness is enforced now (existing locks vs
   waiting on Phase 2/3).
4. **Channel alignment** (gap C) — make build-site-planner emit
   `design_intent.palette.reference_values` (core keys: primary, secondary,
   accent, background, surface, text, text_muted, border) alongside `colour_mood`,
   so the base composition reflects the planned colours instead of relying on the
   webdesign overlay.
5. **Adoption policy** — confirm where on the variant/fidelity axes adoption sits
   by default, and whether Phase 2 (per-item spec status) is the next structural
   investment, since it gates the real dial.

robot-hands-specific cleanup (retire `design` + `structure`; backfill its
`design_intent`) is deferred — it is a worked example, not the target. The
general flow is the target.

---

## 7. Evidence index (for verification)

- site_specs shape / partial-unique `idx_site_specs_current` — `schemas_all` ~2957–2986
- robot-hands current aspects + `design`/`design_intent` payloads — live reads (this session)
- `style_collection_id = NULL` for robot-hands — live read (this session)
- theme-layer render reads — chassis actions ~13548, 13595, 17464, 17517, 27934–27944
- only `design` reader = `tool-recreation-handler` — `bk_agent_definitions_backup.sql:269`
- zero `structure` readers — grep across actions / agent defs / registry
- palette cascade order — `027_design_and_site_planner_v2.md`
- `colour_mood` → `site_plan_directives` flattening — chassis actions ~72019–72050
- `createPalette` core keys — chassis actions ~29068–29069
- design trigger restoration — `emit_design_items_action.go` header + body
