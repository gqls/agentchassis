# 035 FEATURE — component hierarchy: composed components, decomposed generation

**Raised:** 2026-08-15 (owner steer, recorded verbatim in
`docs024_key_docs_latest/inline_guide_imagery/PLAN_2026-08-14_durable_inline_guide_imagery.md`
§9); this slot reserved the same day.

**Written:** 2026-08-22, by Fable, in-session — the fifth attempt. The first four
dispatches died on Fable capacity limits (`news_editorial_features` NOTES,
MISSTEP 3); the owner twice specified that Fable writes this plan and no model was
substituted. Every figure below marked `[MEASURED]` was re-verified against the
live DB or the tree **by this session on 2026-08-22**, not carried forward.

**Status: DESIGN — nothing built.** No DB writes, no code edits ship with this
file. Owning lane for execution: `editorial_design_uplift` (its Phase F), in
coordination with the lanes named in §8.

---

## 1. The brief — three owner steers, quoted

**2026-08-15** (the founding steer):

> *"In the original architecture components were composed of smaller components ad
> infinitum. We could use this pattern so that components can be designed and
> shuffled as we like. The experience loop and other guide agents like the vigilant
> designer and offer and benefit analysis agents might be able to help determine
> the best place for the images as well."*

**2026-08-20:** design uplift is its own workstream, and the composition plan is
to be written by **Fable** specifically (reaffirmed: *"We want to be using Fable
for the plan"*).

**2026-08-22** (this session — a new steer, and it sharpens the whole brief):

> *"I don't like the interleaved content and imagery being in one llm call, I'd
> like to decompose and have more control and consistency over that and more
> control over versions and design variations of the same."*

The 08-22 steer matters because it settles what composition is **for**. It is not
primarily a layout mechanism — section alternation already interleaves at a coarse
grain, live on four editorial pages. It is a **control** mechanism: one LLM call
per unit instead of one call owning a whole interleaved region; regeneration
scoped to the unit; versions and design variants addressable per unit. Layout
flexibility falls out of that; it is not the goal.

## 2. What this must deliver — five goals, each testable

- **G1 — Decomposed generation.** No interleaved region is written in one LLM
  call. Each child (prose block, figure, chart, pull-quote, timeline) is its own
  generation unit with its own guidance, its own truncation check, its own review.
  *Test: rewriting one prose child leaves every sibling row byte-identical — by
  construction (row-scoped write), not by prompt discipline.*
- **G2 — Composition.** A component instance can contain child instances;
  children are addressable, reorderable, reusable ("designed and shuffled as we
  like"). *Test: reordering two children is a `position` update + rerender —
  no LLM call, no content change.*
- **G3 — Version control.** A child instance can pin the template version it
  renders with; content history is recoverable per child. *Test: a template edit
  to a component definition does not change a pinned instance's rendered bytes
  until the pin is deliberately moved.*
- **G4 — Design variants of the same content.** The same child `content_data`
  can render under a different, contract-compatible template. *Test: a variant
  swap changes `rendered_html` and leaves `content_data` byte-identical; a
  contract-incompatible swap is refused, nothing written.*
- **G5 — Consistency.** Children of one parent are generated against one shared
  brief (voice, premise, terminology) and styled by the site's tokens, so
  decomposition does not fragment the voice. *Test: the brief is one stored
  object every child call cites; it is not re-derived per child.*

### CANDIDATE G6 — RECORDED, NOT DESIGNED, AND DELIBERATELY NOT WRITTEN HERE

**Owner steer, 2026-08-26, quoted verbatim** (relayed by the `agentchassis-51`
session, which flagged it rather than acting on it):

> *"There is also a mechanism to have components contain other components that we
> should use and we could **store patterns of such combinations after they've been
> through component and experience loops** etc."*

**This is NOT covered by G1–G5, and that reading was checked rather than taken on
trust.** G2 makes a *child* addressable, reorderable and reusable; G3 pins a
*component's* template version; G4 swaps a contract-compatible variant of *one
component*. All three operate on a component. The steer is a level up: an
**arrangement** — prose → figure-left → pull-quote → time-series → prose — stored
as a unit, carrying the verdict of the loops it passed, and applied to a new page
whole.

**⚠ THE LOAD-BEARING CLAUSE IS "after they've been through … loops", AND THE
PROVENANCE IT ASSUMES DOES NOT EXIST AT THAT GRAIN. `[MEASURED 2026-08-26]`**
Whoever writes G6 should start from this, because it is the half that reads as
already-solved and is not:

- **Component side.** `compute_component_quality` is live and has run — **126 of
  381** `content_components` carry a `quality_score`, last check 2026-08-24. But
  **zero of this feature's families do**: `evidence-timeseries` and
  `mechanism-flow` both hold NULL `quality_score` and NULL `quality_checked_at`,
  and `insight-article`, `prose-block`, `editorial-pullquote` and
  `evidence-figure` **do not exist as component rows at all** — they are D3's
  planned families, not built ones. (This also refines the news_editorial handoff
  §3 line "A2 — `compute_component_quality`, still never run": fleet-wide it has;
  on this lane's components it has not.)
- **Experience side.** The council is real and has ruled — **80 notes across 13
  distinct subjects**, 2026-07-18 → 2026-08-15. But every subject is a FEATURE,
  an experience or a design decision (`site-chat-intake`, `tool-patent-check`,
  `idea.uk-site`, `D-001-free-beside-paid` …). **None is a component
  arrangement, and none is editorial.**

**So G6 is not merely "add storage, identity and an apply-path" to existing
provenance.** Both loops exist and both score a *different object* than the one
the steer wants stored. A stored pattern would need a verdict-bearing object at a
grain neither loop currently emits — which is a design question, not an
implementation detail, and it is the first thing to settle.

**Why this section stops here.** The owner has asked for Fable on this design
three times and 035 arrived on the fifth dispatch after four capacity failures. A
G6 written by a passing lane — or by the lane merely implementing P1 — would read
as authoritative and would not be. So this records the steer and the measurement
and routes it; it does not design it. **Nothing below §2 has been changed to
accommodate it, and no phase has been re-scoped.**

**Interfaces it would touch, noted for whoever does write it:** D1 (meaningless
without parent/child representation), D7 (its apply-path is propose→apply→approve
with a *stored* spec rather than a freshly generated one), and §8's handover to
the inline-imagery lane — whose pressure is live again as of 2026-08-26 (the
owner's motivating complaint was "not enough images on any of the pages").

## 3. What exists today — measured 2026-08-22

Composition has been **designed three times and exercised zero times**, and the
version seam is half-live. The fleet grew since the 08-20 census (1,580 → 1,903
instance rows) and adoption did not move:

| mechanism | state `[MEASURED 2026-08-22]` |
|---|---|
| `page_components.parent_instance_id` | column + FK + index (`idx_page_components_parent`) exist; **0 of 1,903 rows** set it; **zero Go references** (grep across `platform/`, `internal/`, `pkg/`) |
| FK delete semantics | `page_components_parent_instance_id_fkey` has **no ON DELETE action** → deleting a parent that has children **errors** (fail-loud, not orphan). Load-bearing for §6.1 |
| `content_components.render_mode='composite'` | **0 rows**; `deriveRenderMode` (`store_generated_component_action.go:1481-1506`) emits only `agent`\|`template`, so composite is unreachable by construction — and it runs on **both** INSERT and UPDATE paths (:561, :639), so a hand-seeded `composite` would be **silently reverted on the next regeneration**. Wiring the third value is mandatory, not optional |
| `content_components.child_components` (jsonb) | **0 non-empty rows**; nothing reads it |
| `component_versions` | **live and written** — 363 rows; real producers (`scope_component_instance` 98, `036_component_hygiene` 51, `repair_instance_scope_bindings` 27, `component_selector` 16, …). But it is write-only history: `page_components.component_version_id` is **0 rows** — no instance pins a version and no render path reads one |
| assembly | flat concatenation over `componentNames` in render order (`assemble_from_library.go` `assembleComponents`), threading the bugs-283 per-page instance counter (`NewInstanceCounter`/`BindInstanceToken`) |
| template executor | `executeGoTemplate` (`call_agent.go:1170-1220`): `text/template`, `missingkey=zero`, six-function funcmap (`default, eq, ne, lower, upper, isset, safe`), **no `{{template}}`/partial support**; a missing function is a PARSE error (VIZ-007) |
| the save hazard | `save_page_sections_action.go:823` DELETEs a page's **agent-writable** rows page-wide; agent-writable = `locked_at IS NULL OR (timed AND expired)` (`datahelpers.AgentWritableSQLFor`, `chrome_render_inputs.go:91-95`) |
| per-row control surface | locks (`lock_type` permanent/timed/review), digest stamps, the v1.0.1276 archive trigger (`page_component_history`), claims scanning, the 285 dedup index — **all keyed on individual `page_components` rows** |

The last row is the design's centre of gravity: **everything this estate knows how
to do to a component — lock it, version it, archive it, scan it, audit it — it
does per row.** The reason interleaved content is uncontrollable today is that a
whole article is ONE row with one llm blob. Make each unit a row and the entire
existing control surface applies to it with no new machinery.

## 4. The design

### D1 — Representation: children are `page_components` rows with `parent_instance_id` set

The dormant column is adopted as-is: a child is an ordinary instance row —
own `content_data`, own `rendered_html`, own lock, own history, own digest —
whose `parent_instance_id` names its parent row on the same page. `position`
orders siblings **within** the parent (the column already exists and is already
per-row). This is the first live use of the column; §7 takes the scope call that
follows from that.

Rejected alternatives:
- *Markup inside a blob* — the failure being retired (`bugs_open/238` class; the
  inline-imagery plan's whole §2).
- *Children as a jsonb list on the instance* — recreates the blob one level up:
  no per-child locks, history, or scanning; invisible to every existing tool.
- *Children as their own page sections* — `pages.sections` is the page's flat
  spine, materialised from `site_plan_sections`, with a documented
  three-places/resurrection landmine family. Composition happens **below** the
  section grain precisely so that spine is untouched (D2).

### D2 — Grain: composition lives below the section, invisible to the page spine

A composed section occupies exactly one entry in `pages.sections` /
`site_plan_sections` — the parent. Children exist only in `page_components`.
Nothing about section planning, nav, or the sections cache learns about
composition. This is what keeps the blast radius honest: to every existing
consumer of the page spine, a composed section **is** one section.

Child `slot_name` convention: `<parent_slot>.<child_key>` (e.g.
`insight-article.fig-installations`) — unique per page, satisfying
`uq_page_components_no_byte_identical_duplicate` (which keys on `page_id` +
`slot_name` + hashes + `component_id`), greppable, and self-describing in the
`page_component_history` archive rows.

### D3 — Definition side: a composite declares SLOTS, and `deriveRenderMode` learns the third value

A parent's `content_components.input_schema` gains a `slots` block beside
`fields` (additive; nothing reads unknown schema keys today):

```json
{ "fields": { "standfirst": {"source": "llm", "required": false} },
  "slots":  [ {"key": "lead",  "function": "prose-block",      "required": true,  "count": "1"},
              {"key": "fig-*", "function": "evidence-figure",  "required": false, "count": "0..3"},
              {"key": "quote", "function": "editorial-pullquote","required": false, "count": "0..1"} ] }
```

`child_components` (the dormant jsonb column) is **retired from the design
rather than adopted**: slot declarations belong in `input_schema` where the
generation pipeline already looks, and two homes for the same declaration is the
drift class this repo keeps paying for. The column stays unused; the register
entry says so.

`deriveRenderMode` derives `composite` when a schema declares non-empty `slots`
— schema-driven exactly as `agent`/`template` are today, on both its call paths,
so the value survives regeneration. The `check_render_mode` conditional in the
page-content-writer workflow must learn to route `composite` (to the composition
build, D5) **in the same change** — a third value its switch has never seen is
otherwise an unhandled arm; this is the one point where the change is visible to
an existing workflow and it goes to the council with that stated (§7).

### D4 — Rendering: a bounded, fail-closed, Go-side walk; templates see pre-rendered slots

The parent template references `{{.slots.lead}}`, `{{.slots.fig_1}}` — plain
fields, resolved by Go, holding each child's **already-rendered** HTML. The walk:

1. Load the page's rows once; group children by `parent_instance_id`.
2. Depth-first, children in `position` order: render leaves exactly as today
   (template + `content_data`, same executor, same funcmap), then render the
   parent with its `slots` map injected.
3. **Depth cap 3, cycle refusal** — the FK cannot forbid a cycle (a row can
   reference a sibling or itself); the walk detects and **fails the render**
   (bugs-260 idiom: fail rather than stitch), never truncates silently.
   > **Refined by P0 (2026-08-22): the load-bearing cycle guard is a
   > COMPLETENESS assertion, not a path-set.** With one parent pointer per row,
   > a reachable cycle cannot exist (a top-level row has no parent), so every
   > cycle manifests as rows the walk never reaches — and a walk guarded only
   > by a path-set would DROP them silently, the exact omission this rule
   > forbids. The walk must assert every row rendered exactly once and fail
   > naming the unrendered rows. Proven by mutation: removing the assertion
   > let a mutual cycle render "successfully" minus its rows.
4. A `required` slot with no child, or a child whose render fails → parent
   render **fails**. An optional absent slot renders as empty string
   (`missingkey=zero` already gives this for free).
5. The bugs-283 **instance counter threads through the walk in final render
   order** — children consume tokens from the same per-page counter, so
   per-instance element ids stay canonical.
6. ~~Both render paths get the walk: assembly (`assembleComponents`) and the
   sections rerender path.~~ One walk implementation, several callers — not
   several walks.
   > **CORRECTED 2026-08-26 (P1 implementation) — `assembleComponents` CANNOT
   > CARRY COMPOSITION, and putting the walk there would have been dead code.**
   > It renders from a list of component FUNCTION NAMES fetched out of the
   > library (`GetComponentWithFallback` → `content_components`), and
   > `assemble_from_library.go` contains **zero** references to
   > `page_components` — so no row on that path can have a
   > `parent_instance_id` to walk. Measured 2026-08-26: of the files that both
   > read `page_components` AND call `RenderTemplate`, there are **four**
   > (`v3_site_actions.go`, `section_editor_actions.go`,
   > `rerender_page_sections_action.go`, `rerender_pages_actions.go`), and
   > `assemble_from_library.go` is not among them.
   > **Exactly ONE of the four is a page-wide walk over a page's rows:
   > `rerender_page_sections_action.go`.** The others are single-target:
   > `RenderComponentAction` (`v3_site_actions.go:2459`) renders one component,
   > `apply_section_edit` (`section_editor_actions.go:1117,:1284`) renders one
   > section, and `rerender_pages_actions.go:538` renders the **head** (chrome,
   > not sections).
   > **So P1's real scope is the page-wide path plus a decision about the
   > single-target ones** — and that decision is not optional: a parent's stored
   > `rendered_html` has its children embedded, so re-rendering a parent
   > ALONE resolves `{{.slots.*}}` to empty and silently blanks its children,
   > while re-rendering a CHILD alone leaves the parent still serving the old
   > child bytes. Both are content-loss shapes and both need an answer in P1,
   > not P5. Had the walk gone where D4.6 said, it would have satisfied §6.8's
   > own warning about dormant mechanisms while changing nothing.
   >
   > **DECIDED 2026-08-26 — REFUSE direction 1, RECOMPOSE direction 2.** Settled
   > on production evidence contributed by the `bugs_open/384` lane
   > (`agentchassis-51`), which spent two days in this exact failure mode, and on
   > an ASYMMETRY rather than on caution:
   > - **Direction 2 (child edited, parent serves stale bytes) is the one that
   >   reaches production.** 384 is this defect in another vocabulary: the child
   >   object was correct — derived, linked, `active`, byte-perfect — and the
   >   parent listing served old bytes for **four days**. It survived because
   >   *everything you would check said fine*: the child row right, the parent
   >   re-rendered three times, work items `complete`, no error, no empty region,
   >   no visual gap. **Old is invisible.**
   > - **Direction 1 (parent rendered alone, `{{.slots.*}}` → empty) empties
   >   VISIBLY.** It is silent in the logs but loud in the artefact, and this
   >   estate already has instruments that notice a page losing a region
   >   (`empty_section`). It is the more destructive failure and the more
   >   findable one.
   > - **So refusing direction 2 would be the expensive mistake.**
   >   `apply_section_edit` is the live, load-bearing edit path — it is what
   >   delivered this lane's instance-scope change on 2026-08-25 — and a
   >   fail-closed refusal there converts "the parent goes stale" into "the
   >   editor stops working", a regression that would have to be un-shipped.
   >   Refusing direction 1 costs nothing anyone does today: **nothing renders a
   >   composite parent alone.**
   >
   > **THE STALENESS TELL IS A DIGEST, NOT A TIMESTAMP — and the column already
   > exists and is already maintained. `[MEASURED 2026-08-26]`**
   > `page_components.rendered_html_digest` is **md5**, written in the SAME
   > STATEMENT as the bytes (`bugs_open/229`; `save_page_sections_action.go:1066`,
   > `section_editor_actions.go:1569`), so it cannot drift from what it describes.
   > **1,935 of 1,948 stamped rows (99.3%) match `md5(rendered_html)` right now**,
   > including both rows this lane rewrote yesterday. So "the parent is stale"
   > is a pure join — the parent's stored HTML must embed each child's CURRENT
   > digest — with no render, no HTTP and no comparison re-render.
   > ⚠ **Two things any such sweep owes.** (1) **288 rows carry `rendered_html`
   > and NO digest** (12.9%) — that is the sweep's blind spot and it must be
   > counted, not silently excluded. (2) **A demand control up front**: a
   > stale-count of zero is worth nothing until the query is shown to have
   > counted something, which is how 384's own sweep reports
   > (`consumer_pages: 25, stale: 0, current: 25`).
   > **Do NOT use `updated_at` ordering** ("any child newer than its parent").
   > It is wrong in both directions: a no-op child re-render bumps the timestamp
   > without changing bytes (false positive), and a shared-tree write can reorder
   > them (false negative). 384's own verification protocol tripped on exactly
   > that and was corrected the same day.

**No `{{template}}` support is added to the executor.** VIZ-007's constraint
(missing function = parse error; no arithmetic) stays exactly as-is; composition
is data-shape work, not template-language work. This is the cheapest correct
answer and it keeps the funcmap's whole verification story intact.

The parent's final `rendered_html` is stored complete (children embedded), so
every downstream consumer of stored/served HTML — deploy, audits, the claims
gate at page grain — sees what it sees today: one section, one HTML string.
Child rows additionally carry their own `rendered_html`, which is what makes
per-child scanning, diffing and history work (§6.4 handles the double-scan).

### D5 — Generation: one call per llm child, one shared brief for all of them

This is the owner's 08-22 steer made mechanical:

- The composite build resolves the parent's slot plan (which children exist —
  from the section plan, D8's per-family defaults, or an approved composition
  spec, D7), then runs **one generation step per llm-mode child**, each writing
  its own row. Non-llm children (figures via the IMG-056 resolver, charts from
  registered facts) never see a writer at all.
- **The shared brief is a stored object, not a prompt habit**: the parent row
  carries `content_brief` (column exists, today used by other flows) holding
  premise, voice, terminology, and the list of sibling slots with one-line
  summaries. Every child call receives it verbatim. Consistency comes from the
  one brief, scope from the one slot — the exact inversion of today, where scope
  and consistency both live in one blob-sized call.
- Per-child truncation and structure checks at generation time
  (`output_tokens == max_tokens` = CUT, `bugs_open/012` class) — decomposition
  makes each call smaller, which shrinks that risk, but the check is per child
  regardless.
- **Child results are DB-mediated, never return-value-mediated**: each step
  persists its row server-side; the parent walk reads rows. `bugs_open/274`
  (≈15k completed children whose results never reached parents) makes any
  return-payload chaining a known trap.
- Regeneration of one child = one work item naming one row. The page-wide writer
  sweep never owns a composed region's children (§6.1).

### D6 — Versions and design variants: make the two half-built seams meet

**Versions (G3).** `component_versions` already snapshots every template edit
(363 rows, live producers); ~~`page_components.component_version_id` already
exists to pin one, and nothing reads it. The design: the render walk resolves a
child's template as *pinned version if `component_version_id` is set, else
current*.~~

> **CORRECTED 2026-08-22, hours after writing — the pin must NOT be
> `component_version_id`.** Pre-flight for P1 found the `bugfix_357` lane's
> uncommitted work implementing **RFC_046 (RULED 2026-08-22, option 1: stamp
> identity at the point of production)**: that column becomes a universal
> **provenance stamp** — *which version PRODUCED these bytes* — written by
> renders and deliberately carried through rerenders. Under that ruling every
> rendered row is non-NULL, so "NULL = follow current" stops being expressible
> and a pin read of the same column would silently freeze every instance at
> whatever last rendered it. **Intent and record must be different fields,
> because the record overwrites the intent.** The pin is therefore its own
> opt-in declaration: a new nullable
> `page_components.pinned_component_version_id` (FK `component_versions`,
> default NULL = follow the library), read by the walk; RFC_046's stamp then
> records what was actually used — which makes "was the pin honoured?" a
> mechanical equality (`stamp == pin` wherever the pin is set), so the two
> mechanisms verify each other instead of colliding. RFC_022 shape: opt-in,
> unsafe-default-OFF, zero consumers at birth. The 357 lane was told the same
> day (their NOTES, dated append).

That single read makes the whole seam real — pin = control, unpin = follow the
library. Same register obligation as the composition columns.

**Variants (G4).** A design variant is a **sibling `content_components` row**
(`forked_from` provenance, already in the schema): same declared field contract,
different `html_template`/CSS. A variant swap updates the child's
`component_id` and touches nothing else. The swap is guarded mechanically:
refuse unless the variant's schema fields are a superset of the fields the
instance's `content_data` actually uses (`schemaFieldSet` at
`store_generated_component_action.go:1511` is the existing parse to reuse) —
fail closed, nothing written on refusal. `missingkey=zero` means a sloppy swap
would render silently-empty fields, which is exactly the silent failure the
guard exists to prevent.

Content rollback needs no new machinery: `page_component_history` (archive
trigger, v1.0.1276) already holds every destroyed state per row, and children
are rows.

### D7 — Agents propose, the pipeline applies, a human approves

The 08-15 steer's second half (design/experience agents choosing placement),
bounded by design D (`bugs_open/126` — a generated artefact enters **human
review, never auto-repair**):

- A design agent (`design-audit-agent`, `visual-designer`, the experience loop)
  emits a **composition spec** — ordered slot list, variant choice per child,
  optionally a proposed new child — as a work item into human review.
- On approval, the composition pipeline applies it as row operations
  (`position` updates, variant swaps, child inserts) and rerenders. **No agent
  writes `rendered_html` or reorders rows directly.**
- The spec vocabulary is deliberately the same shape as D3's slot declarations,
  so "what the designer proposed" and "what the page is" diff mechanically.

This phase inherits 274's delivery caveat (D5) and stays behind everything else
in §5 — it is the steer's destination, not its first step.

### D8 — First target: the editorial insights pages

The editorial family is the proving ground **because it is the estate's safest
sandbox**: pages are `rebuild_policy='owned'`, rows are permanent-locked, no
writer sweep touches them, and the owning lane (`editorial_design_uplift`) also
owns this plan's execution. Blast radius of the first composed page: that page.
The guides (`article-body`, 93 instances / 18 sites `[MEASURED 2026-08-14]` by
the inline lane) come **last**, in coordination with that lane (§8).

## 5. Phasing — each phase carries its falsifier

**P0 — prove the walk cold (no cluster writes).** Extend the local render
harness (`scratchpad/build/render.go`) with the D4 walk over fixture rows:
depth-1 compose, depth cap, cycle refusal, required-slot failure, optional-slot
absence, instance-counter threading. *Falsifier: any of those six behaviours
needs executor or funcmap changes — if so, D4 is wrong, stop and redesign.*

> **P0 DONE 2026-08-22, same session — the falsifier did not fire.** Harness at
> `docs024_key_docs_latest/editorial_design_uplift/harness/composewalk/` (own
> `go.mod`, outside the platform build; durable, not scratchpad). Eight checks
> (the six above + flat-page byte-identity for the opt-out claim in §7, +
> non-identifier child-key refusal), all PASS against a byte-faithful replica of
> the executor with the funcmap untouched. The checks discriminate: two walk
> mutations (post-order token binding; completeness assertion removed) each
> failed exactly the check aimed at them (6 and 3). One design refinement fed
> back into D4.3 above. Next: P1.

**P1 — the read path, live, one page (council-gated).** Walk in both render
paths + `deriveRenderMode` third value + `check_render_mode` routing arm +
register entry, one commit. Recompose ONE live insights page as parent +
children carrying its current content decomposed. *Acceptance: served page
byte-equivalent to the pre-composition render (modulo instance tokens);
`pages.sections` diff EMPTY; then the class test the whole feature exists for —
fire a rewrite at one prose child and prove every sibling row byte-identical.
Falsifier: the save/rerender paths cannot leave siblings untouched without
edits to `save_page_sections` — that is §6.1's boundary; scope P1 to owned
pages and say so.*

> **P1 PROGRESS, 2026-09-03 — DIRECTION 2 IS WIRED AND LIVE-CAPABLE; THE READ PATH IS NOT.**
> Commit `1007be27d`, council `cab931b1-8b45-461e-8a37-0dbdfa6aa928` (`Council-Submitted:`, verdict
> owed a read). State of the seven P1 deliverables, so the next session does not re-audit:
>
> | deliverable | state |
> |---|---|
> | `deriveRenderMode` third value (`composite`) | **DONE** `1f745e730` |
> | membership helpers (`hierarchyChildrenOf` / `hierarchyAncestorChain` / `hierarchyChildKey`) | **DONE** `bc8167100`; reached in production |
> | direction 1 — refuse to render a composition parent alone | **DONE** `028c3e112` |
> | flat-pass extraction (`rerenderFlatSections`, `classifyStoredSection`, `renderPlannedSection`) | **DONE** `2a0bdb001`, `94f81cc60`, `22ed53ee7` |
> | direction 2 — `recomposeAncestors`, **called** | **DONE 2026-09-03** `1007be27d` |
> | `check_render_mode` routing arm | **REFUTED, not deferred** — `5542a76d6`: nothing reads `render_mode`. P1's routing story cannot work as this document wrote it; routing a composite needs a consumer for that column or a different signal, and that is an open design question |
> | the walk in both render paths + §6.9's filter + register entry + live canary | **NOT DONE — this is what remains of P1** |
>
> **The finding worth carrying, because three rounds of review did not reach it.**
> `recomposeAncestors` shipped 08-31 with a `tx *sql.Tx` parameter whose necessity its own header
> asserted in capitals (*"THE db/tx SPLIT IS FORCED"*), reasoning about reads that must see "the
> uncommitted edit" inside `apply_section_edit`'s transaction. `[MEASURED 2026-09-03]`
> **there is no transaction** — `grep -nE 'BeginTx|\.Begin\(|Commit\(\)|Rollback\(\)'
> section_editor_actions.go` returns nothing; it persists on the autocommit connection. So no call
> could compile, the function sat uncalled for three days and the linker dropped it. **A comment can
> assert a fact about its caller and nothing type-checks it**; the plan reviews read the plan, and
> only compiling a call reads the caller. Attempting the wiring also found three defects in the
> ancestor write that no round had reached — it carried neither the tombstone nor the lock predicate
> its sibling writes carry, it read zero-rows-affected as success, and it was unstamped though it
> writes an archived column. All fixed, all pinned by mutation-proved tests (M1–M4, each killed).
> Full account: `LANDMINES.md#a-function-can-carry-a-parameter-its-only-caller-cannot-supply…`,
> `WRONG_CALLS.md` 2026-09-03, `editorial_design_uplift/NOTES` 2026-09-03.
>
> **Still inert, and re-measured rather than carried forward** (the dated-count rule): **0 of 3,229**
> `page_components` rows carry a `parent_instance_id` `[MEASURED 2026-09-03]`, against 0 of 2,249 on
> 08-31 and 0 of 2,005 on 08-24 — the table grew ~1,000 rows in ten days and the parented count
> stayed at zero. The cost direction 2 adds to the live edit path is one indexed SELECT per edit.
>
> ⚠ **The COMPONENT-side count in this document was stale, and the council's `prior_art_librarian`
> seat is what forced the re-measurement** (corr `cab931b1`, medium: the inertness argument "rests on
> two DB counts asserted as `[MEASURED]`"). `content_components` holds **554** rows
> `[MEASURED 2026-09-03]`, not the **386** quoted here and in this lane's docs since 08-31. The claim
> survives and is larger: **0 of 554** declare a `slots` key in `input_schema`, **0 of 554** carry
> `render_mode='composite'`. Stale by ADDITION — the count did not go wrong, it went *out of date
> upward*, which is precisely why the dated-count rule exists. **Anyone quoting "386" is quoting
> 08-31.**
>
> **⚠ The next session's first obligation is §6.9's filter, and it belongs INSIDE the read-path
> change, not after it.** `loadStoredSections` still has no `parent_instance_id` filter.

**P2 — decomposed generation (council-gated).** D5's fan-out for one composite
family (`insight-article`: lead prose + figure + chart + pull-quote), shared
brief in `content_brief`. First NEW editorial feature built this way instead of
via hand-seeded sections. *Falsifier: per-child calls lose cross-child
coherence a single call had — the brief is the mechanism on trial; judge it on
one real feature before widening.*

**P3 — versions + variants (council-gated).** `component_version_id` read in
the walk; variant-swap guard; one real variant (e.g. a second
`evidence-figure` treatment) swapped on one child and back. *Falsifier: a pin
surviving regeneration proves G3; if regeneration clears or ignores the pin,
the seam is still write-only and the phase is not done.*

**P4 — agent-proposed composition (D7).** One `design-audit-agent` finding
turned into a composition spec, through human review, applied by the pipeline.
*Falsifier: if the spec cannot be applied as pure row operations, D7's
vocabulary is wrong.*

**P5 — the guides, with the inline-imagery lane.** Their plan-as-truth phases
remain the guides' near-term mechanism (§8); P5 is the deliberate migration of
`article-body` pages to composition, superseding the blob per page, only after
P1–P3 have held on editorial pages for real weeks. The un-owned-page question
(§6.1) is settled here or the migration does not start.

## 6. Hazards — every one named, none new to this repo

1. **The page-wide save DELETE vs child rows.** `save_page_sections_action.go:823`
   deletes agent-writable rows page-wide; `bugs_open/178` is a standing stop
   sign against teaching `save_page_sections` new floors. P1/P2 live entirely on
   owned/locked pages, where rows are not agent-writable and the DELETE cannot
   take them — **and the FK's missing ON DELETE action means a sweep that tried
   would ERROR loudly, not orphan children silently.** That error is the
   tripwire, not a bug to "fix" with a CASCADE: cascading would hand the sweep
   exactly the silent-destruction power 178 exists to deny. The general
   unlocked-page answer is P5's opening question, not P1's assumption.
   > **Added 2026-08-25 (357 lane), and it sharpens what P5 must answer:** the
   > same file carries a **completeness floor** (`save_sections_prune_floor.go`)
   > that refuses the WHOLE save when the incoming row set is too small a
   > fraction of two independently-measured cohorts — the existing rows, and
   > `pages.sections`' **planned** count. Composition changes the arithmetic of
   > both, because a composed page's `page_components` holds parents **plus
   > children** while `pages.sections` still lists only the parents (D2/§9). So
   > a later NON-composition-aware save of a composed page offers parents-only
   > against an existing set of parents+children, and the ratio falls — the
   > floor then refuses the entire save. That is fail-closed and arguably
   > correct, but it presents as a `save_refused_incomplete` work item nobody
   > can interpret. **[MEASURED 2026-08-25 by the 357 lane]** 32 such items are
   > parked at `needs_human_review` across ~14 domains since 2026-07-31, from a
   > different cause (a one-section fragment against a 3+ section plan) — so
   > this floor demonstrably fires in production and is not theoretical.
   > **Moot for P1/P2** (owned/locked pages: the rows are not agent-writable and
   > the delete never runs). It is P5's question, and P5 must answer it as
   > "what is the denominator on a composed page?", not by lowering the ratio.
2. **Cycles and depth** — FK permits them; the walk refuses them (D4.3),
   fail-closed, tested in P0 with an induced cycle. A guard proven only by a
   quiet test is not proven (the mutate-to-prove rule).
3. **Instance tokens (bugs 283)** — children consume the same per-page counter
   in final render order or per-instance ids drift between paths. P0 asserts
   token equality between a composed and hand-flattened render of the same rows.
4. **Claims-gate double-scan** — child HTML exists in its own row AND embedded
   in the parent's. The gate scans at generation (per child, where refusal can
   still block persistence) and at page grain (served truth); a duplicate
   finding keyed on both rows is possible and the dedup key must fold
   `parent_instance_id` chains. Named for the P1 council round.
5. **SVG text is invisible to the claims gate** (VIZ-009) — unchanged by
   composition; figure children carrying words keep them in HTML text.
6. **`deriveRenderMode` reverts hand-seeded values on regeneration** (§3) — no
   composite row may be seeded before the derivation ships; ordering is
   image-before-config, the platform's standing rule.
7. **274** — any chain that needs a child *result* uses work items / DB rows,
   never return payloads, until 274 is fixed.
8. **A dormant mechanism stays dormant without a driver** — this estate's most
   familiar failure (three dormant composition mechanisms in §3 are the proof).
   P1 does not merely wire the walk; it ships a live composed page the same
   week, and the register entry records "exercised on `<page>`" with the date,
   not "deployed".
9. **One `RenderContext` reused across the walk silently forges RFC_046
   provenance** — added 2026-08-25 by the news_editorial lane while re-reading
   the seams for P1. `RenderTemplate` does not RETURN the digest; it **mutates
   the context**: `ctx.RenderedTemplateSHA = hex.EncodeToString(sum[:])`
   (`component_library.go:1081`), once per call. Two consequences for a walk,
   and the second is the nasty one:
   - **N renders, one field.** Each node's digest overwrites the last, so after
     rendering children then the parent the context holds only the PARENT's.
     Every child's stamp is gone unless captured immediately after its own call.
   - **An empty-template child inherits its PREDECESSOR's stamp.** The
     empty-template branch returns early and *deliberately does not set* the
     field (RFC_046: "empty means unknown … a provenance token pointing at a
     template that renders nothing is worse than no token"). That reasoning
     assumes a FRESH context. On a reused one the outcome is worse than "no
     token" — it is **another template's token**, i.e. a false provenance claim
     of exactly the kind RFC_046 exists to prevent.

   **[VERIFIED 2026-08-25 — today's code is SAFE, and that is the trap.]**
   ~~Both~~ **FOUR (corrected same day — the 357 lane, who own the stamp, named a
   reader my census missed; a count stale by ADDITION within the hour, which is
   the failure this repo's dated-count rule exists for)** live readers of the
   field render exactly ONE template per context, so nothing
   is wrong now: `rerender_page_sections_action.go` allocates a fresh
   `rc := &RenderContext{…}` at **:632, inside** the per-section loop opened at
   :473, renders at :661 and reads at :759; `v3_site_actions.go` renders at
   :2459 and reads at :2553 with **no loop between them**; and
   `adopt_fragment_section.go` allocates `rc := &RenderContext{…}` at :122,
   renders once and reads at :144 — so **phase 2's own adoption path is
   unaffected too** (contributed by the 357 lane, 2026-08-25).
   (`assemble_from_library.go:303`
   *does* reuse one context across its component loop, mutating it per iteration
   — but it discards all three reports and nothing on that path reads the SHA
   back, so it is unaffected today.) **The walk is the first code on the estate
   that would render many templates per context**, which is why no existing test
   or reader would catch this and why it must be decided before the walk is
   written, not after.

   **The four-reader version of this is stronger than the two-reader one, and it
   is the whole argument:** four independent files, in four lanes, all happen to
   allocate a fresh context per render — and *nothing made them*. It is a
   property held by convention across a corpus, not an invariant, which is
   precisely the shape that breaks the first time someone writes the fifth. The
   357 lane agrees the primitive-level fix (clear the field on entry to
   `RenderTemplate`, so "unknown" is guaranteed rather than coincidental) is
   correct **and is architecture-scope, not a bug patch** — by the 2026-07-29 §1
   test it changes what the shared mechanism GUARANTEES, from conditional-on-
   caller-structure to unconditional. It is theirs to route, and they want the
   falsifier below attached to it rather than shipped after it. **Do not bolt it
   into the P1 commit.**

   **The rule for P1:** ~~capture the digest immediately after each node's
   `RenderTemplate` call, or~~ **give each node its own context.** Do NOT read
   `ctx.RenderedTemplateSHA` after the walk and attribute it to anything.
   > **CORRECTED 2026-08-26 by running the falsifier this section asked for —
   > the FIRST of the two remedies offered above does NOT work, and it is the
   > one a reader would pick as cheaper.** "Capture immediately after the call"
   > is insufficient precisely in the case that motivates the hazard: the
   > empty-template branch `return`s **before** `ctx.RenderedTemplateSHA` is
   > assigned, so on a shared context an immediate read still observes the
   > PREVIOUS node's digest. Immediacy was never the variable — assignment was.
   > `[MEASURED 2026-08-26]` `TestHierarchyWalkEmptyTemplateChildStamp` runs
   > both arms against the real `RenderTemplate`: with a fresh context per node
   > the empty-template child's stamp is empty (correct); with a shared context
   > read immediately after every call it comes back **byte-identical to its
   > sibling's digest** (`3197aee8…`) — a false provenance claim.
   > Only a fresh context per node (or the 357 lane's primitive fix, clearing
   > the field on entry) is sufficient. The shared-context arm is kept in the
   > test as the discriminating control, and `t.Skip`s itself with an
   > explanation once that primitive fix lands, so it reports rather than
   > rots. **The walk enforces this structurally rather than by comment**: its
   > renderer callback must RETURN each node's stamp
   > (`hierarchyRenderedNode.Stamp`), so a caller cannot accidentally read one
   > shared field at the end — the owner's 2026-08-02 §2 ruling applied to this
   > seam.
   Falsifier for P0/P1: render a two-child parent where **one child's template
   is the empty string**, and assert that child's stored stamp is EMPTY — not
   its sibling's digest. A test that only renders non-empty children passes
   whether or not this is handled.

6.9 **`loadStoredSections` HAS NO `parent_instance_id` FILTER, and the flat
   occurrence counter is what per-section imagery binds on — so a nested walk
   silently mis-attaches figures.** Contributed by the `inline_guide_imagery`
   lane (IMG-075) and **verified first-hand here 2026-09-02**: the query selects
   `COALESCE(parent_instance_id::text, '')` but its WHERE is only
   `page_id = $1 AND <not removed>` — every row of the page comes back, flat, in
   `position, id` order. Today that is harmless, because **0 of 2,249 rows carry
   a `parent_instance_id`**, so the flat list and the top-level list are the same
   list.
   **The moment P1's walk renders children in a nested pass, they diverge.**
   `rerenderFlatSections` counts `sectionOccurrences.NextOccurrence(slotName)`
   over whatever `loadStoredSections` returned, and that occurrence is the
   `sectionRef` the per-section imagery map is keyed by
   (`plan_sections_action.go`, `sectionAssetFor`). If children stay in the flat
   list while the nested pass also renders them, every later section's occurrence
   index shifts, and figures attach to the WRONG sections — which renders,
   deploys and looks correct. It is the same failure the drift guard exists to
   stand down, arriving by a route the guard does not watch.
   **The fix is one line INSIDE the composition change** (the flat pass takes
   top-level rows only; the walk owns the children), and it is archaeology if
   found afterwards. *Falsifier for P1: compose one page, then assert the
   occurrence sequence the flat pass computes is unchanged from before
   composition — not merely that the page renders.*
   **Related, and settle it BEFORE writing any plan-vs-live comparison in P1**
   (contributed by the same lane, 2026-09-02): `check_section_source_drift`
   already compares the plan against `pages.sections`, but it views **both sides
   through `MergeLockedPageSlots`, so a locked row is NOT drift** — whereas an
   occurrence-based guard must stand down on exactly that input, because the
   ordinal indexes the plan and a locked insertion shifts every later occurrence.
   Same shape, different second operand, **inverted polarity**. Reusing the
   existing drift check for a composition-ordering guard would therefore be a
   correct-looking reuse that is wrong in the one case it exists for. Their
   second finding is the reason the key matters at all: `slot_name` is NOT the
   component name — the worked case is `loancalculator.co.uk/index`, plan
   ordering 1 `tool-loan-repayment` vs live position 2 `slot_name='tool-3'` with
   component function `tool-loan-repayment`, i.e. identical composition, different
   spelling. **Count rows and resolve the function; never match by name** (a hasty
   by-name match in that file is `bugs_closed/044`).
   ⚠ Note the ordering trap this creates with §6.x's other work: the guard the
   other lane relies on is **not yet in the running binary** (their round-2
   commit missed the 2026-09-02 12:28 roll), so a P1 that lands before their
   next roll has neither the filter nor the guard.

10. **Every composite will MIS-SCORE in `compute_component_quality` until its
    sync check learns D3's `slots` block — and it fails silently, as a low score
    rather than an error.** Identified by Fable while drafting G6, and **verified
    here by probe `[MEASURED 2026-09-02]`**: `extractTemplateVariables`
    (`compute_component_quality.go:352`) walks the parse tree and returns FIELD
    ROOTS, deduplicated. Fed `<article>{{.slots.lead}}{{.slots.quote}}</article>`
    it returns **`[slots]`** — one name, not two — and fed
    `{{.headline}}{{.slots.lead}}` it returns `[headline slots]`.
    So a composite declaring three slots advertises ONE template variable called
    `slots`, which no `fields` entry will match, because D3 puts slots in their own
    block beside `fields` rather than in it. The sync check then reads a template
    referencing "an undeclared field" and scores accordingly.
    **This is a P1/P2 obligation, not G6's** — it bites the first composite that
    exists, long before any pattern is promoted. It also matters more than a
    cosmetic score: G6's promotion gate is designed to require a non-NULL
    `quality_score` on every child family, so a scorer that mis-reads composites
    would gate on a number it computed wrongly.
    **The fix belongs with whoever teaches the checker about `slots`, and the
    falsifier is the probe above**: assert `extractTemplateVariables` yields the
    slot KEYS (or nothing at all for a slots reference), never the bare root.

## 7. Architecture-scope call — and the honest counterweight

Against RFC_022's three conditions, each measured 2026-08-22: composition is
**opt-in** (a row must set `parent_instance_id`; a schema must declare `slots`);
the **unsafe default is OFF** (absent both, assembly and rendering are
byte-identical — the walk's degenerate case IS today's flat loop); **zero live
consumers name it** (0 of 1,903 rows; 0 composite definitions; 0 Go readers).
By the 2026-08-11 ruling that shape is not automatically architecture-scope, and
by the 08-02 §2 ruling opt-in-default-OFF is the prescribed way to ship new
authority on a shared seam.

The honest counterweight, stated so the reviewer does not have to dig: this is
the **first live use of two dormant columns on the estate's most shared
surface**, it adds a third `render_mode` value that one workflow conditional
must learn, and the inline-imagery plan's §9 explicitly flagged first-live-use
as a stronger-than-usual review candidate. So: every code phase goes through the
council gate (`097`, ~30-min budget, `Council-Submitted:` when committing ahead
of the verdict); the P1 submission names the `check_render_mode` routing arm as
its one behaviour-visible edit; the seam is **registered in the same commit that
ships it** (2026-07-28 condition 2 — the whole surviving requirement) with its
landmines (§6.1, §6.6) and open questions written down; and if the architecture
seat objects on scope, that veto is answered by routing, not by resubmission
with better numbers (the 124 lesson). Consumers are **told, not merely
measured** (2026-07-29 §3): the component-library lane, the 283
instance-scope lane, the 238/268 carry lane, the inline-imagery lane, and the
render-path checkers' owners — the message is what changes about their
guarantee ("a `page_components` row may now be a child; page-grain sweeps that
assume row = section are wrong on composed pages"), not a list of new keys.

## 8. Relationship to the inline-imagery plan — interface, not fork

`inline_guide_imagery/PLAN_2026-08-14…` is another lane's design and stays so.
The division:

- **Their plan-as-truth mechanism stands for guides now.** A locked
  `site_plan_imagery` row per figure + the live IMG-056 resolver is the right
  durability answer for `article-body` pages **today**, and nothing here blocks
  or supersedes it on their timeline.
- **Composition consumes the same substrate.** A figure child's image fields
  resolve through the same resolver from the same plan rows — the plan row stays
  the durable declaration ("this page has this figure"), and the child instance
  becomes its durable **placement** ("here, in this order, in this design").
  Their §2 options table gains its option (e) — this document is that option,
  costed.
- **The handover moment is P5**, jointly scheduled, page by page, never a big
  bang: a guide migrates from blob+splice to parent+children, and their Phase-3
  splice becomes unnecessary on migrated pages.

> **OWNER STEER, 2026-08-31 — recorded, no phase re-scoped.** Relayed by the
> `agentchassis-ff` session from the dartsonline guides review:
>
> > *"please can you somehow hint that the guides need a lot more images (accurate
> > images) in between the paragraphs (small sections)… e.g. the grip styles copy…
> > we could have an image for each of those sections (e.g. ring grip, razor grip,
> > shark grip etc) and the same with the other guides."*
>
> **Sized by that lane `[MEASURED 2026-08-31]`:** `grip-styles` has **1** `<img>`
> total — the logo — against **7 h2 + 6 h3** headings; `board-setup`,
> `flight-shapes` and `beginners` each carry exactly one in-body illustration
> against 8–10 headings. Target is roughly one image per h3, i.e. 4–8 per guide
> against today's 0–1. Not from zero — two Banana illustrations exist from
> 2026-08-05 — but the owner's own example has thirteen headings and nothing.
>
> **ANSWERED: do NOT wait for composition, and P5 is the wrong thing to block on.**
> P1 is unfinished (three council REVISE rounds on the wiring; the extraction,
> the single-target guards and the routing migration still unwritten), nothing is
> live (**0 of 2,249** rows carry a `parent_instance_id`, **0 of 386** components
> declare a slots block, 2026-08-31), and P5 sits behind P2, P3 and P4 with §5's
> own gate — "only after P1–P3 have held on editorial pages for real weeks", with
> §6.1's un-owned-page question settled first, and the guides ARE those un-owned
> pages. Months, not weeks.
>
> **The trap is not composition's absence — it is that ONE llm-owned field owns
> both prose and figures.** That is what killed four figures on that lane in
> August, and it is what G1 removes. But per-section durability does not require
> composition: a guide whose h3 sections are separate FLAT `page_components` rows
> gets it today from live mechanisms, because a rewrite then targets one row and
> cannot take a sibling's figure with it. Composition's added value over that is
> nesting a figure INSIDE a prose section as a child — a refinement of placement,
> not the property that stops figures dying. **If the guides genuinely need
> figure-inside-prose rather than figure-between-sections, they are a P2 consumer
> rather than a P5 one, and that is a real input to this design** — it has been
> put back to that lane as the question worth answering.
>
> **Accuracy is already solved and should not be re-litigated.** Same lane, same
> day: the Banana-generated content heroes are correct (four distinguishable,
> correct grip patterns in one frame) while the July SDXL-era leftovers hallucinate
> feathered flights and numberless boards. So **the model is no longer the
> constraint — only placement and durability are.** Whatever mechanism places
> per-section figures should generate them the way the current content heroes are
> generated.
>
> `grip-styles` is offered as the canary when a mechanism exists that survives a
> rewrite — its six h3s are exactly the ring/razor/shark split the owner named.
> Wanted at **P2**, not now.
>
> ---
>
> **⚠⚠ CORRECTED AND NARROWED 2026-09-04 — THIS FEATURE'S DURABILITY ARGUMENT DOES
> NOT COVER STRUCTURE, ONLY ASSETS.** Found with the `infographics` lane; the
> evidence is theirs, the bound is mine, and it is recorded here because it is the
> claim P1 should make at council instead of the broader one.
>
> **What this document has been arguing:** anything placed inside `article-body`
> dies at the next wholesale rewrite, because one llm-owned field holds it —
> therefore composition is required to give it a durable home.
>
> **THE HALF THAT IS FALSE.** A graphic does not have to go *inside* the prose. A
> comparison-table added as a SIBLING SECTION (position 3, pushing the CTA to 4)
> touches no body, cannot be destroyed by a rewrite, and **needs nothing from this
> feature**. `finetuning.uk` ships exactly that today.
>
> **AND THE HALF THAT IS WORSE FOR US — a "ROUTE C" nobody had named.**
> `gamesdesign.co.uk` carries a table on **13 of 13** article bodies, against **3
> across the other 368 on the fleet** `[MEASURED 2026-09-04]`. The mechanism is not a
> component and not a plan row: its `content_direction` spec carries a structural
> writing rule (*"never describe a sequence of steps purely in prose when a table
> would make it scannable"*), so **the WRITER emits the structure**. That inverts our
> premise — the artefact is not durable, **the RULE is, and the next rewrite
> RE-DERIVES the table from the same spec.** Durability by regeneration, with nothing
> built.
>
> **THE BOUND, which is what keeps this feature justified:**
>
> > **If the thing is made of the page's own words, a rule re-derives it. If it is an
> > ASSET, something must own its persistence.**
>
> A `content_direction` rule can *ask* for an image; it cannot produce one, because a
> writer emits prose and not a JPEG. So:
>
> | inside prose | durable by | needs 035? |
> |---|---|---|
> | **structure** (table, checklist, flow) | the RULE re-deriving it | **NO — route C is better and exists** |
> | **an asset** (figure, chart image) | someone owning its persistence | **YES — nothing re-derives an asset** |
>
> **So P1's honest claim is the narrow one:** composition's justification is REMOVED
> for structure-inside-prose and SURVIVES for imagery-inside-prose. State that at
> council. **The broad version does not survive contact with the gamesdesign data,
> and a reviewer who found route C would take the whole justification down with it.**
>
> **ANSWERED 2026-08-31 by that lane, and it NARROWS P2's consumer set: the guides
> are NOT a P2 consumer.** Figure-BETWEEN-sections is sufficient for them; nesting
> is not needed. Their structure is `h2` → `h3 Ring Grip` (2 paragraphs) →
> `h3 Razor Grip` (2 paragraphs) → …, and each h3 wants exactly ONE image about
> that grip. So if each h3 is its own row with its own image field, the figure is a
> **field of the section** rather than something wedged into a shared blob — the
> durability property achieved by making sections FINER, not by nesting.
> **The case that WOULD make a consumer P2 is a figure between a section's own
> paragraphs** (prose, figure, more prose, inside one `h3`); they see none in
> eleven guides, where every candidate image is "here is what this thing looks
> like" and sits at the head or foot of its section.
> ⚠ **Scope the evidence as they did:** eleven guides, one site, judged from
> RENDERED structure. It is not evidence about the editorial corpus this document
> designs for, and must not be quoted as though it were — if the insight pages want
> figure-inside-prose, that is a separate finding and still a P2 case.

> **ADDENDUM 2026-09-02 — the ARTICLE corpus is one blob row per article, so §8's
> "make the sections finer" answer does NOT transfer to it, and 035 must not be cited
> as the fix for the defect the owner is currently looking at.**
> Measured at boxingonline (first paid site, second owner review): all six `/blog/`
> article pages are **exactly ONE `article-body` row** (plus sometimes a
> `call-to-action`) — there is no per-`h3` row to hang an image field on, unlike the
> eleven darts guides that answer was derived from. §8's own warning fires here: that
> evidence "is not evidence about the editorial corpus this document designs for".
> **But the visible defect is not a composition problem at all.** `article-body` has one
> field (`content`) and a template whose only interpolation is `{{.content}}` — no
> `<img>`, `<figure>` or `background-image` anywhere — so it cannot display an image by
> construction, while **six generated, deployed, HTTP-200 header images sit unreferenced**
> (one per article). The fix is component capability: an optional image field plus
> template markup, default absent, byte-identical for all 297 existing instances across
> 30 sites. **It needs nothing from P1 and must not wait for it** — 035 is inert (0 of
> 2,249 rows parented, 0 of 386 components declaring slots) and these are un-owned pages,
> i.e. behind §6.1 and P5.
> **CORRECTED 2026-09-02 (same day, before anything shipped) — "the fix is component
> capability" WAS WRONG, and the migration that implemented it was applied and rolled back.**
> `article-body` genuinely cannot display an image, and the six unreferenced heroes are real —
> but giving it its own image field sourced from `site_assets.hero` renders **the same image
> twice** on any page that also has a `hero` component, and `[MEASURED 2026-09-02]` that is
> **292 of the 301 pages carrying article-body, across 31 sites**. The six boxingonline pages
> that motivated it are in the **nine-page minority with no hero component** — I measured the
> motivating case and generalised it to the population. Migration 686 was applied ~13:56Z and
> rolled back ~15:05Z; **0 of 301 instances ever acquired the field, so nothing was rendered
> with it.**
> **The real defect is one level up, in page composition.** Peer pages show imagery through the
> `hero` component from a page-scope plan row (verified: `agritec.uk/blog/insect-bioconversion.html`
> renders `url('/assets/images/hero-bsf.jpg')`, 1 plan hero row, its only `<img>` the logo). The
> six boxingonline blog pages have neither — which is precisely the case `ContentHeroKey`
> generates per-article images FOR, and nothing renders them because there is no hero section to
> render them in. **So this was never evidence about component capability for 035's corpus**, and
> the paragraph above should not be cited as it stood.

> **What WOULD be a P2 case is unchanged and still unmet:** a figure between a section's
> own paragraphs. Once an article is a single blob, any in-body figure lives inside the
> llm-owned field and dies on the next rewrite — the G1 class. That is the
> `inline_guide_imagery` lane's durability problem today, and becomes a P2 consumer only
> if decomposed generation is what finally places those figures.
> Full measurement + the ownership split: `editorial_design_uplift/NOTES` 2026-09-02 tail.

## 9. What NOT to do

- Do not seed any `composite` row, or set any `parent_instance_id`, before the
  P1 code ships — §6.6 and the image-before-config rule.
- Do not add `{{template}}`/partials or any function to the executor for this.
- Do not let any agent apply a composition spec without human review (design D).
- Do not put children in `pages.sections` or `site_plan_sections` — composition
  is below the section grain, always.
- Do not adopt `child_components` jsonb as a second home for slot declarations.
- Do not answer §6.1 by adding CASCADE to the parent FK or a floor to
  `save_page_sections` inside this feature — that boundary gets its own review
  at P5.
- Do not build P2 generation before P1's read path is live and proven — a
  generation fan-out with no walk to render it is substrate-before-renderer
  inverted, and just as wrong.
