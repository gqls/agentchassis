# HANDOFF — news_editorial_features, 2026-09-03. START HERE.

Supersedes `HANDOFF_2026-08-25_continue_here.md`. **Its predecessors are not dead and
this file does not restate them:**

- **08-21 §3 (the page recipe, proven three times) and §9 (the ten traps)** still
  govern before shipping a feature page.
- **08-24 §8** is the instance-scope evidence trail; **08-25 §1 is DONE** (the three
  SQL scripts ran, the directory is deleted — see that file's banner).
- **The companion lane's `editorial_design_uplift/HANDOFF_2026-09-02_continue_here.md`
  §9 is the post-roll verification for this lane's code.** Read it rather than
  re-probing; its probe hazards are inherited below.

- **`ARCHITECTURE_2026-09-03_who_owns_composition.md`** (this directory) answers the
  owner's "does composition need its own loop" question with measurements. §7 below
  summarises it; do not re-derive it.

Everything is measured 2026-09-03 unless marked. `IMAGE_TAG` is `v1.0.1357`.

---

## 1. THE HEADLINE — P1's code SHIPPED, and P1 is NOT DONE

**`035 P1` is in `v1.0.1355`** (companion lane's fully-controlled binary probe,
controls on opposite sides). `rerenderFlatSections` and `hierarchyChildrenOf` are
PRESENT in the binary.

**⚠ `recomposeAncestors` is ABSENT from the binary, and that is CORRECT.** It has **no
caller** — every source reference is inside its own definition, so the linker
eliminates it. The commit (`3fd617ef6`) shipped; the capability is unreachable.
**An absent capability literal means "not reachable in this build", never "the commit
did not ship".** Verified again 2026-09-03: `grep recomposeAncestors(` outside its own
definition returns nothing.

**So the true state is: the READ path is live and inert; DIRECTION 2 IS WRITTEN AND
UNWIRED.** That is the single most important fact in this file.

| piece | state |
|---|---|
| walk + `deriveRenderMode` third value | shipped, inert (`1f745e730`) |
| `loadStoredSections` reads `id`, `parent_instance_id` | shipped (`bd811fa93`) |
| membership helpers | shipped, called (`bc8167100`) |
| `recomposeAncestors` | shipped, **NO CALLER** (`3fd617ef6`, `fdb03cbc1`) |
| direction-1 guards, both paths | shipped and wired (`028c3e112`) |
| the extraction + decomposition | shipped (`2a0bdb001`, `94f81cc60`, `22ed53ee7`) |
| 137 residue fix, both twins | shipped (`3b1389ca0`) |
| **the walk BRANCH in the render loop** | **NOT WRITTEN** |
| **byte-identity tests + mutation check** | **NOT WRITTEN** |
| **migration 643** | **VOID — see §3** |

Composition remains unreachable: **0 of 3,035 `page_components` carry a
`parent_instance_id`; 0 of 386 `content_components` declare a `slots` block.**

## 2. WHY THE METHOD CHANGED — plans to code

**The council said REVISE three times on correlation
`53d71504-8cd1-49bc-8e2d-d1465ba65103`, and the GATING objection was identical every
round: a sketch calling symbols no edit defines.** Round 1 `recomposeParent`
undefined; round 2 renamed, still undefined; round 3 defined, its three helpers
undefined. One level deeper each time.

That is not a documentation slip. It is what authoring code in prose produces: every
sketch bottoms out in new symbols and nothing stops it. **Round 4 must be real diffs
from compiling, tested code.** Do not submit another sketch.

**Writing it found five things three rounds of review could not.** Each is recorded
where it lives; they are listed here so the next session does not re-derive them.

1. **The `db`/`tx` split is FORCED.** `buildRenderContextFromDB` and
   `DeriveAndBindInstanceToken` take `*sql.DB` and cannot take a `*sql.Tx`.
2. **Two estate guards caught the new `rendered_html` writer within minutes** — the
   floors coverage test and `pattern-check`'s `unrepaired-component-write`. Neither
   was taken as an exemption; both are now called.
3. **The two direction-1 guards are DIFFERENT PREDICATES.** `apply_section_edit`
   holds a ROW → `hierarchyChildrenOf`. `RenderComponentAction` holds a component
   DEFINITION, has no row, and RETURNS html → `hierarchySlotsFromSchema`. The
   submitted plan used the row predicate for both and **would not have compiled**.
4. **The loop runs TWO counters with DIFFERENT advance rules.** `sectionOccurrences`
   advances for EVERY row including carries (per-section imagery); the instance token
   counter advances only for RESOLVED rows. The walk must thread both, differently,
   and getting either wrong is silent.
5. **§3 below.**

## 3. ⚠ NOTHING READS `render_mode` — P1's routing story has no landing place

`[MEASURED 2026-09-02]`, the whole chain:

- **WRITTEN** by `deriveRenderMode` (`store_generated_component_action.go:732` INSERT,
  `:822` UPDATE)
- **SCANNED** into `Component.RenderMode` (`load_component_library_actions.go:228,275`)
- **READ** by nothing. `.RenderMode` has no other use in `platform/` or `internal/`.

And 035 D3's claim that `check_render_mode` "routes sections to LLM generation when
`render_mode == 'agent'`" is **false three ways**: it is not a top-level step (it sits
under `{workflow,steps,process_sections_loop,config,sub_workflow,steps}`); it is a
**binary conditional** (`condition`/`then_step`/`else_step`), not a `conditions` array;
and it branches on `current_section.llm_field_specs`, **not on `render_mode` at all**.
The step containing "rerender_mode" is `check_rerender_mode` on the **page-rerender**
agent, branching on `input_data.spec.reason` — a different step on a different agent.

