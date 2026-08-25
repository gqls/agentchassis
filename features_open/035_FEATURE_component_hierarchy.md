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
6. Both render paths get the walk: assembly (`assembleComponents`) and the
   sections rerender path. One walk implementation, two callers — not two walks.

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

   **The rule for P1:** capture the digest immediately after each node's
   `RenderTemplate` call, or give each node its own context. Do NOT read
   `ctx.RenderedTemplateSHA` after the walk and attribute it to anything.
   Falsifier for P0/P1: render a two-child parent where **one child's template
   is the empty string**, and assert that child's stored stamp is EMPTY — not
   its sibling's digest. A test that only renders non-empty children passes
   whether or not this is handled.

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