**Consequences.** Migration 643 is **VOID — drop it, do not re-derive it.** There is
nothing to add an arm to and adding one routes nothing. `deriveRenderMode`'s third
value is inert in a STRONGER sense than earlier files claim: not "inert until a slots
block exists" but inert even then.

**✅ RULED 2026-09-03 (owner): the routing arm MOVES FROM P1 TO P2.** P1's own
acceptance never needed it — recompose one page by hand, prove byte-equivalence,
rewrite one child, prove siblings untouched: all hand-made rows and the edit path. P1
is therefore exactly what its name says, the READ path, and it can finish. **P2 owns
the routing, and its precondition is a consumer for `render_mode` (or a different
signal).** Do not re-open this in P1.

Also verified live: `jsonb_set(doc, path, NULL || jsonb_build_array(...))` returns
**NULL for the whole document**. Any future edit to that step must guard the array
shape before writing.

## 4. What to do next, in order

1. **Wire direction 2.** `recomposeAncestors` has no caller — call it from
   `apply_section_edit` after the child write, in the same transaction, and file the
   `needs_rerender` item for each stale ancestor it reports. This is the piece whose
   absence the binary probe is currently reporting.
2. **The walk branch** behind the `hierarchy_walk` step-config flag, threading BOTH
   counters (§2.4).
3. **The byte-identity tests** with a mutation check — and note the guarantee now
   RESTS on them (§5).
4. **Round 4** as real diffs, on the same correlation, with migration 643 dropped and
   §3 explained.
5. Then the live recomposed page, and the register entry saying **"exercised on
   `<page>`" with a date** — never "deployed" (035 §6.8).

## 5. A DESIGN FORK I TOOK — the council was told the older story

Rounds 2 and 3 promised the flag-OFF branch would be the **byte-identical original
loop**. It is not, and cannot be: keeping it would mean duplicating ~200 lines,
because a composition child needs the same carry semantics as a section, and two
copies drift.

**So there is ONE implementation, and the guarantee moved from a structural claim to
the byte-identity TEST.** The flag now guards only whether the walk composes children.
**This makes the test's mutation check load-bearing rather than decorative** — a
byte-identity test never shown to fail proves nothing. Say this plainly in round 4;
the seats were told otherwise.

## 6. G6 — Fable's draft is written and its four decisions are RULED

`features_open/035_G6_DRAFT_stored_pattern_library.md` (`30d161249`), written by Fable
at the owner's direction. 035 itself untouched. **A pattern is a promoted
`content_components` row** — concrete slots, `component_id` per slot, a
`pattern_provenance` jsonb, `created_from='promoted'`; **no new table**, because a
second home for slot declarations is the drift class D3 already retired.

The **loops clause is satisfied compositionally** — parts scored at their grain, the
whole ruled by the experience council because a promotion IS a design decision — with
**no third loop**, and Fable flagged that as a **weakening for the owner to overrule**.

**✅ RULED 2026-09-03 (owner): Fable's route is ACCEPTED.** The owner read §7.1 and
§3 verbatim before ruling, having asked not to decide against Fable without the full
reasoning. What was accepted, and what the next session should treat as settled:

1. **The clause's meaning — COMPOSITIONAL NOW, instrument later if wanted.** The
   component loop vouches for the PARTS at the grain it already emits; the experience
   council vouches for the WHOLE, because a promotion IS a design decision and that is
   the object class it already rules on; live service is evidence and is recorded WITH
   its demand control. **No third loop.** The decisive property is that this is
   REVERSIBLE — an arrangement-grain instrument can be commissioned later **without
   changing the stored object**.
   ⚠ Note precisely what this extends: the council's **subject-matter** (it has never
   ruled on anything editorial), **not its grain, machinery or roster**.
2. **Library scope** — advisory scoping via the existing `suitable_site_types` /
   `suitable_page_types`; no new schema for hard per-site scoping.
3. **Version binding** — record the judged versions in provenance, follow the library
   on application, pin nothing.
4. **Auto-apply** — not shipped; revisit only on a measured approval rate.

**⚠ THE READING TO CORRECT IF WRONG:** the owner said "I accept fable's route" after
being shown §7.1 specifically. This file records all four as accepted, because they are
Fable's four recommendations and form one coherent route. **If only §7.1 was meant,
items 2–4 revert to open** — they are cheap to re-decide and none of them blocks P1.

**⚠ 035 hazard 10 blocks G6's gate.** Verified by probe 2026-09-02:
`extractTemplateVariables` returns deduplicated FIELD ROOTS, so
`{{.slots.lead}}{{.slots.quote}}` yields **`[slots]`** — one name for two slots. Every
composite therefore mis-scores in `compute_component_quality`, silently, as a low
score rather than an error. G6's promotion gate requires a non-NULL `quality_score`
per child family, so it would gate on a number computed wrongly. **This is P1/P2's to
fix, not G6's.**

## 7. ARCHITECTURE — asked and ANSWERED 2026-09-03; the answer awaits a ruling

> *"Do we need another loop like the experience loop that sorts out this component
> composition? we have theme kits and page composition and experience loop and visual
> designer and designer, might we extend one of their remits or another agent or
> workflow remit or is it a separate responsibility. I think whatever composition we
> decide on it should act like the components — it can have defaults and sites or
> pages can fork as they want."*

**ANSWERED in full: `ARCHITECTURE_2026-09-03_who_owns_composition.md`** (this
directory, `3964691b9`). Measured against the live DB and tree. **Read that rather than
re-deriving it.** The four load-bearing findings:

1. **"Composition" is TWO things and the word is doing double duty.** **Grain A** is
   WITHIN a section — 035's parent/child, 0 of 3,035 rows. **Grain B** is ACROSS a page
   — which sections in what order — which **already exists and is fully live** as
   `site_plan_sections` (54 plans, 34 sites). A THIRD thing wears the same word and is
   neither: `site-design-planner` "resolves composition (palette, layout, typography)",
   i.e. THEME composition. Almost every downstream question resolves differently
   depending on which is meant.
2. **The owner's defaults-and-fork principle is ALREADY SATISFIED at grain A**, by
   construction: a composite is a `content_components` row, so it inherits the library
   default (`forked_from IS NULL`), the live fork mechanism (**412 library / 85 forks /
   88 page rows across 25 sites**), `component_versions`, and the quality sweep.
   **Grain A needs a decision NOT to build.**
3. **The gap the principle names is at grain B.** `site_plans.site_id` is `NOT NULL` —
   no library plan, no template, no fork. Every site's arrangement is generated from
   scratch. ⚠ And the cautionary measurement: `layouts` HAS a defaults-and-fork
   mechanism (`forked_from_layout_id`) and it is **18 library, 0 forked — never once
   used.** This estate has already built one such mechanism for structure and not
   driven it.
4. **NO NEW LOOP.** A loop needs a driver; the judgement already has a home at the
   right grain (a promotion IS a design decision); the parts already have a scorer; and
   a new loop would have nothing to score.

**Recommended remits, NOT YET RULED:** grain A → the component library, no new owner.
Grain B → **`site-design-planner` extended**, because it already does the analogous job
one layer down (match a library default, fork when the site differs, HITL item carrying
the reasoning). The one genuine schema question is grain B's: an arrangement library
needs plan rows not owned by a site. Five other candidate owners are named in the
document with the reason each is wrong.

**⚠ A CORRECTION THIS LANE OWES ITSELF.** An earlier version of this file, and my note
to the owner, implied the grain A/B split undercut Fable's G6 §7.1. **It does not** —
neither loop judges arrangements at EITHER grain, so §7.1's answer is unchanged by the
split. The grain question bears on §7.2 (a plan-level library is a different schema
question from a component-level one) and on WHICH OBJECT gets stored. It is orthogonal
to the loops clause, not a refutation of it.

## 8. Coordination — four lanes, all answered, nothing owed

- **`bugfix_410` (scan-loss)** — their guard shipped into `loadStoredSections`
  (`7c443aac6`, DBI-027). **Standing constraint on this lane: any column added to
  that projection must be `NOT NULL` or `COALESCE`d** or their guard fires on our
  edit. Ours are fine and round 4 adds none.
- **`inline_guide_imagery` / dartsonline guides** — answered: do NOT wait for P5;
  **the guides are not a P2 consumer** (figure-between-sections suffices, so finer
  FLAT rows give them durability today). Recorded in 035 §8 with their scope caveat.
- **`loanzy.uk` (UK news region)** and **`boxingonline`** — both routed away as
  name-matches on "news"; this lane owns editorial FEATURE PAGES, not feed ingestion.
  The boxingonline handoff carried a verified `evidence_base` finding (2 rows, 3
  facts, 0 `allowed_entities` against sites carrying 111–116 facts per row).
- **`bugfix_357`/RFC_046** — the primitive fix for the provenance stamp is theirs and
  must NOT be bolted into P1.

## 9. Traps from this stretch

1. **A tool's verdict is scoped to its ARGUMENT.** I ran `probe-page-url.sh` against a
   customer domain, got a true CONTROL-FAIL about *that domain*, and restated it as a
   fact about *the site* — then asked a peer to put it in a bug file. The serving host
   discriminates perfectly. **Probe the slug, never the customer domain**
   (`sites.publish_target` names what serves).
2. **`bugs_open/137` is CLOSED** (`bugs_closed/137`, live since 2026-07-31). The
   `pattern-check` advisory's own text still calls it `bugs_open/137`; I repeated the
   tool's string for a day. The gate's message is stale, the bug moved.
3. **Reaching for the built tool is not the end of the discipline.** `probe-page-url.sh`
   exists because four sessions composed URLs by hand; I was occurrence five, six days
   after logging occurrence four myself.
4. **Inherited a denominator, dated it today.** "0 of 1,903" in a council submission:
   numerator mine, denominator four days old. A claim can carry two counts and one
   date, and the composite takes the fresher.
5. **Backticks in `git commit -m` execute.** Cost a word out of a commit message that
   cannot be amended. Use `-F -` with a heredoc.
6. **Probe hazards inherited from the companion lane** (their §9): an expired token
   makes EVERY symbol read absent — so a must-be-present control is mandatory; and
   BusyBox `grep` over `/proc/1/exe` needs 100–120s per grep, where a timeout kill is
   indistinguishable from a negative. Run them singly.

All of the above are in `WRONG_CALLS.md` with their cheap checks.

## 10. What NOT to do

- **Do not submit another council sketch.** Round 4 is real diffs or nothing (§2).
- **Do not re-derive migration 643.** It is void (§3).
- Do not seed any `composite` row or `parent_instance_id` before the walk branch
  ships (035 §9 r1).
- Do not read `recomposeAncestors`' absence from the binary as a missing commit (§1).
- Do not bolt RFC_046's primitive fix into P1 — 357's to route.
- Do not answer 035 §6.1's completeness floor by lowering `prune_floor_ratio`.
- Do not treat the guides as a P2 consumer without NEW evidence — their answer was
  scoped to eleven guides on one site, judged from rendered structure.
- **Do not re-open the routing arm inside P1** — ruled into P2 on 2026-09-03 (§3).
- **Do not commission an arrangement-scoring loop** — ruled against on 2026-09-03
  (§6.1). The alternative stays open and costs no migration if it is ever wanted.
- **Do not build anything to give grain A defaults or forking** — it already has both
  (§7.2). Building it again is the drift class D3 retired.
